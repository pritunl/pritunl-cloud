package vxlan

import (
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/ip"
	"github.com/pritunl/pritunl-cloud/iproute"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

var (
	curIfaces         set.Set
	curDatabase       set.Set
	curDatabaseIfaces set.Set
)

func initIfaces(stat *state.State, internaIfaces []string) (err error) {
	ifaces := set.NewSet()
	newCurIfaces := set.NewSet()
	for _, iface := range stat.Interfaces() {
		ifaces.Add(iface)
		if len(iface) == 14 && (strings.HasPrefix(iface, "k") ||
			strings.HasPrefix(iface, "b")) {

			newCurIfaces.Add(iface)
		}
	}

	parentVxIfaces := map[string]string{}
	parentBrIfaces := map[string]string{}
	newIfaces := set.NewSet()
	for _, iface := range internaIfaces {
		vxIface := vm.GetHostVxlanIface(iface)
		brIface := vm.GetHostBridgeIface(iface)

		parentVxIfaces[vxIface] = iface
		parentBrIfaces[brIface] = iface

		newIfaces.Add(vxIface)
		newIfaces.Add(brIface)
	}

	remIfaces := newCurIfaces.Copy()
	remIfaces.Subtract(newIfaces)
	for ifaceInf := range remIfaces.Iter() {
		iface := ifaceInf.(string)

		if strings.HasPrefix(iface, "b") {
			logrus.WithFields(logrus.Fields{
				"bridge": iface,
			}).Info("vxlan: Removing bridge")

			_, _ = utils.ExecCombinedOutputLogged(
				[]string{
					"Cannot find device",
				},
				"ip", "link",
				"set", "dev",
				iface, "down",
			)
			_ = iproute.BridgeDelete("", iface)
		}
	}
	for ifaceInf := range remIfaces.Iter() {
		iface := ifaceInf.(string)

		if strings.HasPrefix(iface, "k") {
			logrus.WithFields(logrus.Fields{
				"vxlan": iface,
			}).Info("vxlan: Removing vxlan")

			_, _ = utils.ExecCombinedOutputLogged(
				[]string{
					"Cannot find device",
				},
				"ip", "link",
				"del", iface,
			)
		}
	}

	time.Sleep(200 * time.Millisecond)

	newCurIfaces.Intersect(newIfaces)
	for ifaceInf := range newCurIfaces.Iter() {
		iface := ifaceInf.(string)

		if strings.HasPrefix(iface, "b") {
			parentIface := parentBrIfaces[iface]
			if parentIface == "" {
				continue
			}

			vxIface := vm.GetHostVxlanIface(parentIface)
			if ifaces.Contains(vxIface) {
				_, err = utils.ExecCombinedOutputLogged(
					[]string{"does not exist"},
					"ip", "link", "set",
					vxIface, "master", iface,
				)
				if err != nil {
					return
				}
			}
		}
	}

	time.Sleep(300 * time.Millisecond)

	for ifaceInf := range newCurIfaces.Iter() {
		iface := ifaceInf.(string)

		if strings.HasPrefix(iface, "b") {
			parentIface := parentBrIfaces[iface]
			if parentIface == "" {
				continue
			}

			_, err = utils.ExecCombinedOutputLogged(
				nil,
				"ip", "link",
				"set", "dev",
				iface, "up",
			)
			if err != nil {
				return
			}
		}
	}

	time.Sleep(500 * time.Millisecond)

	for ifaceInf := range newCurIfaces.Iter() {
		iface := ifaceInf.(string)

		if strings.HasPrefix(iface, "k") {
			_, err = utils.ExecCombinedOutputLogged(
				nil,
				"ip", "link",
				"set", "dev",
				iface, "up",
			)
			if err != nil {
				return
			}
		}
	}

	ip.ClearIfacesCache("")

	curIfaces = newCurIfaces

	return
}

func initDatabase(stat *state.State, internaIfaces []string) (err error) {
	output, err := utils.ExecOutput("", "bridge", "fdb")
	if err != nil {
		return
	}

	nodeSelf := stat.Node()

	nodeDc := stat.NodeDatacenter()
	if nodeDc == nil {
		return
	}

	nodes := stat.Nodes()
	if nodes == nil {
		nodes = []*node.Node{}
	}

	newDb := set.NewSet()
	for _, nde := range nodes {
		if nde.Id == nodeSelf.Id || nde.Datacenter != nodeDc.Id ||
			nde.PrivateIps == nil || !nodeDc.Vxlan() {

			continue
		}

		for _, privateIp := range nde.PrivateIps {
			newDb.Add(privateIp)
		}
	}

	newCurDb := set.NewSet()
	newCurIfaces := set.NewSet()
	ifaceBridgeDb := map[string]set.Set{}

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 7 || fields[0] != "00:00:00:00:00:00" {
			continue
		}

		iface := fields[2]
		if len(iface) != 14 || !strings.HasPrefix(iface, "k") {
			continue
		}

		dest := fields[4]

		bridgeSet := ifaceBridgeDb[iface]
		if bridgeSet == nil {
			bridgeSet = set.NewSet()
			ifaceBridgeDb[iface] = bridgeSet
		}

		newCurIfaces.Add(iface)
		bridgeSet.Add(dest)
		newCurDb.Add(dest)
	}

	for ifaceInf := range newCurIfaces.Iter() {
		iface := ifaceInf.(string)
		ifaceDb := ifaceBridgeDb[iface]

		addDb := newDb.Copy()
		addDb.Subtract(ifaceDb)
		for destInf := range addDb.Iter() {
			dest := destInf.(string)
			if dest == "" {
				logrus.Warning("vxlan: Empty destination")
				continue
			}

			_, err = utils.ExecCombinedOutputLogged(
				nil,
				"bridge", "fdb",
				"append", "00:00:00:00:00:00",
				"dev", iface,
				"dst", dest,
			)
			if err != nil {
				return
			}
		}

		remDb := ifaceDb.Copy()
		remDb.Subtract(newDb)
		for destInf := range remDb.Iter() {
			dest := destInf.(string)
			if dest == "" {
				logrus.Warning("vxlan: Empty destination")
				continue
			}

			_, err = utils.ExecCombinedOutputLogged(
				[]string{
					"Cannot find device",
					"No such file",
				},
				"bridge", "fdb",
				"del", "00:00:00:00:00:00",
				"dev", iface,
				"dst", dest,
			)
			if err != nil {
				return
			}
		}

	}

	curDatabase = newCurDb
	curDatabaseIfaces = newCurIfaces

	return
}

func syncIfaces(stat *state.State, internaIfaces []string,
	ifacesData map[string]*ip.Iface, retry bool) (err error) {

	cIfaces := curIfaces
	nodeSelf := stat.Node()
	clearCache := false

	lostIfaces := set.NewSet()
	for ifaceInf := range cIfaces.Iter() {
		iface := ifaceInf.(string)
		if ifacesData[iface] == nil {
			logrus.WithFields(logrus.Fields{
				"iface": iface,
			}).Error("vxlan: Lost vxlan interface")
			lostIfaces.Add(iface)
		}
	}
	cIfaces.Subtract(lostIfaces)

	parentVxIfaces := map[string]string{}
	parentBrIfaces := map[string]string{}
	vxBrIfaces := map[string]string{}
	newIfaces := set.NewSet()
	if internaIfaces != nil && stat.VxLan() {
		for _, iface := range internaIfaces {
			vxIface := vm.GetHostVxlanIface(iface)
			brIface := vm.GetHostBridgeIface(iface)

			parentVxIfaces[vxIface] = iface
			parentBrIfaces[brIface] = iface
			vxBrIfaces[vxIface] = brIface

			newIfaces.Add(vxIface)
			newIfaces.Add(brIface)
		}
	}

	remIfaces := cIfaces.Copy()
	remIfaces.Subtract(newIfaces)
	for ifaceInf := range remIfaces.Iter() {
		iface := ifaceInf.(string)

		if strings.HasPrefix(iface, "b") {
			logrus.WithFields(logrus.Fields{
				"bridge": iface,
			}).Info("vxlan: Removing bridge")

			_, _ = utils.ExecCombinedOutputLogged(
				[]string{
					"Cannot find device",
				},
				"ip", "link",
				"set", "dev",
				iface, "down",
			)
			_ = iproute.BridgeDelete("", iface)

			clearCache = true
		}
	}
	for ifaceInf := range remIfaces.Iter() {
		iface := ifaceInf.(string)

		if strings.HasPrefix(iface, "k") {
			logrus.WithFields(logrus.Fields{
				"vxlan": iface,
			}).Info("vxlan: Removing vxlan")

			_, _ = utils.ExecCombinedOutputLogged(
				[]string{
					"Cannot find device",
				},
				"ip", "link",
				"del", iface,
			)

			clearCache = true
		}
	}

	addIfaces := newIfaces.Copy()
	addIfaces.Subtract(cIfaces)
	for ifaceInf := range addIfaces.Iter() {
		iface := ifaceInf.(string)

		if strings.HasPrefix(iface, "k") {
			vxId := settings.Hypervisor.VxlanId
			destPort := settings.Hypervisor.VxlanDestPort
			parentIface := parentVxIfaces[iface]

			localIp := ""
			if nodeSelf.PrivateIps != nil {
				localIp = nodeSelf.PrivateIps[parentIface]
			}

			if localIp == "" {
				if !retry {
					nodeSelf.SyncNetwork(true)
					err = syncIfaces(stat, internaIfaces, ifacesData, true)
					return
				}

				err = &errortypes.NotFoundError{
					errors.New("vxlan: Missing private IP for " +
						"internal interface"),
				}
				return
			}

			logrus.WithFields(logrus.Fields{
				"vxlan": iface,
			}).Info("vxlan: Adding vxlan")

			_, err = utils.ExecCombinedOutputLogged(
				[]string{
					"File exists",
				},
				"ip", "link",
				"add", iface,
				"type", "vxlan",
				"id", strconv.Itoa(vxId),
				"local", localIp,
				"dstport", strconv.Itoa(destPort),
				"dev", parentIface,
			)
			if err != nil {
				return
			}

			_, err = utils.ExecCombinedOutputLogged(
				nil,
				"ip", "link",
				"set", "dev",
				iface, "up",
			)
			if err != nil {
				return
			}

			clearCache = true
		}
	}

	for ifaceInf := range addIfaces.Iter() {
		iface := ifaceInf.(string)

		if strings.HasPrefix(iface, "b") {
			parentIface := parentBrIfaces[iface]

			logrus.WithFields(logrus.Fields{
				"bridge": iface,
			}).Info("vxlan: Adding bridge")

			err = iproute.BridgeAdd("", iface)
			if err != nil {
				return
			}

			_, err = utils.ExecCombinedOutputLogged(
				nil,
				"ip", "link", "set",
				vm.GetHostVxlanIface(parentIface), "master", iface,
			)
			if err != nil {
				return
			}

			_, err = utils.ExecCombinedOutputLogged(
				nil,
				"ip", "link",
				"set", "dev",
				iface, "up",
			)
			if err != nil {
				return
			}

			clearCache = true
		}
	}

	existIfaces := cIfaces.Copy()
	existIfaces.Subtract(remIfaces)
	for ifaceInf := range existIfaces.Iter() {
		iface := ifaceInf.(string)

		if strings.HasPrefix(iface, "k") {
			brIface := vxBrIfaces[iface]

			ifaceData := ifacesData[iface]
			if ifaceData != nil && ifaceData.Master != brIface {
				logrus.WithFields(logrus.Fields{
					"vxlan":  iface,
					"bridge": brIface,
				}).Warn("vxlan: Correct vxlan master")

				_, err = utils.ExecCombinedOutputLogged(
					[]string{"does not exist"},
					"ip", "link", "set",
					iface, "master", brIface,
				)
				if err != nil {
					return
				}

				clearCache = true
			}
		}
	}

	if clearCache {
		ip.ClearIfacesCache("")
	}

	curIfaces = newIfaces

	return
}

func syncDatabase(stat *state.State, internaIfaces []string) (err error) {
	nodeSelf := stat.Node()
	cDatabase := curDatabase
	cIfaces := curDatabaseIfaces

	nodes := stat.Nodes()
	if nodes == nil {
		nodes = []*node.Node{}
	}

	nodeDc := stat.NodeDatacenter()
	if nodeDc == nil {
		return
	}

	newIfaces := set.NewSet()
	for _, iface := range internaIfaces {
		newIfaces.Add(vm.GetHostVxlanIface(iface))
	}

	newDb := set.NewSet()
	for _, nde := range nodes {
		if nde.Id == nodeSelf.Id || nde.Datacenter != nodeDc.Id ||
			nde.PrivateIps == nil || !nodeDc.Vxlan() {

			continue
		}

		for _, privateIp := range nde.PrivateIps {
			newDb.Add(privateIp)
		}
	}

	addDb := newDb.Copy()
	addDb.Subtract(cDatabase)
	for destInf := range addDb.Iter() {
		dest := destInf.(string)
		if dest == "" {
			logrus.Warning("vxlan: Empty destination")
			continue
		}

		for ifaceInf := range newIfaces.Iter() {
			iface := ifaceInf.(string)

			_, err = utils.ExecCombinedOutputLogged(
				nil,
				"bridge", "fdb",
				"append", "00:00:00:00:00:00",
				"dev", iface,
				"dst", dest,
			)
			if err != nil {
				return
			}
		}
	}

	remDb := cDatabase.Copy()
	remDb.Subtract(newDb)
	for destInf := range remDb.Iter() {
		dest := destInf.(string)
		if dest == "" {
			logrus.Warning("vxlan: Empty destination")
			continue
		}

		for ifaceInf := range newIfaces.Iter() {
			iface := ifaceInf.(string)

			_, err = utils.ExecCombinedOutputLogged(
				[]string{
					"Cannot find device",
					"No such file",
				},
				"bridge", "fdb",
				"del", "00:00:00:00:00:00",
				"dev", iface,
				"dst", dest,
			)
			if err != nil {
				return
			}
		}
	}

	addIfaces := newIfaces.Copy()
	addIfaces.Subtract(cIfaces)
	for ifaceInf := range addIfaces.Iter() {
		iface := ifaceInf.(string)

		for destInf := range newDb.Iter() {
			dest := destInf.(string)
			if dest == "" {
				logrus.Warning("vxlan: Empty destination")
				continue
			}

			_, err = utils.ExecCombinedOutputLogged(
				nil,
				"bridge", "fdb",
				"append", "00:00:00:00:00:00",
				"dev", iface,
				"dst", dest,
			)
			if err != nil {
				return
			}
		}
	}

	curDatabase = newDb
	curDatabaseIfaces = newIfaces

	return
}

func ApplyState(stat *state.State) (err error) {
	nodeSelf := stat.Node()
	internaIfaces := nodeSelf.InternalInterfaces

	if curIfaces == nil {
		err = initIfaces(stat, internaIfaces)
		if err != nil {
			return
		}
	}

	if curDatabase == nil {
		err = initDatabase(stat, internaIfaces)
		if err != nil {
			return
		}
	}

	ifacesData, err := ip.GetIfacesCached("")
	if err != nil {
		return
	}

	err = syncIfaces(stat, internaIfaces, ifacesData, false)
	if err != nil {
		return
	}

	err = syncDatabase(stat, internaIfaces)
	if err != nil {
		return
	}

	return
}
