package netconf

import (
	"path"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/interfaces"
	"github.com/pritunl/pritunl-cloud/store"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/tools/commander"
	"github.com/sirupsen/logrus"
)

var (
	clearLock = utils.NewTimeoutLock(1 * time.Minute)
)

func (n *NetConf) Clear(db *database.Database) (err error) {
	lockId := clearLock.Lock()
	defer clearLock.Unlock(lockId)

	clearIface("", n.SystemExternalIface) // bridged
	clearIface("", n.SystemInternalIface) // bridged
	clearIface("", n.SystemHostIface)     // bridged
	clearIface("", n.SystemNodePortIface) // bridged

	clearIface("", n.SpaceExternalIface)
	clearIface("", n.SpaceInternalIface)
	if ifaceExists("", n.SpaceInternalIfaceOld) {
		clearIface("", n.SpaceInternalIfaceOld)
	}
	clearIface("", n.SpaceHostIface)
	clearIface("", n.SpaceNodePortIface)

	clearIface(n.Namespace, n.SpaceExternalIface)
	clearIface(n.Namespace, n.SpaceExternalIfaceMod)
	clearIface(n.Namespace, n.SpaceExternalIfaceMod6)
	clearIface(n.Namespace, n.SpaceInternalIface)
	if ifaceExists(n.Namespace, n.SpaceInternalIfaceOld) {
		clearIface(n.Namespace, n.SpaceInternalIfaceOld)
	}
	clearIface(n.Namespace, n.SpaceHostIface)
	clearIface(n.Namespace, n.SpaceNodePortIface)
	clearIface(n.Namespace, n.SpaceBridgeIface)
	clearIface(n.Namespace, n.SpaceImdsIface)

	clearIface(n.Namespace, n.SpaceCloudIface)
	clearIface(n.Namespace, n.SpaceCloudVirtIface)
	clearIface("", n.SpaceCloudVirtIface)
	restoreIface(n.Namespace, n.SpaceCloudIface)

	interfaces.RemoveVirtIface(n.SystemExternalIface)
	interfaces.RemoveVirtIface(n.SystemInternalIface)
	interfaces.RemoveVirtIface(n.SystemNodePortIface)

	time.Sleep(300 * time.Millisecond)

	return
}

func (n *NetConf) ClearAll(db *database.Database) (err error) {
	if len(n.Virt.NetworkAdapters) == 0 {
		err = &errortypes.NotFoundError{
			errors.New("qemu: Missing network interfaces"),
		}
		return
	}

	err = n.Clear(db)
	if err != nil {
		return
	}

	store.RemAddress(n.Virt.Id)
	store.RemRoutes(n.Virt.Id)
	store.RemArp(n.Virt.Id)

	return
}

func ifaceExists(namespace, iface string) (exists bool) {
	if namespace == "" {
		found, e := utils.Exists(
			path.Join("/", "sys", "class", "net", iface))
		if e != nil {
			logrus.WithFields(logrus.Fields{
				"iface": iface,
				"error": e,
			}).Error("netconf: Failed to check interface")
			return
		}

		exists = found
		return
	}

	_, err := commander.Exec(&commander.Opt{
		Name: "ip",
		Args: []string{
			"netns", "exec", namespace,
			"test", "-e", path.Join("/", "sys", "class", "net", iface),
		},
		PipeOut: true,
		PipeErr: true,
	})
	if err != nil {
		return
	}

	exists = true
	return
}

func clearIface(namespace, iface string) {
	if iface == "" {
		return
	}

	// iproute if<n> prefix fix
	if strings.HasPrefix(iface, "if") && !ifaceExists(namespace, iface) {
		return
	}

	if namespace != "" {
		_, e := commander.Exec(&commander.Opt{
			Name: "ip",
			Args: []string{
				"netns", "exec", namespace,
				"ip", "link", "set", iface, "nomaster",
			},
			PipeOut: true,
			PipeErr: true,
		})
		if e == nil {
			time.Sleep(200 * time.Millisecond)
		}
		commander.Exec(&commander.Opt{
			Name: "ip",
			Args: []string{
				"netns", "exec", namespace,
				"ip", "link", "set", iface, "down",
			},
			PipeOut: true,
			PipeErr: true,
		})
		commander.Exec(&commander.Opt{
			Name: "ip",
			Args: []string{
				"netns", "exec", namespace,
				"ip", "link", "del", iface,
			},
			PipeOut: true,
			PipeErr: true,
		})
	} else {
		_, e := commander.Exec(&commander.Opt{
			Name: "ip",
			Args: []string{
				"link", "set", iface, "nomaster",
			},
			PipeOut: true,
			PipeErr: true,
		})
		if e == nil {
			time.Sleep(200 * time.Millisecond)
		}
		commander.Exec(&commander.Opt{
			Name: "ip",
			Args: []string{
				"link", "set", iface, "down",
			},
			PipeOut: true,
			PipeErr: true,
		})
		commander.Exec(&commander.Opt{
			Name: "ip",
			Args: []string{
				"link", "del", iface,
			},
			PipeOut: true,
			PipeErr: true,
		})
	}

}

func restoreIface(namespace, iface string) {
	if namespace == "" || iface == "" || !ifaceExists(namespace, iface) {
		return
	}

	utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ip", "link", "set", "dev", iface, "netns", "1",
	)
}
