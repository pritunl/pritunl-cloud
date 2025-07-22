package netconf

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/imds"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
)

func (n *NetConf) imdsNet(db *database.Database) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		[]string{
			"File exists",
		},
		"ip", "netns", "exec", n.Namespace,
		"ip", "link",
		"add", n.SpaceImdsIface,
		"type", "dummy",
		//"addr", n.ImdsMacAddr,
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) imdsMtu(db *database.Database) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"ip", "link",
		"set", "dev", n.SpaceImdsIface,
		"mtu", n.ImdsIfaceMtu,
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) imdsAddr(db *database.Database) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"ip", "addr",
		"add", settings.Hypervisor.ImdsAddress,
		"dev", n.SpaceImdsIface,
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) imdsUp(db *database.Database) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"ip", "link",
		"set", "dev", n.SpaceImdsIface, "up",
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) imdsStart(db *database.Database) (err error) {
	err = imds.Start(db, n.Virt)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) Imds(db *database.Database) (err error) {
	err = n.imdsNet(db)
	if err != nil {
		return
	}

	err = n.imdsMtu(db)
	if err != nil {
		return
	}

	err = n.imdsAddr(db)
	if err != nil {
		return
	}

	err = n.imdsUp(db)
	if err != nil {
		return
	}

	err = n.imdsStart(db)
	if err != nil {
		return
	}

	return
}
