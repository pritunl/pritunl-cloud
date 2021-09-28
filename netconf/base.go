package netconf

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func (n *NetConf) Base(db *database.Database) (err error) {
	n.PhysicalExternalIfaceBridge, err = utils.IsInterfaceBridge(
		n.PhysicalExternalIface)
	if err != nil {
		return
	}
	n.PhysicalExternalIfaceBridge6, err = utils.IsInterfaceBridge(
		n.PhysicalExternalIface6)
	if err != nil {
		return
	}
	n.PhysicalInternalIfaceBridge, err = utils.IsInterfaceBridge(
		n.PhysicalInternalIface)
	if err != nil {
		return
	}

	return
}
