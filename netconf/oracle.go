package netconf

import (
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/namespace"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/oracle"
)

func (n *NetConf) oracleInitVnic(db *database.Database) (err error) {
	pv, err := oracle.NewProvider(node.Self.GetOracleAuthProvider())
	if err != nil {
		return
	}

	var vnic *oracle.Vnic

	found := false
	if n.Virt.OracleVnic != "" {
		vnic, err = oracle.GetVnic(pv, n.Virt.OracleVnic)
		if err != nil {
			return
		}

		if vnic.SubnetId != n.Virt.OracleSubnet {
			err = oracle.RemoveVnic(pv, n.Virt.OracleVnic)
			if err != nil {
				return
			}

			vnic = nil
		} else if !n.OracleSubnets.Contains(vnic.SubnetId) {
			err = oracle.RemoveVnic(pv, n.Virt.OracleVnic)
			if err != nil {
				return
			}

			vnic = nil
		} else {
			found = true
		}
	}

	if !n.OracleSubnets.Contains(n.Virt.OracleSubnet) {
		err = &errortypes.NotFoundError{
			errors.New("netconf: Invalid oracle subnet"),
		}
		return
	}

	if !found {
		vnicId, vnicAttachId, e := oracle.CreateVnic(
			pv, n.Virt.Id.Hex(), n.Virt.OracleSubnet)
		if e != nil {
			err = e
			return
		}

		n.Virt.OracleVnic = vnicId
		n.Virt.OracleVnicAttach = vnicAttachId
		err = n.Virt.CommitOracleVnic(db)
		if err != nil {
			_ = oracle.RemoveVnic(pv, vnicId)
			return
		}
	}

	return
}

func (n *NetConf) oracleConfVnic(db *database.Database) (err error) {
	curNamespace := ""
	found := false

	for i := 0; i < 60; i++ {
		time.Sleep(1 * time.Second)

		ifaces, e := oracle.GetIfaces(i == 59)
		if e != nil {
			err = e
			return
		}

		for _, iface := range ifaces {
			if iface.VnicId == n.Virt.OracleVnic {
				curNamespace = iface.Namespace

				n.Virt.OracleIp = iface.Address
				err = n.Virt.CommitOracleIps(db)
				if err != nil {
					return
				}

				found = true

				break
			}
		}

		if found {
			break
		}

		err = oracle.ConfIfaces(i == 59)
		if err != nil {
			return
		}
	}

	if !found {
		err = &errortypes.NotFoundError{
			errors.New("netconf: Failed to find vnic"),
		}
		return
	}

	if curNamespace != n.Namespace {
		err = namespace.Rename(curNamespace, n.Namespace)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) Oracle(db *database.Database) (err error) {
	if n.NetworkMode != node.Oracle || n.Virt.OracleSubnet == "" {
		return
	}

	err = n.oracleInitVnic(db)
	if err != nil {
		return
	}

	err = n.oracleConfVnic(db)
	if err != nil {
		return
	}
	return
}
