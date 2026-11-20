package netconf

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/ip"
)

func ifaceMasterMissing(ifacesData map[string]*ip.Iface,
	iface string) bool {

	ifaceData := ifacesData[iface]
	return ifaceData != nil && ifaceData.Master == ""
}

func (n *NetConf) Master(db *database.Database) (err error) {
	ifacesData, err := ip.GetIfacesCached("")
	if err != nil {
		return
	}

	if ifaceMasterMissing(ifacesData, n.SystemExternalIface) {
		err = n.externalMaster(db)
		if err != nil {
			return
		}
	}

	if ifaceMasterMissing(ifacesData, n.SystemInternalIface) {
		err = n.internalMaster(db)
		if err != nil {
			return
		}
	}

	if ifaceMasterMissing(ifacesData, n.SystemHostIface) {
		err = n.hostMaster(db)
		if err != nil {
			return
		}
	}

	if ifaceMasterMissing(ifacesData, n.SystemNodePortIface) {
		err = n.nodePortMaster(db)
		if err != nil {
			return
		}
	}

	return
}
