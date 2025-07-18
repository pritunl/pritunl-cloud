package netconf

import (
	"net"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/vm"
)

type NetConf struct {
	Virt            *vm.VirtualMachine
	Vxlan           bool
	VlanId          int
	NetworkMode     string
	NetworkMode6    string
	HostNetwork     bool
	HostNat         bool
	HostSubnet      string
	NodePortNetwork bool
	NodePortSubnet  string
	OracleSubnets   set.Set
	Namespace       string
	DhcpPidPath     string
	Dhcp6PidPath    string
	DhcpLeasePath   string
	Dhcp6LeasePath  string
	VmAdapter       *vm.NetworkAdapter

	PublicAddress  string
	PublicAddress6 string

	OracleVlan           int
	OracleAddress        string
	OracleAddressSubnet  string
	OracleRouterAddress  string
	OracleAddress6       string
	OracleAddressSubnet6 string
	OracleRouterAddress6 string
	OracleMetal          bool

	SpaceBridgeIface       string
	VirtIface              string
	SpaceExternalIface     string
	SpaceExternalIfaceMod  string
	SpaceExternalIfaceMod6 string
	SpaceInternalIface     string
	SpaceHostIface         string
	SpaceNodePortIface     string
	SpaceOracleIface       string
	SpaceOracleVirtIface   string
	SpaceImdsIface         string

	BridgeInternalIface string

	SystemExternalIface string
	SystemInternalIface string
	SystemHostIface     string
	SystemNodePortIface string

	PhysicalExternalIface       string
	PhysicalExternalIfaceBridge bool
	PhysicalInternalIface       string
	PhysicalInternalIfaceBridge bool
	PhysicalHostIface           string
	PhysicalNodePortIface       string

	SpaceExternalIfaceMtu  string
	SystemExternalIfaceMtu string

	SpaceInternalIfaceMtu  string
	BridgeInternalIfaceMtu string
	SystemInternalIfaceMtu string

	SpaceHostIfaceMtu  string
	SystemHostIfaceMtu string
	ImdsIfaceMtu       string

	SpaceNodePortIfaceMtu  string
	SystemNodePortIfaceMtu string

	VirtIfaceMtu string

	InternalAddr            net.IP
	InternalGatewayAddr     net.IP
	InternalGatewayAddrCidr string
	InternalAddr6           net.IP
	InternalGatewayAddr6    net.IP

	ExternalVlan         int
	ExternalAddrCidr     string
	ExternalGatewayAddr  net.IP
	ExternalVlan6        int
	ExternalAddrCidr6    string
	ExternalGatewayAddr6 net.IP

	HostAddr        net.IP
	HostAddrCidr    string
	HostGatewayAddr net.IP

	NodePortAddr     net.IP
	NodePortAddrCidr string

	ExternalMacAddr string
	InternalMacAddr string
	HostMacAddr     string
	NodePortMacAddr string
}

func (n *NetConf) Init(db *database.Database) (err error) {
	err = n.Validate()
	if err != nil {
		return
	}

	err = n.Iface1(db)
	if err != nil {
		return
	}

	err = n.Address(db)
	if err != nil {
		return
	}

	err = n.Iface2(db, false)
	if err != nil {
		return
	}

	err = n.Clear(db)
	if err != nil {
		return
	}

	err = n.Base(db)
	if err != nil {
		return
	}

	err = n.Oracle(db)
	if err != nil {
		return
	}

	err = n.External(db)
	if err != nil {
		return
	}

	err = n.Internal(db)
	if err != nil {
		return
	}

	err = n.Host(db)
	if err != nil {
		return
	}

	err = n.NodePort(db)
	if err != nil {
		return
	}

	err = n.Space(db)
	if err != nil {
		return
	}

	err = n.Vlan(db)
	if err != nil {
		return
	}

	err = n.Bridge(db)
	if err != nil {
		return
	}

	err = n.Imds(db)
	if err != nil {
		return
	}

	err = n.Ip(db)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) Clean(db *database.Database) (err error) {
	err = n.Iface1(db)
	if err != nil {
		return
	}

	err = n.Iface2(db, true)
	if err != nil {
		return
	}

	err = n.ClearAll(db)
	if err != nil {
		return
	}

	return
}
