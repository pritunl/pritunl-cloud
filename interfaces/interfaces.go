package interfaces

import (
	"strings"
	"sync"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
)

var (
	curVxlan   = false
	ifaces     = map[string]set.Set{}
	ifacesLock = sync.Mutex{}
	lastChange time.Time
)

func getIfaces(bridge string) (ifacesSet set.Set, err error) {
	output, err := utils.ExecCombinedOutput("", "brctl", "show", bridge)
	if err != nil {
		return
	}

	if strings.Contains(output, "Operation not supported") {
		err = &errortypes.ReadError{
			errors.New("interfaces: Operation not supported"),
		}
		return
	}

	if strings.Contains(output, "No such device") {
		err = &errortypes.ReadError{
			errors.New("interfaces: No such device"),
		}
		return
	}

	ifacesSet = set.NewSet()

	for i, line := range strings.Split(output, "\n") {
		if i < 2 {
			continue
		}

		line = strings.TrimSpace(line)
		if len(line) == 14 {
			ifacesSet.Add(line)
		}
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
				continue
			}

			ifacesNew[iface] = ifaceSet
		}
	} else if externalIface != "" {
		ifaceSet, err := getIfaces(externalIface)
		if err == nil {
			ifacesNew[externalIface] = ifaceSet
		}
	}

	if internalIfaces != nil {
		for _, iface := range internalIfaces {
			ifaceSet, err := getIfaces(iface)
			if err == nil {
				ifacesNew[iface] = ifaceSet
			}

			vxIface := vm.GetHostBridgeIface(iface)
			ifaceSet, err = getIfaces(vxIface)
			if err == nil {
				ifacesNew[vxIface] = ifaceSet
			}

		}
	} else if internalIface != "" {
		ifaceSet, err := getIfaces(internalIface)
		if err == nil {
			ifacesNew[internalIface] = ifaceSet
		}
	}

	if blocks != nil {
		for _, blck := range blocks {
			ifaceSet, err := getIfaces(blck.Interface)
			if err != nil {
				continue
			}

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
	} else {
		externalIface = node.Self.ExternalInterface
	}

	if externalIface == "" {
		externalIface = settings.Local.BridgeName
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

	if externalIface == "" {
		externalIface = settings.Local.BridgeName
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

	if internalIface == "" && !vxlan {
		internalIface = settings.Local.BridgeName
	}

	return
}

func GetBridges() (bridges set.Set) {
	bridges = set.NewSet()

	externalIfaces := node.Self.ExternalInterfaces
	if externalIfaces != nil {
		for _, iface := range externalIfaces {
			bridges.Add(iface)
		}
	} else {
		externalIface := node.Self.ExternalInterface
		if externalIface != "" {
			bridges.Add(externalIface)
		}
	}

	internalIfaces := node.Self.InternalInterfaces
	if internalIfaces != nil {
		for _, iface := range internalIfaces {
			bridges.Add(iface)
		}
	} else {
		internalIface := node.Self.InternalInterface
		if internalIface != "" {
			bridges.Add(internalIface)
		}
	}

	bridge := settings.Local.BridgeName
	if bridge != "" {
		bridges.Add(bridge)
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
