package netconf

import (
	"time"

	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
)

func (n *NetConf) internalNet(db *database.Database) (err error) {
	if n.PhysicalInternalIfaceBridge {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"add", n.SystemInternalIface,
			"type", "veth",
			"peer", "name", n.SpaceInternalIface,
			"addr", n.InternalMacAddr,
		)
		if err != nil {
			return
		}

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"tc", "qdisc", "replace", "dev", n.SystemInternalIface,
			"root", "fq_codel",
		)
		if err != nil {
			return err
		}
	} else {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"add", n.SpaceInternalIface,
			"addr", n.InternalMacAddr,
			"link", n.PhysicalInternalIface,
			"type", "macvlan",
			"mode", "bridge",
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) internalMtu(db *database.Database) (err error) {
	if n.SystemInternalIfaceMtu != "" && n.PhysicalInternalIfaceBridge {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", n.SystemInternalIface,
			"mtu", n.SystemInternalIfaceMtu,
		)
		if err != nil {
			return
		}
	}
	if n.SpaceInternalIfaceMtu != "" {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", n.SpaceInternalIface,
			"mtu", n.SpaceInternalIfaceMtu,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) internalUp(db *database.Database) (err error) {
	if n.PhysicalInternalIfaceBridge {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", n.SystemInternalIface, "up",
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) internalMaster(db *database.Database) (err error) {
	if n.PhysicalInternalIfaceBridge {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", n.SystemInternalIface,
			"master", n.PhysicalInternalIface,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) internalSpace(db *database.Database) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "link",
		"set", "dev", n.SpaceInternalIface,
		"netns", n.Namespace,
	)
	if err != nil {
		return
	}

	if n.PhysicalInternalIfaceBridge {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"tc", "qdisc", "replace", "dev", n.SpaceInternalIface,
			"root", "fq_codel",
		)
		if err != nil {
			return err
		}
	}

	return
}

func (n *NetConf) internalSpaceUp(db *database.Database) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"ip", "link",
		"set", "dev", n.SpaceInternalIface, "up",
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) Internal(db *database.Database) (err error) {
	delay := time.Duration(settings.Hypervisor.ActionRate) * time.Second
	lockId := lock.Lock("internal")
	defer lock.DelayUnlock("internal", lockId, delay)

	err = n.internalNet(db)
	if err != nil {
		return
	}

	err = n.internalMtu(db)
	if err != nil {
		return
	}

	err = n.internalUp(db)
	if err != nil {
		return
	}

	err = n.internalMaster(db)
	if err != nil {
		return
	}

	err = n.internalSpace(db)
	if err != nil {
		return
	}

	err = n.internalSpaceUp(db)
	if err != nil {
		return
	}

	return
}
