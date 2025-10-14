package netconf

import (
	"time"

	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
)

func (n *NetConf) hostNet(db *database.Database) (err error) {
	if n.HostNetwork {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"add", n.SystemHostIface,
			"type", "veth",
			"peer", "name", n.SpaceHostIface,
			"addr", n.HostMacAddr,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) hostMtu(db *database.Database) (err error) {
	if n.HostNetwork && n.SystemHostIfaceMtu != "" {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", n.SystemHostIface,
			"mtu", n.SystemHostIfaceMtu,
		)
		if err != nil {
			return
		}
	}
	if n.HostNetwork && n.SpaceHostIfaceMtu != "" {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", n.SpaceHostIface,
			"mtu", n.SpaceHostIfaceMtu,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) hostUp(db *database.Database) (err error) {
	if n.HostNetwork {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", n.SystemHostIface, "up",
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) hostMaster(db *database.Database) (err error) {
	if n.HostNetwork {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", n.SystemHostIface,
			"master", n.PhysicalHostIface,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) hostSpace(db *database.Database) (err error) {
	if n.HostNetwork {
		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "link",
			"set", "dev", n.SpaceHostIface,
			"netns", n.Namespace,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) hostSpaceUp(db *database.Database) (err error) {
	if n.HostNetwork {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"ip", "link",
			"set", "dev", n.SpaceHostIface, "up",
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) Host(db *database.Database) (err error) {
	delay := time.Duration(settings.Hypervisor.ActionRate) * time.Second
	lockId := lock.Lock("host")
	defer lock.DelayUnlock("host", lockId, delay)

	err = n.hostNet(db)
	if err != nil {
		return
	}

	err = n.hostMtu(db)
	if err != nil {
		return
	}

	err = n.hostUp(db)
	if err != nil {
		return
	}

	err = n.hostMaster(db)
	if err != nil {
		return
	}

	err = n.hostSpace(db)
	if err != nil {
		return
	}

	err = n.hostSpaceUp(db)
	if err != nil {
		return
	}

	return
}
