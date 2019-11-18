package interfaces

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/iproute"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/vm"
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

	if vxlan == curVxlan && time.Since(lastChange) < 30*time.Second {
		return
	}
	curVxlan = vxlan

	ifacesNew := map[string]set.Set{}

	externalIface := nde.ExternalInterface
	externalIfaces := nde.ExternalInterfaces
	internalIface := nde.InternalInterface
	internalIfaces := nde.InternalInterfaces
	blocks := nde.Blocks

	if externalIfaces != nil {
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
	} else if externalIface != "" {
		ifaceSet, err := getIfaces(externalIface)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"bridge": externalIface,
				"error":  err,
			}).Error("interfaces: Bridge ifaces get failed")
		} else {
			ifacesNew[externalIface] = ifaceSet
		}
	}

	if internalIfaces != nil {
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

			vxIface := vm.GetHostBridgeIface(iface)
			ifaceSet, err = getIfaces(vxIface)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"bridge": vxIface,
					"error":  err,
				}).Error("interfaces: Bridge ifaces get failed")
			} else {
				ifacesNew[vxIface] = ifaceSet
			}

		}
	} else if internalIface != "" {
		ifaceSet, err := getIfaces(internalIface)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"bridge": internalIface,
				"error":  err,
			}).Error("interfaces: Bridge ifaces get failed")
		} else {
			ifacesNew[internalIface] = ifaceSet
		}
	}

	if blocks != nil {
		for _, blck := range blocks {
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
	} else {
		externalIface = node.Self.ExternalInterface
	}

	return
}

func HasExternal() (exists bool) {
	externalIfaces := node.Self.ExternalInterfaces
	externalIface := ""

	if externalIfaces != nil {
		if len(externalIfaces) > 0 {
			externalIface = externalIfaces[0]
		}
	} else {
		externalIface = node.Self.ExternalInterface
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
	} else if !vxlan {
		internalIface = node.Self.InternalInterface
	}

	return
}

func GetBridges(nde *node.Node) (bridges set.Set) {
	bridges = set.NewSet()

	externalIfaces := nde.ExternalInterfaces
	if externalIfaces != nil {
		for _, iface := range externalIfaces {
			bridges.Add(iface)
		}
	} else {
		externalIface := nde.ExternalInterface
		if externalIface != "" {
			bridges.Add(externalIface)
		}
	}

	internalIfaces := nde.InternalInterfaces
	if internalIfaces != nil {
		for _, iface := range internalIfaces {
			bridges.Add(iface)
		}
	} else {
		internalIface := nde.InternalInterface
		if internalIface != "" {
			bridges.Add(internalIface)
		}
	}

	ndeBlocks := nde.Blocks
	if ndeBlocks != nil {
		for _, blck := range ndeBlocks {
			bridges.Add(blck.Interface)
		}
	}

	return
}

func RemoveVirtIface(virtIface string) {
	ifacesLock.Lock()
	lastChange = time.Now()
	for iface, ifaceSet := range ifaces {
		ifaceSet.Remove(virtIface)
		ifaces[iface] = ifaceSet
	}
	ifacesLock.Unlock()
}
