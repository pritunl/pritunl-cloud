package qemu

import (
	"os"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/dhcpc"
	"github.com/pritunl/pritunl-cloud/dhcps"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/guest"
	"github.com/pritunl/pritunl-cloud/hugepages"
	"github.com/pritunl/pritunl-cloud/imds"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/permission"
	"github.com/pritunl/pritunl-cloud/qmp"
	"github.com/pritunl/pritunl-cloud/qms"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/store"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/tpm"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/virtiofs"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

func initDirs(virt *vm.VirtualMachine) (err error) {
	vmPath := paths.GetVmPath(virt.Id)

	err = utils.ExistsMkdir(settings.Hypervisor.LibPath, 0755)
	if err != nil {
		return
	}

	err = utils.ExistsMkdir(settings.Hypervisor.RunPath, 0755)
	if err != nil {
		return
	}

	err = utils.ExistsMkdir(paths.GetImdsPath(), 0755)
	if err != nil {
		return
	}

	err = utils.ExistsMkdir(vmPath, 0755)
	if err != nil {
		return
	}

	return
}

func initHugepage(virt *vm.VirtualMachine) (err error) {
	if !virt.Hugepages {
		return
	}

	err = hugepages.UpdateHugepagesSize()
	if err != nil {
		return
	}

	hugepagesPath := paths.GetHugepagePath(virt.Id)
	_ = os.Remove(hugepagesPath)

	return
}

func cleanRun(virt *vm.VirtualMachine) (err error) {
	_ = tpm.Stop(virt)
	_ = dhcpc.Stop(virt)
	_ = imds.Stop(virt)
	_ = dhcps.Stop(virt)
	_ = virtiofs.StopAll(virt)

	runPath := paths.GetInstRunPath(virt.Id)
	pidPath := paths.GetPidPath(virt.Id)
	sockPath := paths.GetSockPath(virt.Id)
	qmpSockPath := paths.GetQmpSockPath(virt.Id)
	guestPath := paths.GetGuestPath(virt.Id)

	err = utils.RemoveAll(runPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(pidPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(sockPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(qmpSockPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(guestPath)
	if err != nil {
		return
	}

	return
}

func initCache(virt *vm.VirtualMachine) (err error) {
	err = utils.ExistsMkdir(paths.GetCachesDir(), 0755)
	if err != nil {
		return
	}

	err = utils.ExistsMkdir(paths.GetCacheDir(virt.Id), 0700)
	if err != nil {
		return
	}

	return
}

func initRun(virt *vm.VirtualMachine) (err error) {
	runPath := paths.GetInstRunPath(virt.Id)

	err = utils.ExistsMkdir(runPath, 0700)
	if err != nil {
		return
	}

	return
}

func initPermissions(virt *vm.VirtualMachine) (err error) {
	err = permission.InitVirt(virt)
	if err != nil {
		return
	}

	err = permission.InitImds(virt)
	if err != nil {
		return
	}

	for _, mount := range virt.Mounts {
		shareId := paths.GetShareId(virt.Id, mount.Name)

		err = permission.InitMount(virt, shareId)
		if err != nil {
			return
		}
	}

	return
}

func writeOvmfVars(virt *vm.VirtualMachine) (err error) {
	if !virt.Uefi {
		return
	}

	ovmfVarsPath := paths.GetOvmfVarsPath(virt.Id)
	ovmfVarsPathSource, err := paths.FindOvmfVarsPath(virt.SecureBoot)
	if err != nil {
		return
	}

	err = utils.ExistsMkdir(paths.GetOvmfDir(), 0755)
	if err != nil {
		return
	}

	err = utils.Exec("", "cp", ovmfVarsPathSource, ovmfVarsPath)
	if err != nil {
		return
	}

	err = utils.Chmod(ovmfVarsPath, 0600)
	if err != nil {
		return
	}

	return
}

func activateDisks(db *database.Database,
	virt *vm.VirtualMachine) (err error) {

	for _, virtDsk := range virt.Disks {
		dsk, e := disk.Get(db, virtDsk.Id)
		if e != nil {
			err = e
			return
		}

		err = data.ActivateDisk(db, dsk)
		if err != nil {
			return
		}
	}

	return
}

func deactivateDisks(db *database.Database,
	virt *vm.VirtualMachine) (err error) {

	for _, virtDsk := range virt.Disks {
		dsk, e := disk.Get(db, virtDsk.Id)
		if e != nil {
			err = e
			return
		}

		err = data.DeactivateDisk(db, dsk)
		if err != nil {
			return
		}
	}

	return
}

func writeService(virt *vm.VirtualMachine) (err error) {
	unitPath := paths.GetUnitPath(virt.Id)

	qm, err := NewQemu(virt)
	if err != nil {
		return
	}

	output, err := qm.Marshal()
	if err != nil {
		return
	}

	err = utils.CreateWrite(unitPath, output, 0644)
	if err != nil {
		return
	}

	err = systemd.Reload()
	if err != nil {
		return
	}

	return
}

func Destroy(db *database.Database, virt *vm.VirtualMachine) (err error) {
	vmPath := paths.GetVmPath(virt.Id)
	unitName := paths.GetUnitName(virt.Id)
	unitPath := paths.GetUnitPath(virt.Id)
	unitPathServer4 := paths.GetUnitPathDhcp4(virt.Id, 0)
	unitPathServer6 := paths.GetUnitPathDhcp6(virt.Id, 0)
	unitPathServerNdp := paths.GetUnitPathNdp(virt.Id, 0)
	tpmPath := paths.GetTpmPath(virt.Id)
	runPath := paths.GetInstRunPath(virt.Id)
	unitPathTpm := paths.GetUnitPathTpm(virt.Id)
	unitPathImds := paths.GetUnitPathImds(virt.Id)
	unitPathDhcpc := paths.GetUnitPathDhcpc(virt.Id)
	unitPathShares := paths.GetUnitPathShares(virt.Id)
	sockPath := paths.GetSockPath(virt.Id)
	sockQmpPath := paths.GetQmpSockPath(virt.Id)
	// TODO Backward compatibility
	sockPathOld := paths.GetSockPath(virt.Id)
	guestPath := paths.GetGuestPath(virt.Id)
	// TODO Backward compatibility
	guestPathOld := paths.GetGuestPathOld(virt.Id)
	pidPath := paths.GetPidPath(virt.Id)
	// TODO Backward compatibility
	pidPathOld := paths.GetPidPathOld(virt.Id)
	ovmfVarsPath := paths.GetOvmfVarsPath(virt.Id)
	hugepagesPath := paths.GetHugepagePath(virt.Id)
	cachePath := paths.GetCacheDir(virt.Id)

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("qemu: Destroying virtual machine")

	exists, err := utils.Exists(unitPath)
	if err != nil {
		return
	}

	if exists {
		vrt, e := GetVmInfo(db, virt.Id, false, true)
		if e != nil {
			err = e
			return
		}

		if vrt != nil && vrt.State == vm.Running {
			guestShutdown := true
			err = guest.Shutdown(virt.Id)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"instance_id": virt.Id.Hex(),
					"error":       err,
				}).Warn("qemu: Failed to send shutdown to guest agent")
				err = nil
				guestShutdown = false
			}

			logged := false
			for i := 0; i < 10; i++ {
				err = qmp.Shutdown(virt.Id)
				if err == nil {
					break
				}

				if guestShutdown {
					err = nil
					break
				}

				if !logged {
					logged = true
					logrus.WithFields(logrus.Fields{
						"instance_id": virt.Id.Hex(),
						"error":       err,
					}).Warn(
						"qemu: Failed to send shutdown to virtual machine")
				}

				time.Sleep(500 * time.Millisecond)
			}

			shutdown := false
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"id":    virt.Id.Hex(),
					"error": err,
				}).Error("qemu: Power off virtual machine error")
				err = nil
			} else {
				for i := 0; i < settings.Hypervisor.StopTimeout; i++ {
					vrt, err = GetVmInfo(db, virt.Id, false, true)
					if err != nil {
						return
					}

					if vrt == nil || vrt.State == vm.Stopped ||
						vrt.State == vm.Failed {

						if vrt != nil {
							err = vrt.Commit(db)
							if err != nil {
								return
							}
						}

						shutdown = true
						break
					}

					time.Sleep(1 * time.Second)

					if (i+1)%15 == 0 {
						go func() {
							qmp.Shutdown(virt.Id)
							qms.Shutdown(virt.Id)
						}()
					}
				}
			}

			if !shutdown {
				logrus.WithFields(logrus.Fields{
					"id": virt.Id.Hex(),
				}).Warning("qemu: Force power off virtual machine")
			}
		}

		err = systemd.Stop(unitName)
		if err != nil {
			return
		}
	}

	time.Sleep(1 * time.Second)

	_ = tpm.Stop(virt)
	_ = dhcpc.Stop(virt)
	_ = imds.Stop(virt)
	_ = dhcps.Stop(virt)
	_ = virtiofs.StopAll(virt)

	err = NetworkConfClear(db, virt)
	if err != nil {
		return
	}

	time.Sleep(3 * time.Second)

	for _, dsk := range virt.Disks {
		ds, e := disk.Get(db, dsk.GetId())
		if e != nil {
			err = e
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
				continue
			}
			return
		}

		err = data.DeactivateDisk(db, ds)
		if err != nil {
			return
		}

		if ds.Index == "0" && ds.SourceInstance == virt.Id {
			err = disk.Delete(db, ds.Id)
			if err != nil {
				if _, ok := err.(*database.NotFoundError); ok {
					err = nil
					continue
				}
				return
			}
		} else {
			err = disk.Detach(db, dsk.GetId())
			if err != nil {
				return
			}
		}
	}

	for i, dsk := range virt.DriveDevices {
		if dsk.Type != vm.Lvm {
			continue
		}

		dskId, ok := utils.ParseObjectId(dsk.Id)
		if dskId.IsZero() || !ok {
			err = &errortypes.ParseError{
				errors.Newf("qemu: Failed to parse LVM disk ID '%s'", dsk.Id),
			}
			return
		}

		ds, e := disk.Get(db, dskId)
		if e != nil {
			err = e
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
				continue
			}
			return
		}

		if i == 0 && ds.SourceInstance == virt.Id {
			err = disk.Delete(db, ds.Id)
			if err != nil {
				if _, ok := err.(*database.NotFoundError); ok {
					err = nil
					continue
				}
				return
			}
		} else {
			err = disk.Detach(db, dskId)
			if err != nil {
				return
			}
		}
	}

	err = utils.RemoveAll(vmPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(unitPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(tpmPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(runPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(unitPathTpm)
	if err != nil {
		return
	}

	err = utils.RemoveAll(unitPathImds)
	if err != nil {
		return
	}

	err = utils.RemoveAll(unitPathDhcpc)
	if err != nil {
		return
	}

	err = utils.RemoveAll(unitPathServer4)
	if err != nil {
		return
	}

	err = utils.RemoveAll(unitPathServer6)
	if err != nil {
		return
	}

	err = utils.RemoveAll(unitPathServerNdp)
	if err != nil {
		return
	}

	_, err = utils.RemoveWildcard(unitPathShares)
	if err != nil {
		return
	}

	err = utils.RemoveAll(sockPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(sockQmpPath)
	if err != nil {
		return
	}

	// TODO Backward compatibility
	err = utils.RemoveAll(sockPathOld)
	if err != nil {
		return
	}

	err = utils.RemoveAll(guestPath)
	if err != nil {
		return
	}

	// TODO Backward compatibility
	err = utils.RemoveAll(guestPathOld)
	if err != nil {
		return
	}

	err = utils.RemoveAll(pidPath)
	if err != nil {
		return
	}

	// TODO Backward compatibility
	err = utils.RemoveAll(pidPathOld)
	if err != nil {
		return
	}

	err = utils.RemoveAll(paths.GetInitPath(virt.Id))
	if err != nil {
		return
	}

	err = utils.RemoveAll(unitPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(ovmfVarsPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(hugepagesPath)
	if err != nil {
		return
	}

	err = utils.RemoveAll(cachePath)
	if err != nil {
		return
	}

	err = permission.UserDelete(virt)
	if err != nil {
		return
	}

	store.RemVirt(virt.Id)
	store.RemDisks(virt.Id)
	store.RemAddress(virt.Id)
	store.RemRoutes(virt.Id)
	store.RemArp(virt.Id)

	return
}

func Cleanup(db *database.Database, virt *vm.VirtualMachine) {
	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("qemu: Stopped virtual machine")

	_ = tpm.Stop(virt)
	_ = dhcpc.Stop(virt)
	_ = imds.Stop(virt)
	_ = dhcps.Stop(virt)
	_ = virtiofs.StopAll(virt)

	hugepagesPath := paths.GetHugepagePath(virt.Id)
	_ = os.Remove(hugepagesPath)

	err := NetworkConfClear(db, virt)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id":    virt.Id.Hex(),
			"error": err,
		}).Error("qemu: Failed to cleanup virtual machine network")
	}

	time.Sleep(3 * time.Second)

	err = deactivateDisks(db, virt)
	if err != nil {
		return
	}

	store.RemVirt(virt.Id)
	store.RemDisks(virt.Id)

	return
}
