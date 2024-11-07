package netconf

import (
	"net"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/vm"
)

type NetConf struct {
	Virt                *vm.VirtualMachine
	Vxlan               bool
	VlanId              int
	NetworkMode         string
	NetworkMode6        string
	HostNetwork         bool
	HostNat             bool
	HostSubnet          string
	HostBlock           primitive.ObjectID
	OracleSubnets       set.Set
	JumboFramesExternal bool
	JumboFramesInternal bool
	Namespace           string
	DhcpPidPath         string
	DhcpLeasePath       string
	VmAdapter           *vm.NetworkAdapter

	PublicAddress  string
	PublicAddress6 string

	OracleVlan          int
	OracleAddress       string
	OracleAddressSubnet string
	OracleRouterAddress string
	OracleMetal         bool

	VirtIface            string
	SpaceExternalIface   string
	SpaceInternalIface   string
	SpaceHostIface       string
	SpaceOracleIface     string
	SpaceOracleVirtIface string
	SpaceImdsIface       string

	BridgeInternalIface string

	SystemExternalIface string
	SystemInternalIface string
	SystemHostIface     string

	PhysicalExternalIface       string
	PhysicalExternalIfaceBridge bool
	PhysicalInternalIface       string
	PhysicalInternalIfaceBridge bool
	PhysicalHostIface           string

	SpaceExternalIfaceMtu  string
	SystemExternalIfaceMtu string

	SpaceInternalIfaceMtu  string
	BridgeInternalIfaceMtu string
	SystemInternalIfaceMtu string

	SpaceHostIfaceMtu  string
	SystemHostIfaceMtu string
	ImdsIfaceMtu       string

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

	ExternalMacAddr string
	InternalMacAddr string
	HostMacAddr     string
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

	err = n.Clear(db)
	if err != nil {
		return
	}

	return
}
