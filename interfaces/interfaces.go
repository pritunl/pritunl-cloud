package interfaces

import (
	"sync"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/iproute"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

var (
	curVxlan   = false
	ifaces     = map[string]set.Set{}
	ifacesLock = sync.Mutex{}
	lastChange time.Time
)

func getIfaces(bridge string) (ifacesSet set.Set, err error) {
	ifacesSet = set.NewSet()

	ifaces, err := iproute.IfaceGetBridgeIfaces("", bridge)
	if err != nil {
		return
	}

	for _, iface := range ifaces {
		ifacesSet.Add(iface.Name)
	}

	return
}

func SyncIfaces(vxlan bool) {
	nde := node.Self

	if vxlan == curVxlan && time.Since(lastChange) < 10*time.Second {
		return
	}
	curVxlan = vxlan

	ifacesNew := map[string]set.Set{}

	externalIfaces := nde.ExternalInterfaces
	internalIfaces := nde.InternalInterfaces
	blocks := nde.Blocks

	for _, iface := range externalIfaces {
		ifaceSet, err := getIfaces(iface)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"bridge": iface,
				"error":  err,
			}).Error("interfaces: Bridge ifaces get failed")
		} else {
			ifacesNew[iface] = ifaceSet
		}
	}

	for _, iface := range internalIfaces {
		ifaceSet, err := getIfaces(iface)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"bridge": iface,
				"error":  err,
			}).Error("interfaces: Bridge ifaces get failed")
		} else {
			ifacesNew[iface] = ifaceSet
		}

		brIface := vm.GetHostBridgeIface(iface)
		ifaceSet, err = getIfaces(brIface)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"bridge": brIface,
				"error":  err,
			}).Error("interfaces: Bridge ifaces get failed")
		} else {
			ifacesNew[brIface] = ifaceSet
		}
	}

	for _, blck := range blocks {
		if blck.Interface == "" {
			continue
		}

		ifaceSet, err := getIfaces(blck.Interface)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"bridge": blck.Interface,
				"error":  err,
			}).Error("interfaces: Bridge ifaces get failed")
		} else {
			ifacesNew[blck.Interface] = ifaceSet
		}
	}

	ifacesLock.Lock()
	lastChange = time.Now()
	ifaces = ifacesNew
	ifacesLock.Unlock()

	return
}

func GetExternal(virtIface string) (externalIface string) {
	externalIfaces := node.Self.ExternalInterfaces

	if externalIfaces != nil {
		curLen := 0

		for _, iface := range externalIfaces {
			ifacesSetLen := 0

			ifacesLock.Lock()
			ifacesSet := ifaces[iface]
			if ifacesSet != nil {
				ifacesSetLen = ifacesSet.Len()
			}
			ifacesLock.Unlock()

			if ifacesSet == nil {
				continue
			}

			if externalIface == "" || ifacesSetLen < curLen {
				curLen = ifacesSetLen
				externalIface = iface
			}
		}

		if externalIface != "" {
			ifacesLock.Lock()
			lastChange = time.Now()
			ifacesSet := ifaces[externalIface]
			if ifacesSet != nil {
				ifacesSet.Add(virtIface)
			}
			ifacesLock.Unlock()
		}
	}

	return
}

func HasExternal() (exists bool) {
	externalIfaces := node.Self.ExternalInterfaces
	externalIface := ""

	if len(externalIfaces) > 0 {
		externalIface = externalIfaces[0]
	}

	if externalIface != "" {
		exists = true
	}

	return
}

func GetInternal(virtIface string, vxlan bool) (internalIface string) {
	internalIfaces := node.Self.InternalInterfaces

	if internalIfaces != nil {
		curLen := 0

		for _, iface := range internalIfaces {
			if vxlan {
				iface = vm.GetHostBridgeIface(iface)
			}

			ifacesSetLen := 0

			ifacesLock.Lock()
			ifacesSet := ifaces[iface]
			if ifacesSet != nil {
				ifacesSetLen = ifacesSet.Len()
			}
			ifacesLock.Unlock()

			if ifacesSet == nil {
				continue
			}

			if internalIface == "" || ifacesSetLen < curLen {
				curLen = ifacesSetLen
				internalIface = iface
			}
		}

		if internalIface != "" {
			ifacesLock.Lock()
			lastChange = time.Now()
			ifacesSet := ifaces[internalIface]
			if ifacesSet != nil {
				ifacesSet.Add(virtIface)
			}
			ifacesLock.Unlock()
		}
	}

	return
}

func GetBridges(nde *node.Node) (bridges set.Set) {
	bridges = set.NewSet()

	externalIfaces := nde.ExternalInterfaces
	for _, iface := range externalIfaces {
		bridges.Add(iface)
	}

	internalIfaces := nde.InternalInterfaces
	for _, iface := range internalIfaces {
		bridges.Add(iface)
	}

	ndeBlocks := nde.Blocks
	for _, blck := range ndeBlocks {
		if blck.Interface == "" {
			continue
		}
		bridges.Add(blck.Interface)
	}

	ndeBlocks6 := nde.Blocks6
	for _, blck := range ndeBlocks6 {
		if blck.Interface == "" {
			continue
		}
		bridges.Add(blck.Interface)
	}

	return
}

func GetBridgesInternal(nde *node.Node) (bridges set.Set) {
	bridges = set.NewSet()

	internalIfaces := nde.InternalInterfaces
	for _, iface := range internalIfaces {
		bridges.Add(iface)
	}

	return
}

func GetBridgesExternal(nde *node.Node) (bridges set.Set) {
	bridges = set.NewSet()

	externalIfaces := nde.ExternalInterfaces
	for _, iface := range externalIfaces {
		bridges.Add(iface)
	}

	ndeBlocks := nde.Blocks
	for _, blck := range ndeBlocks {
		if blck.Interface == "" {
			continue
		}
		bridges.Add(blck.Interface)
	}

	ndeBlocks6 := nde.Blocks6
	for _, blck := range ndeBlocks6 {
		if blck.Interface == "" {
			continue
		}
		bridges.Add(blck.Interface)
	}

	return
}

func RemoveVirtIface(virtIface string) {
	if virtIface == "" {
		return
	}

	ifacesLock.Lock()
	lastChange = time.Now()
	for iface, ifaceSet := range ifaces {
		ifaceSet.Remove(virtIface)
		ifaces[iface] = ifaceSet
	}
	ifacesLock.Unlock()
}
