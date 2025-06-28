package deploy

import (
	"strconv"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/imds"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/nodeport"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

var (
	deploymentsLock = utils.NewMultiTimeoutLock(5 * time.Minute)
)

type Deployments struct {
	stat *state.State
}

func (d *Deployments) migrate(deply *deployment.Deployment) {
	nde := d.stat.Node()
	nodeId := d.stat.Node().Id

	acquired, lockId := deploymentsLock.LockOpenTimeout(
		deply.Id.Hex(), 3*time.Minute)
	if !acquired {
		return
	}

	go func() {
		defer func() {
			deploymentsLock.Unlock(deply.Id.Hex(), lockId)
		}()

		db := database.GetDatabase()
		defer db.Close()

		if deply.Node != nodeId {
			return
		}

		inst, err := instance.Get(db, deply.Instance)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"deployment_id": deply.Id.Hex(),
				"new_spec_id":   deply.NewSpec.Hex(),
				"error":         err,
			}).Error("deploy: Failed to get instance")
			return
		}

		inst.PreCommit()

		curSpec, err := spec.Get(db, deply.Spec)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"deployment_id": deply.Id.Hex(),
				"cur_spec_id":   curSpec.Id.Hex(),
				"new_spec_id":   deply.NewSpec.Hex(),
				"error":         err,
			}).Error("deploy: Failed to get current spec")
			return
		}

		newSpec, err := spec.Get(db, deply.NewSpec)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"deployment_id": deply.Id.Hex(),
				"cur_spec_id":   curSpec.Id.Hex(),
				"new_spec_id":   newSpec.Id.Hex(),
				"error":         err,
			}).Error("deploy: Failed to get new spec")
			return
		}

		errData, err := curSpec.CanMigrate(db, newSpec)
		if err != nil || errData != nil {
			logrus.WithFields(logrus.Fields{
				"deployment_id": deply.Id.Hex(),
				"cur_spec_id":   curSpec.Id.Hex(),
				"new_spec_id":   newSpec.Id.Hex(),
				"error":         err,
				"error_data":    errData,
			}).Error("deploy: Incompatible migrate")

			deply.State = deployment.Deployed
			deply.NewSpec = primitive.NilObjectID
			err = deply.CommitFields(db, set.NewSet("state", "new_spec"))
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"deployment_id": deply.Id.Hex(),
					"cur_spec_id":   curSpec.Id.Hex(),
					"new_spec_id":   newSpec.Id.Hex(),
					"error":         err,
				}).Error("deploy: Failed to commit deployment")
				return
			}

			return
		}

		instFields := set.NewSet()

		if curSpec.Instance.Uefi != newSpec.Instance.Uefi {
			if newSpec.Instance.Uefi != nil {
				instFields.Add("uefi")
				inst.Uefi = *newSpec.Instance.Uefi
			}
		}
		if curSpec.Instance.SecureBoot != newSpec.Instance.SecureBoot {
			if newSpec.Instance.SecureBoot != nil {
				instFields.Add("secure_boot")
				inst.SecureBoot = *newSpec.Instance.SecureBoot
			}
		}
		if curSpec.Instance.CloudType != newSpec.Instance.CloudType {
			if newSpec.Instance.CloudType != "" {
				instFields.Add("cloud_type")
				inst.CloudType = newSpec.Instance.CloudType
			}
		}
		if curSpec.Instance.Tpm != newSpec.Instance.Tpm {
			instFields.Add("tpm")
			inst.Tpm = newSpec.Instance.Tpm
		}
		if curSpec.Instance.Vnc != newSpec.Instance.Vnc {
			instFields.Add("vnc")
			inst.Vnc = newSpec.Instance.Vnc
		}
		if curSpec.Instance.DeleteProtection !=
			newSpec.Instance.DeleteProtection {

			instFields.Add("delete_protection")
			inst.DeleteProtection = newSpec.Instance.DeleteProtection
		}
		if curSpec.Instance.SkipSourceDestCheck !=
			newSpec.Instance.SkipSourceDestCheck {

			instFields.Add("skip_source_dest_check")
			inst.SkipSourceDestCheck = newSpec.Instance.SkipSourceDestCheck
		}
		if curSpec.Instance.Gui != newSpec.Instance.Gui {
			instFields.Add("gui")
			inst.Gui = newSpec.Instance.Gui
		}
		if curSpec.Instance.HostAddress != newSpec.Instance.HostAddress {
			if newSpec.Instance.HostAddress != nil {
				instFields.Add("no_host_address")
				inst.NoHostAddress = !*newSpec.Instance.HostAddress
			}
		}
		if curSpec.Instance.PublicAddress != newSpec.Instance.PublicAddress {
			if newSpec.Instance.PublicAddress != nil {
				instFields.Add("no_public_address")
				inst.NoPublicAddress = !*newSpec.Instance.PublicAddress
			} else {
				instFields.Add("no_public_address")
				inst.NoPublicAddress = nde.DefaultNoPublicAddress
			}
		}
		if curSpec.Instance.PublicAddress6 != newSpec.Instance.PublicAddress6 {
			if newSpec.Instance.PublicAddress6 != nil {
				instFields.Add("no_public_address6")
				inst.NoPublicAddress6 = !*newSpec.Instance.PublicAddress6
			} else {
				instFields.Add("no_public_address6")
				inst.NoPublicAddress6 = nde.DefaultNoPublicAddress6
			}
		}
		if curSpec.Instance.DhcpServer != newSpec.Instance.DhcpServer {
			instFields.Add("dhcp_server")
			inst.DhcpServer = newSpec.Instance.DhcpServer
		}

		if curSpec.Instance.Processors != newSpec.Instance.Processors ||
			curSpec.Instance.Memory != newSpec.Instance.Memory {

			flexible := true
			if !newSpec.Instance.Shape.IsZero() {
				shp, e := shape.Get(db, newSpec.Instance.Shape)
				if e != nil {
					err = e

					logrus.WithFields(logrus.Fields{
						"deployment_id": deply.Id.Hex(),
						"cur_spec_id":   curSpec.Id.Hex(),
						"new_spec_id":   newSpec.Id.Hex(),
						"error":         err,
					}).Error("deploy: Failed to get spec shape")
					return
				}

				flexible = shp.Flexible
			}

			if flexible && inst != nil {
				inst.Processors = newSpec.Instance.Processors
				instFields.Add("processors")
				inst.Memory = newSpec.Instance.Memory
				instFields.Add("memory")
			}
		}

		if !utils.CompareStringSlicesUnsorted(curSpec.Instance.Roles,
			newSpec.Instance.Roles) {

			instFields.Add("network_roles")
			inst.NetworkRoles = newSpec.Instance.Roles
		}

		if curSpec.Instance.DiffNodePorts(newSpec.Instance.NodePorts) {
			instFields.Add("node_ports")

			newNodePorts := []*nodeport.Mapping{}
			for _, ndePort := range newSpec.Instance.NodePorts {
				newNodePorts = append(newNodePorts, &nodeport.Mapping{
					Protocol:     ndePort.Protocol,
					ExternalPort: ndePort.ExternalPort,
					InternalPort: ndePort.InternalPort,
				})
			}

			inst.UpsertNodePorts(newNodePorts)
		}

		inst.Mounts = []*instance.Mount{}
		for _, mount := range newSpec.Instance.Mounts {
			if mount.Type != spec.HostPath {
				continue
			}

			inst.Mounts = append(inst.Mounts, &instance.Mount{
				Name:     mount.Name,
				Type:     instance.HostPath,
				Path:     mount.Path,
				HostPath: mount.HostPath,
			})
		}
		instFields.Add("mounts")

		if instFields.Len() > 0 {
			errData, err = inst.Validate(db)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"deployment_id": deply.Id.Hex(),
					"cur_spec_id":   curSpec.Id.Hex(),
					"new_spec_id":   newSpec.Id.Hex(),
					"error":         err,
					"error_data":    errData,
				}).Error("deploy: Migrate failed, invalid instance options")
				return
			}

			dskChange, err := inst.PostCommit(db)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"deployment_id": deply.Id.Hex(),
					"cur_spec_id":   curSpec.Id.Hex(),
					"new_spec_id":   newSpec.Id.Hex(),
					"error":         err,
					"error_data":    errData,
				}).Error("deploy: Migrate failed, instance post commit error")
				return
			}

			err = inst.CommitFields(db, instFields)
			if err != nil {
				_ = inst.Cleanup(db)

				logrus.WithFields(logrus.Fields{
					"deployment_id": deply.Id.Hex(),
					"cur_spec_id":   curSpec.Id.Hex(),
					"new_spec_id":   newSpec.Id.Hex(),
					"error":         err,
				}).Error("deploy: Failed to migrate instance")
				return
			}

			err = inst.Cleanup(db)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"deployment_id": deply.Id.Hex(),
					"cur_spec_id":   curSpec.Id.Hex(),
					"new_spec_id":   newSpec.Id.Hex(),
					"error":         err,
				}).Error("deploy: Failed to cleanup instance")
				err = nil
			}

			event.PublishDispatch(db, "instance.change")
			if dskChange {
				event.PublishDispatch(db, "disk.change")
			}
		}

		deply.Action = ""
		deply.Spec = newSpec.Id
		deply.NewSpec = primitive.NilObjectID
		err = deply.CommitFields(db, set.NewSet("action", "spec", "new_spec"))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"deployment_id": deply.Id.Hex(),
				"cur_spec_id":   curSpec.Id.Hex(),
				"new_spec_id":   newSpec.Id.Hex(),
				"error":         err,
			}).Error("deploy: Failed to commit deployment")
			return
		}

		logrus.WithFields(logrus.Fields{
			"deployment_id": deply.Id.Hex(),
			"cur_spec_id":   curSpec.Id.Hex(),
			"new_spec_id":   newSpec.Id.Hex(),
		}).Info("deploy: Migrated deployment")

		return
	}()
}

func (d *Deployments) destroy(deply *deployment.Deployment) {
	acquired, lockId := deploymentsLock.LockOpenTimeout(
		deply.Id.Hex(), 3*time.Minute)
	if !acquired {
		return
	}

	go func() {
		defer func() {
			deploymentsLock.Unlock(deply.Id.Hex(), lockId)
		}()

		db := database.GetDatabase()
		defer db.Close()

		if deply.Node != d.stat.Node().Id {
			return
		}

		if deply.Kind == deployment.Image && !deply.Image.IsZero() {
			img, err := image.Get(db, deply.Image)
			if err != nil {
				if _, ok := err.(*database.NotFoundError); ok {
					img = nil
					err = nil
				} else {
					logrus.WithFields(logrus.Fields{
						"deployment_id": deply.Id.Hex(),
						"error":         err,
					}).Error("deploy: Failed to get image")
					return
				}
			}

			if img != nil {
				err = data.DeleteImage(db, img.Id)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"deployment_id": deply.Id.Hex(),
						"error":         err,
					}).Error("deploy: Failed to remove deployment image")
					return
				}

				event.PublishDispatch(db, "image.change")
			}
		}

		inst := d.stat.GetInstace(deply.Instance)
		if inst != nil {
			if inst.DeleteProtection {
				logrus.WithFields(logrus.Fields{
					"deployment": deply.Id.Hex(),
					"instance":   inst.Id.Hex(),
				}).Warning("deploy: Cannot destroy deployment with " +
					"instance delete protection, archiving...")

				deply.Action = deployment.Archive
				err := deply.CommitFields(db, set.NewSet("action"))
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"deployment_id": deply.Id.Hex(),
						"error":         err,
					}).Error("deploy: Failed to commit deployment")
					return
				}
				return
			}

			if inst.Action != instance.Destroy {
				logrus.WithFields(logrus.Fields{
					"deployment": deply.Id.Hex(),
					"instance":   inst.Id.Hex(),
				}).Info("deploy: Delete deployment instance")

				err := instance.Delete(db, inst.Id)
				if err != nil {
					if _, ok := err.(*database.NotFoundError); !ok {
						err = nil
					} else {
						logrus.WithFields(logrus.Fields{
							"deployment_id": deply.Id.Hex(),
							"error":         err,
						}).Error("deploy: Failed to delete instance")
						return
					}
				}
			}
		} else {
			err := deployment.Remove(db, deply.Id)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"deployment_id": deply.Id.Hex(),
					"error":         err,
				}).Error("deploy: Failed to remove deployment")
				return
			}

			event.PublishDispatch(db, "pod.change")
		}

		return
	}()
}

func (d *Deployments) archive(deply *deployment.Deployment) (err error) {
	inst := d.stat.GetInstace(deply.Instance)
	nodeId := d.stat.Node().Id

	acquired, lockId := deploymentsLock.LockOpenTimeout(
		deply.Id.Hex(), 3*time.Minute)
	if !acquired {
		return
	}

	go func() {
		defer func() {
			deploymentsLock.Unlock(deply.Id.Hex(), lockId)
		}()

		db := database.GetDatabase()
		defer db.Close()

		if deply.Node != nodeId {
			return
		}

		if !inst.IsActive() {
			deply.State = deployment.Archived
			deply.Action = ""
			err = deply.CommitFields(db, set.NewSet("state", "action"))
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"deployment_id": deply.Id.Hex(),
					"error":         err,
				}).Error("deploy: Failed to commit deployment")
				return
			}

			return
		}

		if inst.Action != instance.Stop {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
			}).Info("deploy: Stopping instance for deployment archive")

			err = instance.SetAction(db, inst.Id, instance.Stop)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
					"error":       err,
				}).Error("deploy: Failed to set instance state")

				return
			}
		}
	}()

	return
}

func (d *Deployments) restore(deply *deployment.Deployment) (err error) {
	inst := d.stat.GetInstace(deply.Instance)
	instDisks := d.stat.GetInstaceDisks(deply.Instance)
	spc := d.stat.Spec(deply.Spec)
	nodeId := d.stat.Node().Id

	acquired, lockId := deploymentsLock.LockOpenTimeout(
		deply.Id.Hex(), 3*time.Minute)
	if !acquired {
		return
	}

	go func() {
		defer func() {
			deploymentsLock.Unlock(deply.Id.Hex(), lockId)
		}()

		db := database.GetDatabase()
		defer db.Close()

		if deply.Node != nodeId {
			return
		}

		if inst.IsActive() {
			deply.State = deployment.Deployed
			deply.Action = ""
			err = deply.CommitFields(db, set.NewSet("state", "action"))
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"deployment_id": deply.Id.Hex(),
					"error":         err,
				}).Error("deploy: Failed to commit deployment")
				return
			}

			return
		}

		index := 0
		curDisks := set.NewSet()
		for _, dsk := range instDisks {
			dskIndex, _ := strconv.Atoi(dsk.Index)
			index = max(index, dskIndex)
			curDisks.Add(dsk.Id)
		}

		if inst.Action != instance.Start {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
			}).Info("deploy: Starting instance for deployment restore")

			reservedDisks := []*disk.Disk{}
			deplyMounts := []*deployment.Mount{}

			for _, mount := range spc.Instance.Mounts {
				if mount.Type != spec.Disk {
					continue
				}

				index += 1
				diskReserved := false

				for _, dskId := range mount.Disks {
					if curDisks.Contains(dskId) {
						diskReserved = true
						break
					}
				}

				if !diskReserved {
					for _, dskId := range mount.Disks {
						dsk, e := disk.Get(db, dskId)
						if e != nil {
							err = e

							for _, dsk := range reservedDisks {
								err = dsk.Unreserve(db, inst.Id, deply.Id)
								if err != nil {
									return
								}
							}

							return
						}

						if dsk.Node != node.Self.Id || !dsk.Instance.IsZero() {
							continue
						}

						diskReserved, err = dsk.Reserve(
							db, inst.Id, index, deply.Id)
						if err != nil {
							for _, dsk := range reservedDisks {
								err = dsk.Unreserve(db, inst.Id, deply.Id)
								if err != nil {
									return
								}
							}
							return
						}

						if !diskReserved {
							continue
						}

						deplyMounts = append(deplyMounts, &deployment.Mount{
							Disk: dsk.Id,
							Path: mount.Path,
							Uuid: dsk.Uuid,
						})

						reservedDisks = append(reservedDisks, dsk)
						break
					}
				}

				if !diskReserved {
					for _, dsk := range reservedDisks {
						err = dsk.Unreserve(db, inst.Id, deply.Id)
						if err != nil {
							return
						}
					}

					logrus.WithFields(logrus.Fields{
						"mount_path": mount.Path,
					}).Error("deploy: Failed to reserve disk for mount")

					deply.State = deployment.Archived
					deply.Action = ""
					err = deply.CommitFields(db, set.NewSet("state", "action"))
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"deployment_id": deply.Id.Hex(),
							"error":         err,
						}).Error("deploy: Failed to commit deployment")
						return
					}

					return
				}
			}

			err = instance.SetAction(db, inst.Id, instance.Start)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
					"error":       err,
				}).Error("deploy: Failed to set instance state")

				return
			}
		}
	}()

	return
}

func (d *Deployments) imageShutdown(db *database.Database,
	deply *deployment.Deployment, virt *vm.VirtualMachine) {

	logrus.WithFields(logrus.Fields{
		"instance_id": virt.Id.Hex(),
	}).Info("deploy: Stopping instance for deployment image")

	time.Sleep(3 * time.Second)

	err := imds.Pull(db, virt.Id, virt.Deployment, virt.ImdsHostSecret)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"instance_id": virt.Id.Hex(),
			"error":       err,
		}).Error("deploy: Failed to pull imds state for shutdown")
	}

	err = instance.SetAction(db, virt.Id, instance.Stop)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"instance_id": virt.Id.Hex(),
			"error":       err,
		}).Error("deploy: Failed to set instance state")

		return
	}

	if deply.GetImageState() == "" {
		deply.SetImageState(deployment.Ready)
		err = deply.CommitFields(db, set.NewSet("image_data.state"))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": virt.Id.Hex(),
				"error":       err,
			}).Error("deploy: Failed to commit deployment state")
			return
		}
	}

	event.PublishDispatch(db, "pod.change")
}

func (d *Deployments) image(deply *deployment.Deployment) (err error) {
	acquired, lockId := deploymentsLock.LockOpenTimeout(
		deply.Id.Hex(), 5*time.Minute)
	if !acquired {
		return
	}

	go func() {
		defer func() {
			time.Sleep(3 * time.Second)
			deploymentsLock.Unlock(deply.Id.Hex(), lockId)
		}()

		db := database.GetDatabase()
		defer db.Close()

		if deply.Node != d.stat.Node().Id {
			return
		}

		inst := d.stat.GetInstace(deply.Instance)
		if inst == nil {
			return
		}

		virt := d.stat.GetVirt(inst.Id)

		if inst.Guest == nil {
			return
		}

		if inst.IsActive() && inst.Guest.Status == types.Imaged &&
			inst.Action != instance.Stop {

			d.imageShutdown(db, deply, virt)
			return
		}

		if deply.State == deployment.Deployed &&
			deply.GetImageState() == deployment.Ready &&
			!inst.IsActive() && inst.Guest.Status == types.Imaged {

			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
			}).Info("deploy: Creating deployment image")

			dsk, e := disk.GetInstanceIndex(db, inst.Id, "0")
			if e != nil {
				if _, ok := e.(*database.NotFoundError); ok {
					logrus.WithFields(logrus.Fields{
						"instance_id": inst.Id.Hex(),
						"error":       err,
					}).Error("deploy: Failed to find instance disk for image")

					deply.SetImageState(deployment.Failed)
					err = deply.CommitFields(db, set.NewSet("image_data.state"))
					if err != nil {
						return
					}

					dsk = nil
					err = nil
				} else {
					return
				}
			}

			dsk.Action = disk.Snapshot
			err = dsk.CommitFields(db, set.NewSet("action"))
			if err != nil {
				return
			}

			deply.SetImageState(deployment.Snapshot)
			err = deply.CommitFields(db, set.NewSet("image_data.state"))
			if err != nil {
				return
			}
		}
	}()

	return
}

func (d *Deployments) domainCommit(deply *deployment.Deployment,
	domn *domain.Domain, newRecs []*domain.Record) {

	acquired, lockId := deploymentsLock.LockOpenTimeout(
		deply.Id.Hex(), 3*time.Minute)
	if !acquired {
		return
	}

	go func() {
		defer func() {
			deploymentsLock.Unlock(deply.Id.Hex(), lockId)
		}()

		db := database.GetDatabase()
		defer db.Close()

		time.Sleep(500 * time.Millisecond)

		deply, err := deployment.Get(db, deply.Id)
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			}
			return
		}

		if deply.State != deployment.Deployed {
			return
		}

		logrus.WithFields(logrus.Fields{
			"domain_id": domn.Id.Hex(),
		}).Info("deploy: Committing domain records")

		recs := []*deployment.RecordData{}
		for _, rec := range newRecs {
			recs = append(recs, &deployment.RecordData{
				Domain: rec.SubDomain + "." + domn.RootDomain,
				Value:  rec.Value,
			})
		}
		domnData := &deployment.DomainData{
			Records: recs,
		}

		err = domn.CommitRecords(db)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"domain_id": domn.Id.Hex(),
				"error":     err,
			}).Error("deploy: Failed to commit domain records")
			return
		}

		deply.DomainData = domnData
		err = deply.CommitFields(db, set.NewSet("domain_data"))
		if err != nil {
			return
		}

		event.PublishDispatch(db, "domain.change")
		event.PublishDispatch(db, "pod.change")

		time.Sleep(500 * time.Millisecond)

		newDeply, err := deployment.Get(db, deply.Id)
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				newDeply = nil
				err = nil
			} else {
				return
			}
		}

		if newDeply == nil || newDeply.State != deployment.Deployed {
			logrus.WithFields(logrus.Fields{
				"domain_id": domn.Id.Hex(),
			}).Info("deploy: Undo domains commit for deactivated deployment")

			err = deployment.RemoveDomains(db, deply.Id)
			if err != nil {
				return
			}
		}
	}()
}

func (d *Deployments) domain(db *database.Database,
	deply *deployment.Deployment, spc *spec.Spec) (err error) {

	if spc.Domain != nil && deply.InstanceData != nil {
		newRecs := map[primitive.ObjectID][]*domain.Record{}

		for _, specRec := range spc.Domain.Records {
			domn := d.stat.SpecDomain(specRec.Domain)
			if domn == nil {
				continue
			}

			switch specRec.Type {
			case spec.Public:
				for _, val := range deply.InstanceData.PublicIps {
					rec := &domain.Record{
						Domain:     specRec.Domain,
						SubDomain:  specRec.Name,
						Deployment: deply.Id,
						Type:       domain.A,
						Value:      val,
					}

					errData, e := rec.Validate(db)
					if e != nil {
						err = e
						return
					}
					if errData != nil {
						err = errData.GetError()
						return
					}

					newRecs[domn.Id] = append(newRecs[domn.Id], rec)
				}
				break
			case spec.Public6:
				for _, val := range deply.InstanceData.PublicIps6 {
					rec := &domain.Record{
						Domain:     specRec.Domain,
						SubDomain:  specRec.Name,
						Deployment: deply.Id,
						Type:       domain.AAAA,
						Value:      val,
					}

					errData, e := rec.Validate(db)
					if e != nil {
						err = e
						return
					}
					if errData != nil {
						err = errData.GetError()
						return
					}

					newRecs[domn.Id] = append(newRecs[domn.Id], rec)
				}
				break
			case spec.Host:
				for _, val := range deply.InstanceData.HostIps {
					rec := &domain.Record{
						Domain:     specRec.Domain,
						SubDomain:  specRec.Name,
						Deployment: deply.Id,
						Type:       domain.A,
						Value:      val,
					}

					errData, e := rec.Validate(db)
					if e != nil {
						err = e
						return
					}
					if errData != nil {
						err = errData.GetError()
						return
					}

					newRecs[domn.Id] = append(newRecs[domn.Id], rec)
				}
				break
			case spec.Private:
				for _, val := range deply.InstanceData.PrivateIps {
					rec := &domain.Record{
						Domain:     specRec.Domain,
						SubDomain:  specRec.Name,
						Deployment: deply.Id,
						Type:       domain.A,
						Value:      val,
					}

					errData, e := rec.Validate(db)
					if e != nil {
						err = e
						return
					}
					if errData != nil {
						err = errData.GetError()
						return
					}

					newRecs[domn.Id] = append(newRecs[domn.Id], rec)
				}
				break
			case spec.Private6:
				for _, val := range deply.InstanceData.PrivateIps6 {
					rec := &domain.Record{
						Domain:     specRec.Domain,
						SubDomain:  specRec.Name,
						Deployment: deply.Id,
						Type:       domain.AAAA,
						Value:      val,
					}

					errData, e := rec.Validate(db)
					if e != nil {
						err = e
						return
					}
					if errData != nil {
						err = errData.GetError()
						return
					}

					newRecs[domn.Id] = append(newRecs[domn.Id], rec)
				}
				break
			case spec.OraclePublic:
				for _, val := range deply.InstanceData.OraclePublicIps {
					rec := &domain.Record{
						Domain:     specRec.Domain,
						SubDomain:  specRec.Name,
						Deployment: deply.Id,
						Type:       domain.A,
						Value:      val,
					}

					errData, e := rec.Validate(db)
					if e != nil {
						err = e
						return
					}
					if errData != nil {
						err = errData.GetError()
						return
					}

					newRecs[domn.Id] = append(newRecs[domn.Id], rec)
				}
				break
			case spec.OraclePublic6:
				for _, val := range deply.InstanceData.OraclePublicIps6 {
					rec := &domain.Record{
						Domain:     specRec.Domain,
						SubDomain:  specRec.Name,
						Deployment: deply.Id,
						Type:       domain.AAAA,
						Value:      val,
					}

					errData, e := rec.Validate(db)
					if e != nil {
						err = e
						return
					}
					if errData != nil {
						err = errData.GetError()
						return
					}

					newRecs[domn.Id] = append(newRecs[domn.Id], rec)
				}
				break
			case spec.OraclePrivate:
				for _, val := range deply.InstanceData.OraclePrivateIps {
					rec := &domain.Record{
						Domain:     specRec.Domain,
						SubDomain:  specRec.Name,
						Deployment: deply.Id,
						Type:       domain.A,
						Value:      val,
					}

					errData, e := rec.Validate(db)
					if e != nil {
						err = e
						return
					}
					if errData != nil {
						err = errData.GetError()
						return
					}

					newRecs[domn.Id] = append(newRecs[domn.Id], rec)
				}
				break
			}
		}

		for domnId, domnNewRecs := range newRecs {
			domn := d.stat.SpecDomain(domnId)
			if domn == nil {
				continue
			}

			changedDomn := domn.MergeRecords(deply.Id, domnNewRecs)
			if changedDomn != nil {
				d.domainCommit(deply, changedDomn, domnNewRecs)
			}
		}
	}

	return
}

func (d *Deployments) Deploy(db *database.Database) (err error) {
	activeDeployments := d.stat.DeploymentsDeployed()
	inactiveDeployments := d.stat.DeploymentsInactive()

	for _, deply := range inactiveDeployments {
		switch deply.Action {
		case deployment.Migrate:
			d.migrate(deply)
			break
		case deployment.Destroy:
			d.destroy(deply)
			break
		case deployment.Archive:
			err = d.archive(deply)
			if err != nil {
				return
			}
			break
		case deployment.Restore:
			err = d.restore(deply)
			if err != nil {
				return
			}
			break
		}
	}

	for _, deply := range activeDeployments {
		if deply.Action == deployment.Migrate {
			d.migrate(deply)
			break
		}

		if deply.State == deployment.Deployed &&
			deply.Kind == deployment.Instance {

			spec := d.stat.Spec(deply.Spec)
			if spec != nil && spec.Domain != nil {
				err = d.domain(db, deply, spec)
				if err != nil {
					return
				}
			}
		}

		if deply.Kind == deployment.Image {
			err = d.image(deply)
			if err != nil {
				return
			}
		}
	}

	return
}

func NewDeployments(stat *state.State) *Deployments {
	return &Deployments{
		stat: stat,
	}
}
