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
	if n.Virt.CloudVnic != "" {
		vnic, err = oracle.GetVnic(pv, n.Virt.CloudVnic)
		if err != nil {
			if _, ok := err.(*errortypes.NotFoundError); ok {
				logrus.WithFields(logrus.Fields{
					"vnic_id": n.Virt.CloudVnic,
					"error":   err,
				}).Warn("netconf: Cloud vnic not found, creating new vnic")

				err = nil
			} else {
				return
			}
		}

		if vnic == nil {
			found = false
		} else if vnic.SubnetId != n.Virt.CloudSubnet {
			err = oracle.RemoveVnic(pv, n.Virt.CloudVnicAttach)
			if err != nil {
				return
			}

			vnic = nil
		} else if !n.CloudSubnets.Contains(vnic.SubnetId) {
			err = oracle.RemoveVnic(pv, n.Virt.CloudVnicAttach)
			if err != nil {
				return
			}

			vnic = nil
		} else {
			found = true
		}
	}

	if !n.CloudSubnets.Contains(n.Virt.CloudSubnet) {
		err = &errortypes.NotFoundError{
			errors.New("netconf: Invalid cloud subnet"),
		}
		return
	}

	if !found {
		vnicId, vnicAttachId, e := oracle.CreateVnic(
			pv, n.Virt.Id.Hex(), n.Virt.CloudSubnet, !n.Virt.NoPublicAddress,
			!n.Virt.NoPublicAddress6)
		if e != nil {
			err = e
			return
		}

		n.Virt.CloudVnic = vnicId
		n.Virt.CloudVnicAttach = vnicAttachId
		err = n.Virt.CommitCloudVnic(db)
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

	n.CloudMetal = mdata.IsBareMetal()

	if n.CloudMetal {
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
			if vnic.Id == n.Virt.CloudVnic {
				n.Virt.CloudPrivateIp = vnic.PrivateIp
				n.CloudVlan = vnic.VlanTag
				n.CloudAddress = vnic.PrivateIp
				n.CloudAddressSubnet = vnic.SubnetCidrBlock
				n.CloudRouterAddress = vnic.VirtualRouterIp

				if len(vnic.Ipv6Addresses) > 0 {
					n.CloudAddress6 = vnic.Ipv6Addresses[0]
					n.CloudAddressSubnet6 = vnic.Ipv6SubnetCidrBlock
					n.CloudRouterAddress6 = vnic.Ipv6VirtualRouterIp
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

	vnic, err := oracle.GetVnic(pv, n.Virt.CloudVnic)
	if err != nil {
		return
	}

	n.Virt.CloudPublicIp = vnic.PublicIp
	n.Virt.CloudPublicIp6 = vnic.PublicIp6

	err = n.Virt.CommitCloudIps(db)
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
			errors.New("netconf: Failed to find cloud physical interface"),
		}
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link",
		"add", "link", physicalIface,
		"name", n.SpaceCloudVirtIface,
		"address", macAddr,
		"type", "macvlan",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "link",
		"set", "dev", n.SpaceCloudVirtIface,
		"netns", n.Namespace,
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns", "exec", n.Namespace,
		"ip", "link",
		"add", "link", n.SpaceCloudVirtIface,
		"name", n.SpaceCloudIface,
		"type", "vlan",
		"id", strconv.Itoa(n.CloudVlan),
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) oracleConfVnicVirt(db *database.Database) (err error) {
	found := false
	cloudMacAddr := ""

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
			if vnic.Id == n.Virt.CloudVnic {
				n.Virt.CloudPrivateIp = vnic.PrivateIp
				n.CloudAddress = vnic.PrivateIp
				n.CloudAddressSubnet = vnic.SubnetCidrBlock
				n.CloudRouterAddress = vnic.VirtualRouterIp

				if len(vnic.Ipv6Addresses) > 0 {
					n.CloudAddress6 = vnic.Ipv6Addresses[0]
					n.CloudAddressSubnet6 = vnic.Ipv6SubnetCidrBlock
					n.CloudRouterAddress6 = vnic.Ipv6VirtualRouterIp
				}

				cloudMacAddr = strings.ToLower(vnic.MacAddr)

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

	vnic, err := oracle.GetVnic(pv, n.Virt.CloudVnic)
	if err != nil {
		return
	}

	n.Virt.CloudPublicIp = vnic.PublicIp
	n.Virt.CloudPublicIp6 = vnic.PublicIp6

	err = n.Virt.CommitCloudIps(db)
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

	cloudIface := ""
	for _, iface := range ifaces {
		if strings.ToLower(iface.HardwareAddr.String()) == cloudMacAddr {
			cloudIface = iface.Name
			break
		}
	}

	if cloudIface == "" {
		err = &errortypes.NotFoundError{
			errors.New("netconf: Failed to find cloud interface"),
		}
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link",
		"set", "dev", cloudIface, "down",
	)
	if err != nil {
		return
	}

	if cloudIface != n.SpaceCloudIface {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", cloudIface,
			"name", n.SpaceCloudIface,
		)
		if err != nil {
			return
		}
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "link",
		"set", "dev", n.SpaceCloudIface,
		"netns", n.Namespace,
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) oracleMtu(db *database.Database) (err error) {
	if n.CloudMetal {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"ip", "link",
			"set", "dev", n.SpaceCloudVirtIface,
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
		"set", "dev", n.SpaceCloudIface,
		"mtu", "9000",
	)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) oracleIp(db *database.Database) (err error) {
	subnetSplit := strings.Split(n.CloudAddressSubnet, "/")
	if len(subnetSplit) != 2 {
		err = &errortypes.ParseError{
			errors.Newf("netconf: Failed to get cloud cidr %s",
				n.CloudAddressSubnet),
		}
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns", "exec", n.Namespace,
		"ip", "addr",
		"add", n.CloudAddress+"/"+subnetSplit[1],
		"dev", n.SpaceCloudIface,
	)
	if err != nil {
		return
	}

	if n.CloudAddress6 != "" {
		subnetSplit6 := strings.Split(n.CloudAddressSubnet6, "/")
		if len(subnetSplit6) != 2 {
			err = &errortypes.ParseError{
				errors.Newf("netconf: Failed to get cloud cidr6 %s",
					n.CloudAddressSubnet6),
			}
			return
		}

		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "netns", "exec", n.Namespace,
			"ip", "addr",
			"add", n.CloudAddress6+"/"+subnetSplit6[1],
			"dev", n.SpaceCloudIface,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) oracleUp(db *database.Database) (err error) {
	if n.CloudMetal {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"ip", "link",
			"set", "dev", n.SpaceCloudVirtIface, "up",
		)
		if err != nil {
			return
		}
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", n.Namespace,
		"ip", "link",
		"set", "dev", n.SpaceCloudIface, "up",
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
		"via", n.CloudRouterAddress,
		"dev", n.SpaceCloudIface,
	)
	if err != nil {
		return
	}

	if n.CloudAddress6 != "" && n.CloudRouterAddress6 != "" {
		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "netns", "exec", n.Namespace,
			"ip", "-6", "route",
			"add", "default",
			"via", n.CloudRouterAddress6,
			"dev", n.SpaceCloudIface,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) Oracle(db *database.Database) (err error) {
	if n.NetworkMode != node.Cloud || n.Virt.CloudSubnet == "" {
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
