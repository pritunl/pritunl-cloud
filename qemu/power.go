package qemu

import (
	"os"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/cloudinit"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/dhcpc"
	"github.com/pritunl/pritunl-cloud/dhcps"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/guest"
	"github.com/pritunl/pritunl-cloud/imds"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/qmp"
	"github.com/pritunl/pritunl-cloud/qms"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/store"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/tpm"
	"github.com/pritunl/pritunl-cloud/virtiofs"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
	"github.com/sirupsen/logrus"
)

func PowerOn(db *database.Database, inst *instance.Instance,
	virt *vm.VirtualMachine) (err error) {
	unitName := paths.GetUnitName(virt.Id)

	if constants.Interrupt {
		return
	}

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("qemu: Starting virtual machine")

	err = inst.InitUnixId(db)
	if err != nil {
		return
	}
	virt.UnixId = inst.UnixId

	if inst.Vnc {
		err = inst.InitVncDisplay(db)
		if err != nil {
			return
		}
		virt.VncDisplay = inst.VncDisplay
	}

	if inst.Spice {
		err = inst.InitSpicePort(db)
		if err != nil {
			return
		}
		virt.SpicePort = inst.SpicePort
	}

	err = initDirs(virt)
	if err != nil {
		return
	}

	err = cleanRun(virt)
	if err != nil {
		return
	}

	if len(virt.NetworkAdapters) == 0 {
		err = &errortypes.NotFoundError{
			errors.Wrap(err, "cloudinit: Instance missing network adapters"),
		}
		return
	}

	adapter := virt.NetworkAdapters[0]

	if adapter.Vpc.IsZero() {
		err = &errortypes.NotFoundError{
			errors.Wrap(err, "cloudinit: Instance missing VPC"),
		}
		return
	}

	if adapter.Subnet.IsZero() {
		err = &errortypes.NotFoundError{
			errors.Wrap(err, "cloudinit: Instance missing VPC subnet"),
		}
		return
	}

	dc, err := datacenter.Get(db, node.Self.Datacenter)
	if err != nil {
		return
	}

	zne, err := zone.Get(db, node.Self.Zone)
	if err != nil {
		return
	}

	vc, err := vpc.Get(db, adapter.Vpc)
	if err != nil {
		return
	}

	err = virt.GenerateImdsSecret()
	if err != nil {
		return
	}

	err = cloudinit.Write(db, inst, virt, dc, zne, vc, false)
	if err != nil {
		return
	}

	err = initCache(virt)
	if err != nil {
		return
	}

	err = initHugepage(virt)
	if err != nil {
		return
	}

	err = writeOvmfVars(virt)
	if err != nil {
		return
	}

	err = activateDisks(db, virt)
	if err != nil {
		return
	}

	err = writeService(virt)
	if err != nil {
		return
	}

	err = initRun(virt)
	if err != nil {
		return
	}

	err = virtiofs.StartAll(db, virt)
	if err != nil {
		return
	}

	err = initPermissions(virt)
	if err != nil {
		return
	}

	if virt.DhcpServer {
		err = dhcps.Start(db, virt, dc, zne, vc)
		if err != nil {
			return
		}
	} else {
		err = dhcps.Stop(virt)
		if err != nil {
			return
		}
	}

	if virt.Tpm {
		err = tpm.Start(db, virt)
		if err != nil {
			return
		}
	} else {
		err = tpm.Stop(virt)
		if err != nil {
			return
		}
	}

	err = systemd.Start(unitName)
	if err != nil {
		return
	}

	err = Wait(db, virt)
	if err != nil {
		return
	}

	if virt.Vnc {
		err = qmp.VncPassword(virt.Id, inst.VncPassword)
		if err != nil {
			return
		}
	}

	if virt.Spice {
		err = qmp.SetPassword(virt.Id, qmp.Spice, inst.SpicePassword)
		if err != nil {
			return
		}
	}

	err = NetworkConf(db, virt)
	if err != nil {
		return
	}

	store.RemVirt(virt.Id)
	store.RemDisks(virt.Id)

	return
}

func PowerOff(db *database.Database, virt *vm.VirtualMachine) (err error) {
	unitName := paths.GetUnitName(virt.Id)

	if constants.Interrupt {
		return
	}

	logrus.WithFields(logrus.Fields{
		"id": virt.Id.Hex(),
	}).Info("qemu: Stopping virtual machine")

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
			}).Warn("qemu: Failed to send shutdown to virtual machine")
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
			vrt, e := GetVmInfo(db, virt.Id, false, true)
			if e != nil {
				err = e
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
			"instance_id": virt.Id.Hex(),
		}).Warning("qemu: Force power off virtual machine")

		err = systemd.Stop(unitName)
		if err != nil {
			return
		}
	}

	_ = tpm.Stop(virt)
	_ = dhcpc.Stop(virt)
	_ = imds.Stop(virt)
	_ = dhcps.Stop(virt)
	_ = virtiofs.StopAll(virt)

	hugepagesPath := paths.GetHugepagePath(virt.Id)
	_ = os.Remove(hugepagesPath)

	err = NetworkConfClear(db, virt)
	if err != nil {
		return
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

func ForcePowerOffErr(db *database.Database, virt *vm.VirtualMachine,
	e error) (err error) {

	unitName := paths.GetUnitName(virt.Id)

	if constants.Interrupt {
		return
	}

	logrus.WithFields(logrus.Fields{
		"instance_id": virt.Id.Hex(),
		"error":       e,
	}).Error("qemu: Force power off virtual machine")

	go guest.Shutdown(virt.Id)
	go qmp.Shutdown(virt.Id)
	go qms.Shutdown(virt.Id)

	time.Sleep(15 * time.Second)

	err = systemd.Stop(unitName)
	if err != nil {
		return
	}

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

	err = deactivateDisks(db, virt)
	if err != nil {
		return
	}

	store.RemVirt(virt.Id)
	store.RemDisks(virt.Id)

	return
}

func ForcePowerOff(db *database.Database, virt *vm.VirtualMachine) (
	err error) {

	unitName := paths.GetUnitName(virt.Id)

	if constants.Interrupt {
		return
	}

	logrus.WithFields(logrus.Fields{
		"instance_id": virt.Id.Hex(),
	}).Warning("qemu: Force power off virtual machine")

	go guest.Shutdown(virt.Id)
	go qmp.Shutdown(virt.Id)
	go qms.Shutdown(virt.Id)

	time.Sleep(5 * time.Second)

	err = systemd.Stop(unitName)
	if err != nil {
		return
	}

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

	err = deactivateDisks(db, virt)
	if err != nil {
		return
	}

	store.RemVirt(virt.Id)
	store.RemDisks(virt.Id)

	return
}
