package netconf

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/ip"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

var (
	lastCorrect time.Time
)

type missingMaster struct {
	Virt   *vm.VirtualMachine
	Ifaces []string
}

func getMissingMasters(instances []*instance.Instance,
	virtsMap map[bson.ObjectID]*vm.VirtualMachine,
	ifacesData map[string]*ip.Iface) (missing []*missingMaster) {

	missing = []*missingMaster{}

	for _, inst := range instances {
		if inst.Action != instance.Start {
			continue
		}

		virt := virtsMap[inst.Id]
		if virt == nil || virt.State != vm.Running {
			continue
		}

		ifaces := []string{
			vm.GetIfaceNodeExternal(inst.Id, 0),
			vm.GetIfaceNodeInternal(inst.Id, 0),
			vm.GetIfaceHost(inst.Id, 0),
			vm.GetIfaceNodePort(inst.Id, 0),
		}

		var mstr *missingMaster
		for _, iface := range ifaces {
			if !ifaceMasterMissing(ifacesData, iface) {
				continue
			}

			if mstr == nil {
				mstr = &missingMaster{
					Virt:   virt,
					Ifaces: []string{},
				}
				missing = append(missing, mstr)
			}
			mstr.Ifaces = append(mstr.Ifaces, iface)
		}
	}

	return
}

func checkState(db *database.Database, instances []*instance.Instance,
	virtsMap map[bson.ObjectID]*vm.VirtualMachine) (err error) {

	lastCorrect = time.Now()

	ip.ClearIfacesCache("")

	ifacesData, err := ip.GetIfacesCached("")
	if err != nil {
		return
	}

	missing := getMissingMasters(instances, virtsMap, ifacesData)
	if len(missing) == 0 {
		return
	}

	for _, mstr := range missing {
		logrus.WithFields(logrus.Fields{
			"instance_id": mstr.Virt.Id.Hex(),
			"ifaces":      mstr.Ifaces,
		}).Warn("netconf: Repairing instance interface master")

		nc := New(mstr.Virt)

		e := nc.Repair(db)
		if e != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": mstr.Virt.Id.Hex(),
				"ifaces":      mstr.Ifaces,
				"error":       e,
			}).Error("netconf: Failed to repair instance interface master")
		}
	}

	ip.ClearIfacesCache("")

	return
}

func ApplyState(db *database.Database, instances []*instance.Instance,
	virtsMap map[bson.ObjectID]*vm.VirtualMachine) (err error) {

	ifacesData, err := ip.GetIfacesCached("")
	if err != nil {
		return
	}

	missing := getMissingMasters(instances, virtsMap, ifacesData)
	if len(missing) == 0 || time.Since(lastCorrect) < time.Duration(
		settings.Hypervisor.NetRepairTtl)*time.Second {

		return
	}

	err = checkState(db, instances, virtsMap)
	if err != nil {
		return
	}

	return
}
