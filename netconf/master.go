package netconf

import (
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/ip"
	"github.com/pritunl/pritunl-cloud/utils"
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

	ifaces := []string{}
	var ifaceErr error

	if ifaceMasterMissing(ifacesData, n.SystemExternalIface) {
		ifaces = append(ifaces, n.SystemExternalIface)

		err = n.externalMaster(db)
		if err != nil {
			ifaceErr = err
		}
	}

	if ifaceMasterMissing(ifacesData, n.SystemInternalIface) {
		ifaces = append(ifaces, n.SystemInternalIface)

		err = n.internalMaster(db)
		if err != nil {
			ifaceErr = err
		}
	}

	if ifaceMasterMissing(ifacesData, n.SystemHostIface) {
		ifaces = append(ifaces, n.SystemHostIface)

		err = n.hostMaster(db)
		if err != nil {
			ifaceErr = err
		}
	}

	if ifaceMasterMissing(ifacesData, n.SystemNodePortIface) {
		ifaces = append(ifaces, n.SystemNodePortIface)

		err = n.nodePortMaster(db)
		if err != nil {
			ifaceErr = err
		}
	}

	if ifaceErr != nil {
		err = ifaceErr
		return
	}

	for _, iface := range ifaces {
		port, e := utils.IsInterfaceBridgePort(iface)
		if e != nil {
			err = e
			return
		}

		if !port {
			err = &errortypes.NetworkError{
				errors.Newf("netconf: Interface '%s' master missing", iface),
			}
			return
		}
	}

	return
}
