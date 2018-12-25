package interfaces

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"strings"
	"sync"
	"time"
)

var (
	ifaces     = map[string]set.Set{}
	ifacesLock = sync.Mutex{}
	lastChange time.Time
)

func getIfaces(bridge string) (ifacesSet set.Set, err error) {
	output, err := utils.ExecCombinedOutput("", "brctl", "show", bridge)
	if err != nil {
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

func SyncIfaces(force bool) {
	nde := node.Self

	if !force && time.Since(lastChange) < 15*time.Second {
		return
	}

	ifacesNew := map[string]set.Set{}

	externalIface := nde.ExternalInterface
	externalIfaces := nde.ExternalInterfaces
	internalIface := nde.InternalInterface
	internalIfaces := nde.InternalInterfaces

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
			if err != nil {
				continue
			}

			ifacesNew[iface] = ifaceSet
		}
	} else if internalIface != "" {
		ifaceSet, err := getIfaces(internalIface)
		if err == nil {
			ifacesNew[internalIface] = ifaceSet
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

func GetInternal(virtIface string) (internalIface string) {
	internalIfaces := node.Self.InternalInterfaces

	if internalIfaces != nil {
		curLen := 0

		for _, iface := range internalIfaces {
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
	} else {
		internalIface = node.Self.InternalInterface
	}

	if internalIface == "" {
		internalIface = settings.Local.BridgeName
	}

	return
}
