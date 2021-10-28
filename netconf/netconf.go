package netconf

import (
	"net"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/vm"
)

type NetConf struct {
	Virt          *vm.VirtualMachine
	Vxlan         bool
	VlanId        int
	NetworkMode   string
	NetworkMode6  string
	HostNetwork   bool
	HostBlock     primitive.ObjectID
	OracleSubnets set.Set
	JumboFrames   bool
	Namespace     string
	DhcpPidPath   string
	DhcpLeasePath string
	VmAdapter     *vm.NetworkAdapter

	PublicAddress  string
	PublicAddress6 string

	OracleAddress       string
	OracleAddressSubnet string
	OracleRouterAddress string

	VirtIface           string
	SpaceExternalIface  string
	SpaceExternalIface6 string
	SpaceInternalIface  string
	SpaceHostIface      string

	BridgeInternalIface string

	SystemExternalIface  string
	SystemExternalIface6 string
	SystemInternalIface  string
	SystemHostIface      string

	PhysicalExternalIface        string
	PhysicalExternalIfaceBridge  bool
	PhysicalExternalIface6       string
	PhysicalExternalIfaceBridge6 bool
	PhysicalInternalIface        string
	PhysicalInternalIfaceBridge  bool
	PhysicalHostIface            string

	SpaceExternalIfaceMtu   string
	SpaceExternalIfaceMtu6  string
	SystemExternalIfaceMtu  string
	SystemExternalIfaceMtu6 string

	SpaceInternalIfaceMtu  string
	BridgeInternalIfaceMtu string
	SystemInternalIfaceMtu string

	SpaceHostIfaceMtu  string
	SystemHostIfaceMtu string

	VirtIfaceMtu string

	InternalAddr            net.IP
	InternalGatewayAddr     net.IP
	InternalGatewayAddrCidr string
	InternalAddr6           net.IP
	InternalGatewayAddr6    net.IP

	ExternalAddrCidr     string
	ExternalGatewayAddr  net.IP
	ExternalAddrCidr6    string
	ExternalGatewayAddr6 net.IP

	HostAddr        net.IP
	HostAddrCidr    string
	HostGatewayAddr net.IP

	ExternalMacAddr  string
	ExternalMacAddr6 string
	InternalMacAddr  string
	HostMacAddr      string
}

func (n *NetConf) Init(db *database.Database) (err error) {
	err = n.Validate()
	if err != nil {
		return
	}

	err = n.Iface(db)
	if err != nil {
		return
	}

	err = n.Address(db)
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

	err = n.Ip(db)
	if err != nil {
		return
	}

	return
}

func (n *NetConf) Clean(db *database.Database) (err error) {
	err = n.Iface(db)
	if err != nil {
		return
	}

	err = n.ClearAll(db)
	if err != nil {
		return
	}

	err = n.Clear(db)
	if err != nil {
		return
	}

	return
}
