package netconf

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/oracle"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
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
			if _, ok := err.(*errortypes.NotFoundError); ok {
				logrus.WithFields(logrus.Fields{
					"vnic_id": n.Virt.OracleVnic,
					"error":   err,
				}).Warn("netconf: Oracle vnic not found, creating new vnic")

				err = nil
			} else {
				return
			}
		}

		if vnic == nil {
			found = false
		} else if vnic.SubnetId != n.Virt.OracleSubnet {
			err = oracle.RemoveVnic(pv, n.Virt.OracleVnicAttach)
			if err != nil {
				return
			}

			vnic = nil
		} else if !n.OracleSubnets.Contains(vnic.SubnetId) {
			err = oracle.RemoveVnic(pv, n.Virt.OracleVnicAttach)
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
			pv, n.Virt.Id.Hex(), n.Virt.OracleSubnet, !n.Virt.NoPublicAddress,
			!n.Virt.NoPublicAddress6)
		if e != nil {
			err = e
			return
		}

		n.Virt.OracleVnic = vnicId
		n.Virt.OracleVnicAttach = vnicAttachId
		err = n.Virt.CommitOracleVnic(db)
		if err != nil {
			_ = oracle.RemoveVnic(pv, vnicAttachId)
			return
		}
	}

	return
}

func (n *NetConf) oracleConfVnic(db *database.Database) (err error) {
	mdata, e := oracle.GetOciMetadata()
	if e != nil {
		err = e
		return
	}

	n.OracleMetal = mdata.IsBareMetal()

	if n.OracleMetal {
		err = n.oracleConfVnicMetal(db)
		if err != nil {
			return
		}
	} else {
		err = n.oracleConfVnicVirt(db)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) oracleConfVnicMetal(db *database.Database) (err error) {
	found := false
	nicIndex := 0
	macAddr := ""
	physicalMacAddr := ""

	pv, err := oracle.NewProvider(node.Self.GetOracleAuthProvider())
	if err != nil {
		return
	}

	for i := 0; i < 120; i++ {
		time.Sleep(2 * time.Second)

		mdata, e := oracle.GetOciMetadata()
		if e != nil {
			err = e
			return
		}

		for _, vnic := range mdata.Vnics {
			if vnic.Id == n.Virt.OracleVnic {
				n.Virt.OraclePrivateIp = vnic.PrivateIp
				n.OracleVlan = vnic.VlanTag
				n.OracleAddress = vnic.PrivateIp
				n.OracleAddressSubnet = vnic.SubnetCidrBlock
				n.OracleRouterAddress = vnic.VirtualRouterIp

				if len(vnic.Ipv6Addresses) > 0 {
					n.OracleAddress6 = vnic.Ipv6Addresses[0]
					n.OracleAddressSubnet6 = vnic.Ipv6SubnetCidrBlock
					n.OracleRouterAddress6 = vnic.Ipv6VirtualRouterIp
				}

				nicIndex = vnic.NicIndex
				macAddr = strings.ToLower(vnic.MacAddr)

				found = true
				break
			}
		}

		if found {
			break
		}
	}

	if !found {
		err = &errortypes.NotFoundError{
			errors.New("netconf: Failed to find vnic"),
		}
		return
	}

	mdata, err := oracle.GetOciMetadata()
	if err != nil {
		return
	}

	found = false
	for _, vnic := range mdata.Vnics {
		if vnic.NicIndex == nicIndex && vnic.VlanTag == 0 {
			physicalMacAddr = strings.ToLower(vnic.MacAddr)

			found = true
			break
		}
	}

	if !found {
		err = &errortypes.NotFoundError{
			errors.New("netconf: Failed to find physical nic"),
		}
		return
	}

	vnic, err := oracle.GetVnic(pv, n.Virt.OracleVnic)
	if err != nil {
		return
	}

	n.Virt.OraclePublicIp = vnic.PublicIp
	n.Virt.OraclePublicIp6 = vnic.PublicIp6

	err = n.Virt.CommitOracleIps(db)
	if err != nil {
		return
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "netconf: Failed get network interfaces"),
		}
		return
	}

	physicalIface := ""
	for _, iface := range ifaces {
		if strings.ToLower(iface.HardwareAddr.String()) == physicalMacAddr {
			physicalIface = iface.Name
			break
		}
	}

	if physicalIface == "" {
		err = &errortypes.NotFoundError{
			errors.New("netconf: Failed to find oracle physical interface"),
		}
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link",
		"add", "link", physicalIface,
		"name", n.SpaceOracleVirtIface,
		"address", macAddr,
		"type", "macvlan",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "link",
		"set", "dev", n.SpaceOracleVirtIface,
		"netns", n.Namespace,
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns", "exec", n.Namespace,
		"ip", "link",
		"add", "link", n.SpaceOracleVirtIface,
		"name", n.SpaceOracleIface,
		"type", "vlan",
		"id", strconv.Itoa(n.OracleVlan),
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) oracleConfVnicVirt(db *database.Database) (err error) {
	found := false
	oracleMacAddr := ""

	pv, err := oracle.NewProvider(node.Self.GetOracleAuthProvider())
	if err != nil {
		return
	}

	for i := 0; i < 120; i++ {
		time.Sleep(2 * time.Second)

		mdata, e := oracle.GetOciMetadata()
		if e != nil {
			err = e
			return
		}

		for _, vnic := range mdata.Vnics {
			if vnic.Id == n.Virt.OracleVnic {
				n.Virt.OraclePrivateIp = vnic.PrivateIp
				n.OracleAddress = vnic.PrivateIp
				n.OracleAddressSubnet = vnic.SubnetCidrBlock
				n.OracleRouterAddress = vnic.VirtualRouterIp

				if len(vnic.Ipv6Addresses) > 0 {
					n.OracleAddress6 = vnic.Ipv6Addresses[0]
					n.OracleAddressSubnet6 = vnic.Ipv6SubnetCidrBlock
					n.OracleRouterAddress6 = vnic.Ipv6VirtualRouterIp
				}

				oracleMacAddr = strings.ToLower(vnic.MacAddr)

				found = true
				break
			}
		}

		if found {
			break
		}
	}

	if !found {
		err = &errortypes.NotFoundError{
			errors.New("netconf: Failed to find vnic"),
		}
		return
	}

	vnic, err := oracle.GetVnic(pv, n.Virt.OracleVnic)
	if err != nil {
		return
	}

	n.Virt.OraclePublicIp = vnic.PublicIp
	n.Virt.OraclePublicIp6 = vnic.PublicIp6

	err = n.Virt.CommitOracleIps(db)
	if err != nil {
		return
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "netconf: Failed get network interfaces"),
		}
		return
	}

	oracleIface := ""
	for _, iface := range ifaces {
		if strings.ToLower(iface.HardwareAddr.String()) == oracleMacAddr {
			oracleIface = iface.Name
			break
		}
	}

	if oracleIface == "" {
		err = &errortypes.NotFoundError{
			errors.New("netconf: Failed to find oracle interface"),
		}
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link",
		"set", "dev", oracleIface, "down",
	)
	if err != nil {
		return
	}

	if oracleIface != n.SpaceOracleIface {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", oracleIface,
			"name", n.SpaceOracleIface,
		)
		if err != nil {
			return
		}
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "link",
		"set", "dev", n.SpaceOracleIface,
		"netns", n.Namespace,
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) oracleMtu(db *database.Database) (err error) {
	if n.OracleMetal {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"ip", "link",
			"set", "dev", n.SpaceOracleVirtIface,
			"mtu", "9000",
		)
		if err != nil {
			return
		}
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"ip", "link",
		"set", "dev", n.SpaceOracleIface,
		"mtu", "9000",
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) oracleIp(db *database.Database) (err error) {
	subnetSplit := strings.Split(n.OracleAddressSubnet, "/")
	if len(subnetSplit) != 2 {
		err = &errortypes.ParseError{
			errors.Newf("netconf: Failed to get oracle cidr %s",
				n.OracleAddressSubnet),
		}
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns", "exec", n.Namespace,
		"ip", "addr",
		"add", n.OracleAddress+"/"+subnetSplit[1],
		"dev", n.SpaceOracleIface,
	)
	if err != nil {
		return
	}

	if n.OracleAddress6 != "" {
		subnetSplit6 := strings.Split(n.OracleAddressSubnet6, "/")
		if len(subnetSplit6) != 2 {
			err = &errortypes.ParseError{
				errors.Newf("netconf: Failed to get oracle cidr6 %s",
					n.OracleAddressSubnet6),
			}
			return
		}

		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "netns", "exec", n.Namespace,
			"ip", "addr",
			"add", n.OracleAddress6+"/"+subnetSplit6[1],
			"dev", n.SpaceOracleIface,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) oracleUp(db *database.Database) (err error) {
	if n.OracleMetal {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"ip", "link",
			"set", "dev", n.SpaceOracleVirtIface, "up",
		)
		if err != nil {
			return
		}
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"ip", "link",
		"set", "dev", n.SpaceOracleIface, "up",
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) oracleRoute(db *database.Database) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns", "exec", n.Namespace,
		"ip", "route",
		"add", "default",
		"via", n.OracleRouterAddress,
		"dev", n.SpaceOracleIface,
	)
	if err != nil {
		return
	}

	if n.OracleAddress6 != "" && n.OracleRouterAddress6 != "" {
		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "netns", "exec", n.Namespace,
			"ip", "-6", "route",
			"add", "default",
			"via", n.OracleRouterAddress6,
			"dev", n.SpaceOracleIface,
		)
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

	err = n.oracleMtu(db)
	if err != nil {
		return
	}

	err = n.oracleIp(db)
	if err != nil {
		return
	}

	err = n.oracleUp(db)
	if err != nil {
		return
	}

	err = n.oracleRoute(db)
	if err != nil {
		return
	}

	return
}
