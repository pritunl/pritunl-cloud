package netconf

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func (n *NetConf) nodePortNet(db *database.Database) (err error) {
	if n.NodePortNetwork {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"add", n.SystemNodePortIface,
			"type", "veth",
			"peer", "name", n.SpaceNodePortIface,
			"addr", n.NodePortMacAddr,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) nodePortMtu(db *database.Database) (err error) {
	if n.NodePortNetwork && n.SystemNodePortIfaceMtu != "" {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", n.SystemNodePortIface,
			"mtu", n.SystemNodePortIfaceMtu,
		)
		if err != nil {
			return
		}
	}
	if n.NodePortNetwork && n.SpaceNodePortIfaceMtu != "" {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", n.SpaceNodePortIface,
			"mtu", n.SpaceNodePortIfaceMtu,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) nodePortUp(db *database.Database) (err error) {
	if n.NodePortNetwork {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", n.SystemNodePortIface, "up",
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) nodePortMaster(db *database.Database) (err error) {
	if n.NodePortNetwork {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", n.SystemNodePortIface,
			"master", n.PhysicalNodePortIface,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) nodePortSpace(db *database.Database) (err error) {
	if n.NodePortNetwork {
		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "link",
			"set", "dev", n.SpaceNodePortIface,
			"netns", n.Namespace,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) nodePortSpaceUp(db *database.Database) (err error) {
	if n.NodePortNetwork {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"ip", "link",
			"set", "dev", n.SpaceNodePortIface, "up",
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) NodePort(db *database.Database) (err error) {
	err = n.nodePortNet(db)
	if err != nil {
		return
	}

	err = n.nodePortMtu(db)
	if err != nil {
		return
	}

	err = n.nodePortUp(db)
	if err != nil {
		return
	}

	err = n.nodePortMaster(db)
	if err != nil {
		return
	}

	err = n.nodePortSpace(db)
	if err != nil {
		return
	}

	err = n.nodePortSpaceUp(db)
	if err != nil {
		return
	}

	return
}
