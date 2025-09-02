package demo

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/audit"
	"github.com/pritunl/pritunl-cloud/balancer"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/cloud"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/drive"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/ip"
	"github.com/pritunl/pritunl-cloud/log"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/organization"
	"github.com/pritunl/pritunl-cloud/pci"
	"github.com/pritunl/pritunl-cloud/plan"
	"github.com/pritunl/pritunl-cloud/policy"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/session"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/pritunl/pritunl-cloud/subscription"
	"github.com/pritunl/pritunl-cloud/usb"
	"github.com/pritunl/pritunl-cloud/user"
	"github.com/pritunl/pritunl-cloud/useragent"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
)

func IsDemo() bool {
	return settings.System.Demo
}

func Blocked(c *gin.Context) bool {
	if !IsDemo() {
		return false
	}

	errData := &errortypes.ErrorData{
		Error:   "demo_unavailable",
		Message: "Not available in demo mode",
	}
	c.JSON(400, errData)

	return true
}

func BlockedSilent(c *gin.Context) bool {
	if !IsDemo() {
		return false
	}

	c.JSON(200, nil)
	return true
}

// Users
var Users = []*user.User{
	&user.User{
		Id:            utils.ObjectIdHex("5b6cd11857e4a9a88cbf072e"),
		Type:          "local",
		Provider:      primitive.ObjectID{},
		Username:      "demo",
		Token:         "",
		Secret:        "",
		LastActive:    time.Now(),
		LastSync:      time.Now(),
		Roles:         []string{"demo"},
		Administrator: "super",
		Disabled:      false,
		ActiveUntil:   time.Time{},
		Permissions:   []string{},
	},
	&user.User{
		Id:            utils.ObjectIdHex("5a7542190accad1a8a53b568"),
		Type:          "local",
		Provider:      primitive.ObjectID{},
		Username:      "pritunl",
		Token:         "",
		Secret:        "",
		LastActive:    time.Time{},
		LastSync:      time.Time{},
		Roles:         []string{},
		Administrator: "super",
		Disabled:      false,
		ActiveUntil:   time.Time{},
		Permissions:   []string{},
	},
}

var Agent = &useragent.Agent{
	OperatingSystem: useragent.Linux,
	Browser:         useragent.Chrome,
	Ip:              "8.8.8.8",
	Isp:             "Google",
	Continent:       "North America",
	ContinentCode:   "NA",
	Country:         "United States",
	CountryCode:     "US",
	Region:          "Washington",
	RegionCode:      "WA",
	City:            "Seattle",
	Latitude:        47.611,
	Longitude:       -122.337,
}

var Audits = []*audit.Audit{
	&audit.Audit{
		Id:        utils.ObjectIdHex("5a17f9bf051a45ffacf2b352"),
		Timestamp: time.Unix(1498018860, 0),
		Type:      "admin_login",
		Fields: audit.Fields{
			"method": "local",
		},
		Agent: Agent,
	},
}

var Sessions = []*session.Session{
	&session.Session{
		Id:         "jhgRu4n3oY0iXRYmLb77Ql5jNs2o7uWM",
		Type:       session.User,
		Timestamp:  time.Unix(1498018860, 0),
		LastActive: time.Unix(1498018860, 0),
		Removed:    false,
		Agent:      Agent,
	},
}

var Logs = []*log.Entry{
	&log.Entry{
		Id:        utils.ObjectIdHex("5a18e6ae051a45ffac0e5b67"),
		Level:     log.Info,
		Timestamp: time.Unix(1498018860, 0),
		Message:   "router: Starting redirect server",
		Stack:     "",
		Fields: map[string]interface{}{
			"port":       80,
			"production": true,
			"protocol":   "http",
		},
	},
	&log.Entry{
		Id:        utils.ObjectIdHex("5a190b42051a45ffac129bbc"),
		Level:     log.Info,
		Timestamp: time.Unix(1498018860, 0),
		Message:   "router: Starting web server",
		Stack:     "",
		Fields: map[string]interface{}{
			"port":       443,
			"production": true,
			"protocol":   "https",
		},
	},
}

var Subscription = &subscription.Subscription{
	Active:            true,
	Status:            "active",
	Plan:              "zero",
	Quantity:          1,
	Amount:            5000,
	PeriodEnd:         time.Unix(1893499200, 0),
	TrialEnd:          time.Time{},
	CancelAtPeriodEnd: false,
	Balance:           0,
	UrlKey:            "demo",
}

// Nodes
var Nodes = []*node.Node{
	{
		Id:                   utils.ObjectIdHex("689733b2a7a35eae0dbaea09"),
		Datacenter:           utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Zone:                 utils.ObjectIdHex("689733b7a7a35eae0dbaea1e"),
		Name:                 "pritunl-east0",
		Comment:              "",
		Types:                []string{"admin", "hypervisor"},
		Timestamp:            time.Now(),
		Port:                 443,
		NoRedirectServer:     false,
		Protocol:             "https",
		Hypervisor:           "kvm",
		Vga:                  "virtio",
		VgaRender:            "",
		AvailableRenders:     []string{},
		Gui:                  false,
		GuiUser:              "",
		GuiMode:              "",
		Certificates:         []primitive.ObjectID{},
		AdminDomain:          "",
		UserDomain:           "",
		WebauthnDomain:       "",
		RequestsMin:          23,
		ForwardedForHeader:   "",
		ForwardedProtoHeader: "",
		ExternalInterfaces:   []string{},
		ExternalInterfaces6:  []string{},
		InternalInterfaces:   []string{"bond0.2"},
		AvailableInterfaces: []ip.Interface{
			{Name: "bond0", Address: ""},
			{Name: "bond0.2", Address: "10.8.0.10"},
			{Name: "bond0.4", Address: "10.253.67.90"},
			{Name: "bonding_masters", Address: ""},
			{Name: "enp1s0f0", Address: ""},
			{Name: "enp1s0f1", Address: ""},
			{Name: "veth0", Address: ""},
		},
		AvailableBridges: []ip.Interface{
			{Name: "pritunlhost0", Address: "198.18.84.1"},
			{Name: "pritunlport0", Address: "198.19.96.1"},
		},
		AvailableVpcs:    []*cloud.Vpc{},
		CloudSubnets:     []string{},
		DefaultInterface: "bond0.4",
		NetworkMode:      "static",
		NetworkMode6:     "static",
		Blocks: []*node.BlockAttachment{
			{
				Interface: "bond0.4",
				Block:     utils.ObjectIdHex("689733b7a7a35eae0dbaea2f"),
			},
		},
		Blocks6: []*node.BlockAttachment{
			{
				Interface: "bond0.4",
				Block:     utils.ObjectIdHex("68973a47b5844593cf99cc7a"),
			},
		},
		Pools:  []primitive.ObjectID{},
		Shares: []*node.Share{},
		AvailableDrives: []*drive.Device{
			{Id: "nvme-INTEL_27Z1P0FGN"},
			{Id: "nvme-INTEL_27Z1P0FGN-part1"},
			{Id: "nvme-INTEL_27Z1P0FGN-part2"},
			{Id: "nvme-INTEL_27Z1P0FGN-part3"},
			{Id: "nvme-INTEL_42K1P0FGN"},
			{Id: "nvme-INTEL_42K1P0FGN-part1"},
			{Id: "nvme-INTEL_42K1P0FGN-part2"},
			{Id: "nvme-INTEL_42K1P0FGN-part3"},
		},
		InstanceDrives:          []*drive.Device{},
		NoHostNetwork:           false,
		NoNodePortNetwork:       false,
		HostNat:                 true,
		DefaultNoPublicAddress:  true,
		DefaultNoPublicAddress6: false,
		JumboFrames:             true,
		JumboFramesInternal:     true,
		Iscsi:                   false,
		LocalIsos:               nil,
		UsbPassthrough:          false,
		UsbDevices:              []*usb.Device{},
		PciPassthrough:          false,
		PciDevices:              []*pci.Device{},
		Hugepages:               false,
		HugepagesSize:           0,
		Firewall:                false,
		Roles:                   []string{"shape-m2"},
		Memory:                  41.02,
		HugePagesUsed:           0,
		Load1:                   43.17,
		Load5:                   43.33,
		Load15:                  43.83,
		CpuUnits:                128,
		MemoryUnits:             512,
		CpuUnitsRes:             68,
		MemoryUnitsRes:          198,
		PublicIps:               []string{"10.253.67.90"},
		PublicIps6: []string{
			"2001:db8:85a3:4d2f:1319:8a2e:370:7348",
		},
		PrivateIps: map[string]string{
			"bond0.2": "10.8.0.10",
		},
		SoftwareVersion: constants.Version,
		Hostname:        "pritunl-east0",
		VirtPath:        "/var/lib/pritunl-cloud",
		CachePath:       "/var/cache/pritunl-cloud",
		TempPath:        "",
		OracleUser:      "",
		OracleTenancy:   "",
		OraclePublicKey: `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxWtYOIzsHsLlBI1jeepJ
q8dyR1JH3QLdAJ2IFGZDtHCCi46Lvmx7hC8bAutj5s37qOfBrom6UOJf0f9zEP8K
y8qTb2S4XOAWBHuGpaBqFEhtpW+vIxiy26vdZN85P3xzYle0uodr86+y2bVHMHKB
0oEHnqu+CmH/r4GedBVFVBASo9C5iILsyISf4oep390V/u23RAXXNfcKvUYR4c2u
fZBwlSVEDrK+X21ocJc+8VGbbLhXBvMEdqXzs1bbFzFHow8TjduxDNTbntIRpo6W
0O7xMahUHxDWDro5fAkzvpk6wUBM6yWXXgwkDLLHW50dUnqgFJgOTIHXEtPSt4eU
2wIDAQAB
-----END PUBLIC KEY-----`,
		Operation: "",
	},
	{
		Id:                   utils.ObjectIdHex("689733b2a7a35eae0dbaea09"),
		Datacenter:           utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Zone:                 utils.ObjectIdHex("689733b7a7a35eae0dbaea1e"),
		Name:                 "pritunl-east1",
		Comment:              "",
		Types:                []string{"admin", "hypervisor"},
		Timestamp:            time.Now(),
		Port:                 443,
		NoRedirectServer:     false,
		Protocol:             "https",
		Hypervisor:           "kvm",
		Vga:                  "virtio",
		VgaRender:            "",
		AvailableRenders:     []string{},
		Gui:                  false,
		GuiUser:              "",
		GuiMode:              "",
		Certificates:         []primitive.ObjectID{},
		AdminDomain:          "",
		UserDomain:           "",
		WebauthnDomain:       "",
		RequestsMin:          44,
		ForwardedForHeader:   "",
		ForwardedProtoHeader: "",
		ExternalInterfaces:   []string{},
		ExternalInterfaces6:  []string{},
		InternalInterfaces:   []string{"bond0.2"},
		AvailableInterfaces: []ip.Interface{
			{Name: "bond0", Address: ""},
			{Name: "bond0.2", Address: "10.8.0.11"},
			{Name: "bond0.4", Address: "10.253.67.91"},
			{Name: "bonding_masters", Address: ""},
			{Name: "enp1s0f0", Address: ""},
			{Name: "enp1s0f1", Address: ""},
			{Name: "veth0", Address: ""},
		},
		AvailableBridges: []ip.Interface{
			{Name: "pritunlhost0", Address: "198.18.84.1"},
			{Name: "pritunlport0", Address: "198.19.96.1"},
		},
		AvailableVpcs:    []*cloud.Vpc{},
		CloudSubnets:     []string{},
		DefaultInterface: "bond0.4",
		NetworkMode:      "static",
		NetworkMode6:     "static",
		Blocks: []*node.BlockAttachment{
			{
				Interface: "bond0.4",
				Block:     utils.ObjectIdHex("689733b7a7a35eae0dbaea2f"),
			},
		},
		Blocks6: []*node.BlockAttachment{
			{
				Interface: "bond0.4",
				Block:     utils.ObjectIdHex("68973a47b5844593cf99cc7a"),
			},
		},
		Pools:  []primitive.ObjectID{},
		Shares: []*node.Share{},
		AvailableDrives: []*drive.Device{
			{Id: "nvme-INTEL_27Z1P0FGN"},
			{Id: "nvme-INTEL_27Z1P0FGN-part1"},
			{Id: "nvme-INTEL_27Z1P0FGN-part2"},
			{Id: "nvme-INTEL_27Z1P0FGN-part3"},
			{Id: "nvme-INTEL_42K1P0FGN"},
			{Id: "nvme-INTEL_42K1P0FGN-part1"},
			{Id: "nvme-INTEL_42K1P0FGN-part2"},
			{Id: "nvme-INTEL_42K1P0FGN-part3"},
		},
		InstanceDrives:          []*drive.Device{},
		NoHostNetwork:           false,
		NoNodePortNetwork:       false,
		HostNat:                 true,
		DefaultNoPublicAddress:  true,
		DefaultNoPublicAddress6: false,
		JumboFrames:             true,
		JumboFramesInternal:     true,
		Iscsi:                   false,
		LocalIsos:               nil,
		UsbPassthrough:          false,
		UsbDevices:              []*usb.Device{},
		PciPassthrough:          false,
		PciDevices:              []*pci.Device{},
		Hugepages:               false,
		HugepagesSize:           0,
		Firewall:                false,
		Roles:                   []string{"shape-m2"},
		Memory:                  62.02,
		HugePagesUsed:           0,
		Load1:                   67.12,
		Load5:                   57.43,
		Load15:                  55.23,
		CpuUnits:                128,
		MemoryUnits:             512,
		CpuUnitsRes:             78,
		MemoryUnitsRes:          324,
		PublicIps:               []string{"10.253.67.91"},
		PublicIps6: []string{
			"2001:db8:85a3:4d2f:7c3a:9f42:8e5b:a234",
		},
		PrivateIps: map[string]string{
			"bond0.2": "10.8.0.11",
		},
		SoftwareVersion: constants.Version,
		Hostname:        "pritunl-east0",
		VirtPath:        "/var/lib/pritunl-cloud",
		CachePath:       "/var/cache/pritunl-cloud",
		TempPath:        "",
		OracleUser:      "",
		OracleTenancy:   "",
		OraclePublicKey: `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAz4K8Lm3QvR7WxN5YdE2P
jX9TpQ6HgM1wV0nS4KaF3ZcB8LrY5UvO2JmN7XsPqI1AgK8EoH3RdWzM9LfY2VtN
kP4QxGsJ7YnR8LwVmT3AqZ5HvK2NdP1XoS8JgR4LmW7YxQ3VnH5TsK9PpL2MdX8Rg
vJ3KqN5WxT1LsM4HgY7RdP8NqV2JmK5XwL3TsR8YgN4HxP1LdK9VwQ2MsT3XpR7Y
nL8KgJ5WdH3TmR9XsL2PqN7VxK4MgT3HdJ8YwP2LsK5RxT1NqM4JgY7PxR8WsL3T
mK9XwN2HgJ5YdL3RsP8VqT2MxK4NhR3JdY8WwL2TsM5QxN1PqK4YgJ7RxP8VsT3M
PwIDAQAB
-----END PUBLIC KEY-----`,
		Operation: "",
	},
	{
		Id:                   utils.ObjectIdHex("689733b2a7a35eae0dbaea09"),
		Datacenter:           utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Zone:                 utils.ObjectIdHex("689733b7a7a35eae0dbaea1e"),
		Name:                 "pritunl-east2",
		Comment:              "",
		Types:                []string{"admin", "hypervisor"},
		Timestamp:            time.Now(),
		Port:                 443,
		NoRedirectServer:     false,
		Protocol:             "https",
		Hypervisor:           "kvm",
		Vga:                  "virtio",
		VgaRender:            "",
		AvailableRenders:     []string{},
		Gui:                  false,
		GuiUser:              "",
		GuiMode:              "",
		Certificates:         []primitive.ObjectID{},
		AdminDomain:          "",
		UserDomain:           "",
		WebauthnDomain:       "",
		RequestsMin:          26,
		ForwardedForHeader:   "",
		ForwardedProtoHeader: "",
		ExternalInterfaces:   []string{},
		ExternalInterfaces6:  []string{},
		InternalInterfaces:   []string{"bond0.2"},
		AvailableInterfaces: []ip.Interface{
			{Name: "bond0", Address: ""},
			{Name: "bond0.2", Address: "10.8.0.12"},
			{Name: "bond0.4", Address: "10.253.67.92"},
			{Name: "bonding_masters", Address: ""},
			{Name: "enp1s0f0", Address: ""},
			{Name: "enp1s0f1", Address: ""},
			{Name: "veth0", Address: ""},
		},
		AvailableBridges: []ip.Interface{
			{Name: "pritunlhost0", Address: "198.18.84.1"},
			{Name: "pritunlport0", Address: "198.19.96.1"},
		},
		AvailableVpcs:    []*cloud.Vpc{},
		CloudSubnets:     []string{},
		DefaultInterface: "bond0.4",
		NetworkMode:      "static",
		NetworkMode6:     "static",
		Blocks: []*node.BlockAttachment{
			{
				Interface: "bond0.4",
				Block:     utils.ObjectIdHex("689733b7a7a35eae0dbaea2f"),
			},
		},
		Blocks6: []*node.BlockAttachment{
			{
				Interface: "bond0.4",
				Block:     utils.ObjectIdHex("68973a47b5844593cf99cc7a"),
			},
		},
		Pools:  []primitive.ObjectID{},
		Shares: []*node.Share{},
		AvailableDrives: []*drive.Device{
			{Id: "nvme-INTEL_27Z1P0FGN"},
			{Id: "nvme-INTEL_27Z1P0FGN-part1"},
			{Id: "nvme-INTEL_27Z1P0FGN-part2"},
			{Id: "nvme-INTEL_27Z1P0FGN-part3"},
			{Id: "nvme-INTEL_42K1P0FGN"},
			{Id: "nvme-INTEL_42K1P0FGN-part1"},
			{Id: "nvme-INTEL_42K1P0FGN-part2"},
			{Id: "nvme-INTEL_42K1P0FGN-part3"},
		},
		InstanceDrives:          []*drive.Device{},
		NoHostNetwork:           false,
		NoNodePortNetwork:       false,
		HostNat:                 true,
		DefaultNoPublicAddress:  true,
		DefaultNoPublicAddress6: false,
		JumboFrames:             true,
		JumboFramesInternal:     true,
		Iscsi:                   false,
		LocalIsos:               nil,
		UsbPassthrough:          false,
		UsbDevices:              []*usb.Device{},
		PciPassthrough:          false,
		PciDevices:              []*pci.Device{},
		Hugepages:               false,
		HugepagesSize:           0,
		Firewall:                false,
		Roles:                   []string{"shape-m2"},
		Memory:                  41.02,
		HugePagesUsed:           0,
		Load1:                   76.53,
		Load5:                   67.34,
		Load15:                  67.87,
		CpuUnits:                128,
		MemoryUnits:             512,
		CpuUnitsRes:             56,
		MemoryUnitsRes:          286,
		PublicIps:               []string{"10.253.67.92"},
		PublicIps6: []string{
			"2001:db8:85a3:4d2f:2e91:5cd8:f1a6:7b89",
		},
		PrivateIps: map[string]string{
			"bond0.2": "10.8.0.12",
		},
		SoftwareVersion: constants.Version,
		Hostname:        "pritunl-east0",
		VirtPath:        "/var/lib/pritunl-cloud",
		CachePath:       "/var/cache/pritunl-cloud",
		TempPath:        "",
		OracleUser:      "",
		OracleTenancy:   "",
		OraclePublicKey: `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA3Fk7Tp4NwV8XmQ6LdJ9R
hY2BqG8Vz0kL5HgM7Xa4FrC9NsU1YvQ3KnR8WxL6ZqE2DfH5JtX3YmVgB8KsNp4Q
rL9XvT2HdQ8YwK5JmF3NxR7Pq4VsL8TgW6ZnM1KfY9DpX2RvQ8JwB3GtL5YxH4Nm
sK7VpX3LdR9FwT8MqN2HxY5BgJ6TsL4RpW8NvK3YxH5QmD9LwT2JsN8KgR4XpM7V
fY3BwL8TmN5KxQ2JgH7RsP9VwM4LnT8KpY3XgN5BwJ2HsR8TmL4VxQ9JpN3KgY7B
wR5TnM8LpJ2XgH4VsN9BwK3TmQ5RxL8JpY7NwH2VsR4KgM9TnL5BxJ3QwN8VpH7Y
KwIDAQAB
-----END PUBLIC KEY-----`,
		Operation: "",
	},
	{
		Id:                   utils.ObjectIdHex("689733b2a7a35eae0dbaea09"),
		Datacenter:           utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Zone:                 utils.ObjectIdHex("689733b7a7a35eae0dbaea1e"),
		Name:                 "pritunl-east3",
		Comment:              "",
		Types:                []string{"admin", "hypervisor"},
		Timestamp:            time.Now(),
		Port:                 443,
		NoRedirectServer:     false,
		Protocol:             "https",
		Hypervisor:           "kvm",
		Vga:                  "virtio",
		VgaRender:            "",
		AvailableRenders:     []string{},
		Gui:                  false,
		GuiUser:              "",
		GuiMode:              "",
		Certificates:         []primitive.ObjectID{},
		AdminDomain:          "",
		UserDomain:           "",
		WebauthnDomain:       "",
		RequestsMin:          56,
		ForwardedForHeader:   "",
		ForwardedProtoHeader: "",
		ExternalInterfaces:   []string{},
		ExternalInterfaces6:  []string{},
		InternalInterfaces:   []string{"bond0.2"},
		AvailableInterfaces: []ip.Interface{
			{Name: "bond0", Address: ""},
			{Name: "bond0.2", Address: "10.8.0.13"},
			{Name: "bond0.4", Address: "10.253.67.93"},
			{Name: "bonding_masters", Address: ""},
			{Name: "enp1s0f0", Address: ""},
			{Name: "enp1s0f1", Address: ""},
			{Name: "veth0", Address: ""},
		},
		AvailableBridges: []ip.Interface{
			{Name: "pritunlhost0", Address: "198.18.84.1"},
			{Name: "pritunlport0", Address: "198.19.96.1"},
		},
		AvailableVpcs:    []*cloud.Vpc{},
		CloudSubnets:     []string{},
		DefaultInterface: "bond0.4",
		NetworkMode:      "static",
		NetworkMode6:     "static",
		Blocks: []*node.BlockAttachment{
			{
				Interface: "bond0.4",
				Block:     utils.ObjectIdHex("689733b7a7a35eae0dbaea2f"),
			},
		},
		Blocks6: []*node.BlockAttachment{
			{
				Interface: "bond0.4",
				Block:     utils.ObjectIdHex("68973a47b5844593cf99cc7a"),
			},
		},
		Pools:  []primitive.ObjectID{},
		Shares: []*node.Share{},
		AvailableDrives: []*drive.Device{
			{Id: "nvme-INTEL_27Z1P0FGN"},
			{Id: "nvme-INTEL_27Z1P0FGN-part1"},
			{Id: "nvme-INTEL_27Z1P0FGN-part2"},
			{Id: "nvme-INTEL_27Z1P0FGN-part3"},
			{Id: "nvme-INTEL_42K1P0FGN"},
			{Id: "nvme-INTEL_42K1P0FGN-part1"},
			{Id: "nvme-INTEL_42K1P0FGN-part2"},
			{Id: "nvme-INTEL_42K1P0FGN-part3"},
		},
		InstanceDrives:          []*drive.Device{},
		NoHostNetwork:           false,
		NoNodePortNetwork:       false,
		HostNat:                 true,
		DefaultNoPublicAddress:  true,
		DefaultNoPublicAddress6: false,
		JumboFrames:             true,
		JumboFramesInternal:     true,
		Iscsi:                   false,
		LocalIsos:               nil,
		UsbPassthrough:          false,
		UsbDevices:              []*usb.Device{},
		PciPassthrough:          false,
		PciDevices:              []*pci.Device{},
		Hugepages:               false,
		HugepagesSize:           0,
		Firewall:                false,
		Roles:                   []string{"shape-m2"},
		Memory:                  41.02,
		HugePagesUsed:           0,
		Load1:                   24.63,
		Load5:                   23.32,
		Load15:                  21.43,
		CpuUnits:                128,
		MemoryUnits:             512,
		CpuUnitsRes:             64,
		MemoryUnitsRes:          298,
		PublicIps:               []string{"10.253.67.93"},
		PublicIps6: []string{
			"2001:db8:85a3:4d2f:9a47:3b6e:c829:4f16",
		},
		PrivateIps: map[string]string{
			"bond0.2": "10.8.0.13",
		},
		SoftwareVersion: constants.Version,
		Hostname:        "pritunl-east0",
		VirtPath:        "/var/lib/pritunl-cloud",
		CachePath:       "/var/cache/pritunl-cloud",
		TempPath:        "",
		OracleUser:      "",
		OracleTenancy:   "",
		OraclePublicKey: `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAu7nP2Kx4YmRa5TgL9Hs3
dW8JnF6QzV1MpC4BwX7Ka3Gt5RvN8YmL2Hs9Dx4TnQ6JpF8LwK3Vz5YmRs7NdT9H
gX2BpL4KwR8TmQ5NvJ3YxF7DsH9LwK2VmT8RgJ4NpX5BwL3HsY7TmQ9KxF2DpN8V
wJ3LsT4HgR9YmK5NxB2FwL7TsH8KpQ3XmJ4DgN9BwY2RsL5HxK7VmT3NpF8YwJ4D
gR2TsL9KxH5BmN3VwQ8JpY7LsT4HgK2NxR9FmL3BwJ5TsN8YpK4VxH7DgQ2LmR9N
wJ3BsT5HxL8KpN2YmF4DgR7VwQ3TsJ9KxH5BmL8NpY2FgT4RwK7VsH9JmQ3LxN8D
FwIDAQAB
-----END PUBLIC KEY-----`,
		Operation: "",
	},
	{
		Id:                   utils.ObjectIdHex("689733b2a7a35eae0dbaea09"),
		Datacenter:           utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Zone:                 utils.ObjectIdHex("689733b7a7a35eae0dbaea1e"),
		Name:                 "pritunl-east4",
		Comment:              "",
		Types:                []string{"admin", "hypervisor"},
		Timestamp:            time.Now(),
		Port:                 443,
		NoRedirectServer:     false,
		Protocol:             "https",
		Hypervisor:           "kvm",
		Vga:                  "virtio",
		VgaRender:            "",
		AvailableRenders:     []string{},
		Gui:                  false,
		GuiUser:              "",
		GuiMode:              "",
		Certificates:         []primitive.ObjectID{},
		AdminDomain:          "",
		UserDomain:           "",
		WebauthnDomain:       "",
		RequestsMin:          76,
		ForwardedForHeader:   "",
		ForwardedProtoHeader: "",
		ExternalInterfaces:   []string{},
		ExternalInterfaces6:  []string{},
		InternalInterfaces:   []string{"bond0.2"},
		AvailableInterfaces: []ip.Interface{
			{Name: "bond0", Address: ""},
			{Name: "bond0.2", Address: "10.8.0.14"},
			{Name: "bond0.4", Address: "10.253.67.94"},
			{Name: "bonding_masters", Address: ""},
			{Name: "enp1s0f0", Address: ""},
			{Name: "enp1s0f1", Address: ""},
			{Name: "veth0", Address: ""},
		},
		AvailableBridges: []ip.Interface{
			{Name: "pritunlhost0", Address: "198.18.84.1"},
			{Name: "pritunlport0", Address: "198.19.96.1"},
		},
		AvailableVpcs:    []*cloud.Vpc{},
		CloudSubnets:     []string{},
		DefaultInterface: "bond0.4",
		NetworkMode:      "static",
		NetworkMode6:     "static",
		Blocks: []*node.BlockAttachment{
			{
				Interface: "bond0.4",
				Block:     utils.ObjectIdHex("689733b7a7a35eae0dbaea2f"),
			},
		},
		Blocks6: []*node.BlockAttachment{
			{
				Interface: "bond0.4",
				Block:     utils.ObjectIdHex("68973a47b5844593cf99cc7a"),
			},
		},
		Pools:  []primitive.ObjectID{},
		Shares: []*node.Share{},
		AvailableDrives: []*drive.Device{
			{Id: "nvme-INTEL_27Z1P0FGN"},
			{Id: "nvme-INTEL_27Z1P0FGN-part1"},
			{Id: "nvme-INTEL_27Z1P0FGN-part2"},
			{Id: "nvme-INTEL_27Z1P0FGN-part3"},
			{Id: "nvme-INTEL_42K1P0FGN"},
			{Id: "nvme-INTEL_42K1P0FGN-part1"},
			{Id: "nvme-INTEL_42K1P0FGN-part2"},
			{Id: "nvme-INTEL_42K1P0FGN-part3"},
		},
		InstanceDrives:          []*drive.Device{},
		NoHostNetwork:           false,
		NoNodePortNetwork:       false,
		HostNat:                 true,
		DefaultNoPublicAddress:  true,
		DefaultNoPublicAddress6: false,
		JumboFrames:             true,
		JumboFramesInternal:     true,
		Iscsi:                   false,
		LocalIsos:               nil,
		UsbPassthrough:          false,
		UsbDevices:              []*usb.Device{},
		PciPassthrough:          false,
		PciDevices:              []*pci.Device{},
		Hugepages:               false,
		HugepagesSize:           0,
		Firewall:                false,
		Roles:                   []string{"shape-m2"},
		Memory:                  41.02,
		HugePagesUsed:           0,
		Load1:                   88.45,
		Load5:                   83.13,
		Load15:                  72.34,
		CpuUnits:                128,
		MemoryUnits:             512,
		CpuUnitsRes:             45,
		MemoryUnitsRes:          242,
		PublicIps:               []string{"10.253.67.94"},
		PublicIps6: []string{
			"2001:db8:85a3:4d2f:6f15:d8c3:2a94:e7b1",
		},
		PrivateIps: map[string]string{
			"bond0.2": "10.8.0.14",
		},
		SoftwareVersion: constants.Version,
		Hostname:        "pritunl-east0",
		VirtPath:        "/var/lib/pritunl-cloud",
		CachePath:       "/var/cache/pritunl-cloud",
		TempPath:        "",
		OracleUser:      "",
		OracleTenancy:   "",
		OraclePublicKey: `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAx9SLk4Pm3VaRtN8Yw2Hf
jK5TgQ7BnM2XpL4DwV8Ha6FsR9YmJ3KxC5NtQ8LwP7DgH2VmF4TsK9XpN3BwJ8Rg
mL5YxH7TsQ2NpK4VwF8DgJ3BmR9LxT5KsN2HwY7FpQ8VmJ4TgL3NxK9BwH5DsR2Y
mF7KpT8LwQ3VgN4JxH9BsK2TmL5DwR8YpF3NxQ7KgJ4VsH9BmT2LwK8DxN5YpF3R
gJ7TsL4KwH9VmQ2NxB8FpY3DgK5TsR9LwJ2HmN4BxF7VpQ8KgL3TsY9NwH5DmK4J
xR2BgF8TpL7VsN3KwQ9YmH4DgJ5BxT2LsK8FpN3VwR7YgH4TmQ9KsL2BxJ5NpF8D
RwIDAQAB
-----END PUBLIC KEY-----`,
		Operation: "",
	},
	{
		Id:                   utils.ObjectIdHex("689733b2a7a35eae0dbaea09"),
		Datacenter:           utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Zone:                 utils.ObjectIdHex("689733b7a7a35eae0dbaea1e"),
		Name:                 "pritunl-east5",
		Comment:              "",
		Types:                []string{"admin", "hypervisor"},
		Timestamp:            time.Now(),
		Port:                 443,
		NoRedirectServer:     false,
		Protocol:             "https",
		Hypervisor:           "kvm",
		Vga:                  "virtio",
		VgaRender:            "",
		AvailableRenders:     []string{},
		Gui:                  false,
		GuiUser:              "",
		GuiMode:              "",
		Certificates:         []primitive.ObjectID{},
		AdminDomain:          "",
		UserDomain:           "",
		WebauthnDomain:       "",
		RequestsMin:          12,
		ForwardedForHeader:   "",
		ForwardedProtoHeader: "",
		ExternalInterfaces:   []string{},
		ExternalInterfaces6:  []string{},
		InternalInterfaces:   []string{"bond0.2"},
		AvailableInterfaces: []ip.Interface{
			{Name: "bond0", Address: ""},
			{Name: "bond0.2", Address: "10.8.0.15"},
			{Name: "bond0.4", Address: "10.253.67.95"},
			{Name: "bonding_masters", Address: ""},
			{Name: "enp1s0f0", Address: ""},
			{Name: "enp1s0f1", Address: ""},
			{Name: "veth0", Address: ""},
		},
		AvailableBridges: []ip.Interface{
			{Name: "pritunlhost0", Address: "198.18.84.1"},
			{Name: "pritunlport0", Address: "198.19.96.1"},
		},
		AvailableVpcs:    []*cloud.Vpc{},
		CloudSubnets:     []string{},
		DefaultInterface: "bond0.4",
		NetworkMode:      "static",
		NetworkMode6:     "static",
		Blocks: []*node.BlockAttachment{
			{
				Interface: "bond0.4",
				Block:     utils.ObjectIdHex("689733b7a7a35eae0dbaea2f"),
			},
		},
		Blocks6: []*node.BlockAttachment{
			{
				Interface: "bond0.4",
				Block:     utils.ObjectIdHex("68973a47b5844593cf99cc7a"),
			},
		},
		Pools:  []primitive.ObjectID{},
		Shares: []*node.Share{},
		AvailableDrives: []*drive.Device{
			{Id: "nvme-INTEL_27Z1P0FGN"},
			{Id: "nvme-INTEL_27Z1P0FGN-part1"},
			{Id: "nvme-INTEL_27Z1P0FGN-part2"},
			{Id: "nvme-INTEL_27Z1P0FGN-part3"},
			{Id: "nvme-INTEL_42K1P0FGN"},
			{Id: "nvme-INTEL_42K1P0FGN-part1"},
			{Id: "nvme-INTEL_42K1P0FGN-part2"},
			{Id: "nvme-INTEL_42K1P0FGN-part3"},
		},
		InstanceDrives:          []*drive.Device{},
		NoHostNetwork:           false,
		NoNodePortNetwork:       false,
		HostNat:                 true,
		DefaultNoPublicAddress:  true,
		DefaultNoPublicAddress6: false,
		JumboFrames:             true,
		JumboFramesInternal:     true,
		Iscsi:                   false,
		LocalIsos:               nil,
		UsbPassthrough:          false,
		UsbDevices:              []*usb.Device{},
		PciPassthrough:          false,
		PciDevices:              []*pci.Device{},
		Hugepages:               false,
		HugepagesSize:           0,
		Firewall:                false,
		Roles:                   []string{"shape-m2"},
		Memory:                  41.02,
		HugePagesUsed:           0,
		Load1:                   53.44,
		Load5:                   43.75,
		Load15:                  43.93,
		CpuUnits:                128,
		MemoryUnits:             512,
		CpuUnitsRes:             76,
		MemoryUnitsRes:          346,
		PublicIps:               []string{"10.253.67.95"},
		PublicIps6: []string{
			"2001:db8:85a3:4d2f:8d2c:1fa7:b653:9e48",
		},
		PrivateIps: map[string]string{
			"bond0.2": "10.8.0.15",
		},
		SoftwareVersion: constants.Version,
		Hostname:        "pritunl-east0",
		VirtPath:        "/var/lib/pritunl-cloud",
		CachePath:       "/var/cache/pritunl-cloud",
		TempPath:        "",
		OracleUser:      "",
		OracleTenancy:   "",
		OraclePublicKey: `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1pK7Gm4XwQ9RsT2NvH8Bf
jL5DgY3KpM2TwV8NxH4FsQ9BmL7YgK3VpT8DwR5NxJ2HmF4KgQ7TsL9BwY3VpN8R
xK5DmH2FgT9LwJ4NsQ7BpV3KxY8TmL5HgR2DwN9FpK4VsT3BxJ7YmQ8LgH2NwK5D
pF9TxR3VsL4KgY7BmN2HwJ8FpQ5TxK9DsL3VgH4NmR7BwY2KpT8FxJ9LsQ3DgN5V
mH7BwK4TpL2RxF9NsY3KgQ8DmJ5VwH7BpT2LxK4FgN9RsQ3TmL8YwJ5DpH7KxF2N
gV4BsT9LmQ3KwR8YpN5DxH7FgJ2TsL4BwK9VmQ3NpY8RxT5LgH2DsK7BmF4JwN9Q
TwIDAQAB
-----END PUBLIC KEY-----`,
		Operation: "",
	},
}

// Policies
var Policies = []*policy.Policy{
	{
		Id:       utils.ObjectIdHex("67b8a03e4866ba90e6c45a8c"),
		Name:     "policy",
		Comment:  "",
		Disabled: false,
		Roles: []string{
			"pritunl",
		},
		Rules: map[string]*policy.Rule{
			"location": {
				Type:    "location",
				Disable: false,
				Values: []string{
					"US",
				},
			},
			"whitelist_networks": {
				Type:    "whitelist_networks",
				Disable: false,
				Values: []string{
					"10.0.0.0/8",
				},
			},
		},
		AdminSecondary:       primitive.ObjectID{},
		UserSecondary:        primitive.ObjectID{},
		AdminDeviceSecondary: true,
		UserDeviceSecondary:  true,
	},
}

// Certificates
var Certificates = []*certificate.Certificate{
	{
		Id:           utils.ObjectIdHex("67b89ef24866ba90e6c459e8"),
		Name:         "cloud-pritunl-com",
		Comment:      "",
		Organization: primitive.ObjectID{},
		Type:         "lets_encrypt",
		Key: `-----BEGIN RSA PRIVATE KEY-----
MIIJKQIBAAKCAgEAx9Y3Lk2AwV6ap7L/Sx9XC5mXaUf8hvMmDbLBqDZ1Y7xKJM2h
zQ8Xm1rK9q0wzQC6qiL6xHmTpKWTzNVzGsQdM3/qNPLNA7W8PIYCzjkSe5X1YktY
vxldBxYxPRJxXk5S9P8dFYVmFFKF2bvJ5pSMLq9w3z3nTm3TQtRPqWx2Vk3DqV2D
QKmNtqJnhVqYvVKa3QpLLwz8xKqB1sPXLr4XqQ3bz3fLjLxPmYV5WxLhgdKLYZTv
YxQPLPTJkX3Pw4XD4Qs4CrKLW5bYsqYKQ7kKDXgJmTxYzZLjZKf4vSqLxqV5bDPY
rR2YxQ9TKLkYKVMpNtY5J9X2fWzyPSvXqXZfVx7D8xJzDY8YKPLXmvxKQZxLJxSx
zxHQzYKJpX3YmVfqYYmfYxXYzLmYxDzSxXqLvKxVqXxQDsPxQVKfKqQx5KvxsVqD
-----END RSA PRIVATE KEY-----`,
		Certificate: `-----BEGIN CERTIFICATE-----
MIIGGTCCBQGgAwIBAgISBXx9YmN2KQm9g3Y5XmKbvx9YMA0GCSqGSIb3DQEBCwUA
MDMxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MQwwCgYDVQQD
EwNSMTEwHhcNMjUwODA4MDY0NzI3WhcNMjUxMTA2MDY0NzI2WjAcMRowGAYDVQQD
ExFjbG91ZC5wcml0dW5sLnJlZDCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoC
ggIBAMfWNy5NgMFemqey/0sfVwuZl2lH/IbzJg2ywag2dWO8SiTNoc0PF5tayvat
MM0AuKoi+sR5k6Slk8zVcxrEHTN/6jTyzQO1vDyGAs45EnuV9WJLWL8ZXQcWMT0S
cV5OUvT/HRWFZhRShdn5iQ2Sry6vcN8950Dt00LUT6lsdlZNw6ldg0CpjbaiZ4Va
mL1Smt0KSy8M/MSqgdbD1y6+F6kN2893y4y8T5mFeVsS4YHSi2GU72MUDyz0yZF9
z8OFw+ELOAqyi1uW2LKmCkO5Cg14CZk8WM2S42Sn+L0qi8aleWwz2K0dmMUPUyi5
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIFBjCCAu6gAwIBAgIRAIp9PhPWLzDvI4a9KQdrNPgwDQYJKoZIhvcNAQELBQAw
TzELMAkGA1UEBhMCVVMxKTAnBgNVBAoTIEludGVybmV0IFNlY3VyaXR5IFJlc2Vh
cmNoIEdyb3VwMRUwEwYDVQQDEwxJU1JHIFJvb3QgWDEwHhcNMjQwMzEzMDAwMDAw
WhcNMjcwMzEyMjM1OTU5WjAzMQswCQYDVQQGEwJVUzEWMBQGA1UEChMNTGV0J3Mg
RW5jcnlwdDEMMAoGA1UEAxMDUjExMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEAuoe8XBsAOcvKCs3UZxD5ATylTqVhyybKUvsVAbe5KPUoHu0nsyQYOWcJ
DAjs4DqwO3cOvfPlOVRBDE6uQdaZdN5R2+97/1i9qLcT9t4x1fJyyXJqC4N0lZxG
AGQUmfOx2SLZzaiSqhwmej/+71gFewiVgdtxD4774zEJuwm+UE1fj5F2PVqdnoPy
-----END CERTIFICATE-----`,
		Info: &certificate.Info{
			Hash:         "bba8a3941280c8466a6a2a723cc06f26",
			SignatureAlg: "SHA256-RSA",
			PublicKeyAlg: "RSA",
			Issuer:       "R11",
			IssuedOn:     time.Now(),
			ExpiresOn:    time.Now().Add(2160 * time.Hour),
			DnsNames: []string{
				"cloud.pritunl.com",
				"user.cloud.pritunl.com",
			},
		},
		AcmeDomains: []string{
			"cloud.pritunl.com",
			"user.cloud.pritunl.com",
		},
		AcmeType:   "acme_dns",
		AcmeAuth:   "acme_cloudflare",
		AcmeSecret: utils.ObjectIdHex("67b89e8d4866ba90e6c459ba"),
	},
}

// Secrets
var Secrets = []*secret.Secret{
	{
		Id:           utils.ObjectIdHex("67b89e8d4866ba90e6c459ba"),
		Name:         "cloudflare-pritunl-com",
		Comment:      "",
		Organization: primitive.ObjectID{},
		Type:         "cloudflare",
		Key:          "a7kX9mN2vP8Q-4jL6wS3tR5Y-uH1gF7dZ0xC-vB8nM",
		Value:        "",
		Region:       "",
		PublicKey: `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAz4K8Lm3QvR7WxN5YdE2P
jX9TpQ6HgM1wV0nS4KaF3ZcB8LrY5UvO2JmN7XsPqI1AgK8EoH3RdWzM9LfY2VtN
kP4QxGsJ7YnR8LwVmT3AqZ5HvK2NdP1XoS8JgR4LmW7YxQ3VnH5TsK9PpL2MdX8Rg
vJ3KqN5WxT1LsM4HgY7RdP8NqV2JmK5XwL3TsR8YgN4HxP1LdK9VwQ2MsT3XpR7Y
nL8KgJ5WdH3TmR9XsL2PqN7VxK4MgT3HdJ8YwP2LsK5RxT1NqM4JgY7PxR8WsL3T
mK9XwN2HgJ5YdL3RsP8VqT2MxK4NhR3JdY8WwL2TsM5QxN1PqK4YgJ7RxP8VsT3M
PwIDAQAB
-----END PUBLIC KEY-----`,
		Data: "",
	},
}

// Organizations
var Organizations = []*organization.Organization{
	{
		Id: utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
		Roles: []string{
			"pritunl",
		},
		Name:    "pritunl",
		Comment: "",
	},
}

// Datacenters
var Datacenters = []*datacenter.Datacenter{
	{
		Id:                 utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Name:               "us-west-1",
		Comment:            "",
		MatchOrganizations: false,
		Organizations:      []primitive.ObjectID{},
		NetworkMode:        "vxlan_vlan",
		WgMode:             "",
		PublicStorages: []primitive.ObjectID{
			utils.ObjectIdHex("689733b7a7a35eae0dbaea15"),
		},
		PrivateStorage:      primitive.ObjectID{},
		PrivateStorageClass: "",
		BackupStorage:       primitive.ObjectID{},
		BackupStorageClass:  "",
	},
}

// Zones
var Zones = []*zone.Zone{
	{
		Id:          utils.ObjectIdHex("689733b7a7a35eae0dbaea1e"),
		Datacenter:  utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Name:        "us-west-1a",
		Comment:     "",
		DnsServers:  []string{},
		DnsServers6: []string{},
	},
}

// Shapes
var Shapes = []*shape.Shape{
	{
		Id:               utils.ObjectIdHex("65e6e303ceeebbb3dabaec96"),
		Name:             "m2-small",
		Comment:          "",
		Type:             "instance",
		DeleteProtection: false,
		Datacenter:       utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Roles: []string{
			"shape-m2",
		},
		Flexible:   true,
		DiskType:   "qcow2",
		DiskPool:   primitive.ObjectID{},
		Memory:     2048,
		Processors: 1,
		NodeCount:  1,
	},
	{
		Id:               utils.ObjectIdHex("65e6e2ecceeebbb3dabaec79"),
		Name:             "m2-medium",
		Comment:          "",
		Type:             "instance",
		DeleteProtection: false,
		Datacenter:       utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Roles: []string{
			"shape-m2",
		},
		Flexible:   true,
		DiskType:   "qcow2",
		DiskPool:   primitive.ObjectID{},
		Memory:     4096,
		Processors: 2,
		NodeCount:  1,
	},
	{
		Id:               utils.ObjectIdHex("66f63282aac06d53e8c9c435"),
		Name:             "m2-large",
		Comment:          "",
		Type:             "instance",
		DeleteProtection: false,
		Datacenter:       utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Roles: []string{
			"shape-m2",
		},
		Flexible:   true,
		DiskType:   "qcow2",
		DiskPool:   primitive.ObjectID{},
		Memory:     8192,
		Processors: 4,
		NodeCount:  1,
	},
}

// Blocks
var Blocks = []*block.Block{
	{
		Id:      utils.ObjectIdHex("689733b7a7a35eae0dbaea2f"),
		Name:    "east-public",
		Comment: "",
		Type:    "ipv4",
		Vlan:    0,
		Subnets: []string{
			"10.253.67.0/24",
		},
		Subnets6: []string{},
		Excludes: []string{
			"10.253.67.90/24",
			"10.253.67.91/24",
			"10.253.67.92/24",
			"10.253.67.93/24",
			"10.253.67.94/24",
			"10.253.67.95/24",
		},
		Netmask:  "255.255.255.0",
		Gateway:  "10.253.67.1",
		Gateway6: "",
	},
	{
		Id:      utils.ObjectIdHex("68973a47b5844593cf99cc7a"),
		Name:    "east-public6",
		Comment: "",
		Type:    "ipv6",
		Vlan:    0,
		Subnets: []string{},
		Subnets6: []string{
			"2001:db8:85a3:4d2f::/64",
		},
		Excludes: []string{},
		Netmask:  "",
		Gateway:  "",
		Gateway6: "2001:db8:85a3:4d2f::1",
	},
}

// Vpcs
var Vpcs = []*vpc.Vpc{
	{
		Id:       utils.ObjectIdHex("689733b7a7a35eae0dbaea23"),
		Name:     "production",
		Comment:  "",
		VpcId:    2996,
		Network:  "10.196.0.0/14",
		Network6: "fd97:30bf:d456:a3bc::/64",
		Subnets: []*vpc.Subnet{
			{
				Id:      utils.ObjectIdHex("66a076d5fafc270786e93461"),
				Name:    "primary",
				Network: "10.196.1.0/24",
			},
			{
				Id:      utils.ObjectIdHex("66a076d5fafc270786e93462"),
				Name:    "management",
				Network: "10.196.2.0/24",
			},
			{
				Id:      utils.ObjectIdHex("66a076d5fafc270786e93463"),
				Name:    "link",
				Network: "10.196.3.0/24",
			},
			{
				Id:      utils.ObjectIdHex("66a076d5fafc270786e93464"),
				Name:    "database",
				Network: "10.196.4.0/24",
			},
			{
				Id:      utils.ObjectIdHex("66a076d5fafc270786e93465"),
				Name:    "web",
				Network: "10.196.5.0/24",
			},
			{
				Id:      utils.ObjectIdHex("66a076d5fafc270786e93466"),
				Name:    "search",
				Network: "10.196.6.0/24",
			},
			{
				Id:      utils.ObjectIdHex("66a076d5fafc270786e93467"),
				Name:    "vpn",
				Network: "10.196.7.0/24",
			},
			{
				Id:      utils.ObjectIdHex("66a076d5fafc270786e93468"),
				Name:    "balancer",
				Network: "10.196.8.0/24",
			},
		},
		Organization:  utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
		Datacenter:    utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		IcmpRedirects: false,
		Routes: []*vpc.Route{
			&vpc.Route{
				Destination: "10.24.0.0/16",
				Target:      "10.196.7.2",
			},
		},
		Maps:             []*vpc.Map{},
		Arps:             []*vpc.Arp{},
		DeleteProtection: false,
	},
	{
		Id:       utils.ObjectIdHex("689733b7a7a35eae0dbaea23"),
		Name:     "testing",
		Comment:  "",
		VpcId:    2732,
		Network:  "10.224.0.0/14",
		Network6: "fd97:30bf:d456:a3bc::/64",
		Subnets: []*vpc.Subnet{
			{
				Id:      utils.ObjectIdHex("689733b7a7a35eae0dbaea61"),
				Name:    "primary",
				Network: "10.224.1.0/24",
			},
			{
				Id:      utils.ObjectIdHex("689733b7a7a35eae0dbaea62"),
				Name:    "management",
				Network: "10.224.2.0/24",
			},
			{
				Id:      utils.ObjectIdHex("689733b7a7a35eae0dbaea63"),
				Name:    "link",
				Network: "10.224.3.0/24",
			},
			{
				Id:      utils.ObjectIdHex("689733b7a7a35eae0dbaea64"),
				Name:    "database",
				Network: "10.224.4.0/24",
			},
			{
				Id:      utils.ObjectIdHex("689733b7a7a35eae0dbaea65"),
				Name:    "web",
				Network: "10.224.5.0/24",
			},
			{
				Id:      utils.ObjectIdHex("689733b7a7a35eae0dbaea66"),
				Name:    "search",
				Network: "10.224.6.0/24",
			},
			{
				Id:      utils.ObjectIdHex("689733b7a7a35eae0dbaea67"),
				Name:    "vpn",
				Network: "10.224.7.0/24",
			},
			{
				Id:      utils.ObjectIdHex("689733b7a7a35eae0dbaea68"),
				Name:    "balancer",
				Network: "10.224.8.0/24",
			},
		},
		Organization:  utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
		Datacenter:    utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		IcmpRedirects: false,
		Routes: []*vpc.Route{
			&vpc.Route{
				Destination: "10.36.0.0/16",
				Target:      "10.224.7.2",
			},
		},
		Maps:             []*vpc.Map{},
		Arps:             []*vpc.Arp{},
		DeleteProtection: false,
	},
}

// Storages
var Storages = []*storage.Storage{
	{
		Id:        utils.ObjectIdHex("689733b7a7a35eae0dbaea15"),
		Name:      "pritunl-images",
		Comment:   "",
		Type:      "public",
		Endpoint:  "images.pritunl.com",
		Bucket:    "stable",
		AccessKey: "",
		SecretKey: "",
		Insecure:  false,
	},
	{
		Id:        utils.ObjectIdHex("689733b7a7a35eae0dbaea16"),
		Name:      "pritunl-storage",
		Comment:   "",
		Type:      "private",
		Endpoint:  "s3.amazonaws.com",
		Bucket:    "pritunl-cloud-2943",
		AccessKey: "AKIAJTVJ15RORHDU7M1M",
		SecretKey: "VLBGHOVTKDP5SIRSEC8R4XFQWLCIYN4HK",
		Insecure:  false,
	},
}

// Balancers
var Balancers = []*balancer.Balancer{
	{
		Id:           utils.ObjectIdHex("61ba27ccf149d4c222b23247"),
		Name:         "web-app",
		Comment:      "",
		Type:         "http",
		State:        true,
		Organization: utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
		Datacenter:   utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Certificates: []primitive.ObjectID{
			utils.ObjectIdHex("67b89ef24866ba90e6c459e8"),
		},
		ClientAuthority: primitive.ObjectID{},
		WebSockets:      false,
		Domains: []*balancer.Domain{
			{
				Domain: "demo.cloud.pritunl.com",
				Host:   "",
			},
		},
		Backends: []*balancer.Backend{
			{
				Protocol: "http",
				Hostname: "10.234.10.22",
				Port:     8000,
			},
			{
				Protocol: "http",
				Hostname: "10.234.10.24",
				Port:     8000,
			},
		},
		States: map[string]*balancer.State{
			"65b5d7e1c2e9a21159765955": {
				Timestamp:  time.Now(),
				Requests:   125,
				Retries:    0,
				WebSockets: 0,
				Online: []string{
					"10.234.10.22:8000",
					"10.234.10.24:8000",
				},
				UnknownHigh: []string{},
				UnknownMid:  []string{},
				UnknownLow:  []string{},
				Offline:     []string{},
			},
		},
		CheckPath: "/check",
	},
}

// Plans
var Plans = []*plan.Plan{
	{
		Id:           utils.ObjectIdHex("66e8993f1fbc6db8e20819f8"),
		Name:         "primary",
		Comment:      "",
		Organization: utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
		Statements: []*plan.Statement{
			{
				Id:        utils.ObjectIdHex("67c9bed42c125c5ddf24d0a1"),
				Statement: "IF instance.last_timestamp < 60 AND instance.last_heartbeat > 60 FOR 15 THEN 'stop'",
			},
			{
				Id:        utils.ObjectIdHex("683d645e2956cdd93d3e08d2"),
				Statement: "IF instance.state != 'running' THEN 'start'",
			},
		},
	},
}

// Domains
var Domains = []*domain.Domain{
	{
		Id:            utils.ObjectIdHex("67b8a1d24866ba90e6c45b5b"),
		Name:          "pritunl-com",
		Comment:       "",
		Organization:  utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
		Type:          "cloudflare",
		Secret:        utils.ObjectIdHex("67b89e8d4866ba90e6c459ba"),
		RootDomain:    "pritunl.com",
		LockId:        primitive.ObjectID{},
		LockTimestamp: time.Time{},
		LastUpdate:    time.Now(),
		Records: []*domain.Record{
			{
				Id:              utils.ObjectIdHex("68076c9f06fd0087c078dfdc"),
				Domain:          utils.ObjectIdHex("67b8a1d24866ba90e6c45b5b"),
				Resource:        primitive.ObjectID{},
				Deployment:      utils.ObjectIdHex("68076bb954e947708aa6d651"),
				Timestamp:       time.Now(),
				DeleteTimestamp: time.Time{},
				SubDomain:       "demo",
				Type:            "A",
				Value:           "10.196.8.2",
				Operation:       "",
			},
			{
				Id:              utils.ObjectIdHex("68076ca306fd0087c078dfdd"),
				Domain:          utils.ObjectIdHex("67b8a1d24866ba90e6c45b5b"),
				Resource:        primitive.ObjectID{},
				Deployment:      utils.ObjectIdHex("68076bb954e947708aa6d651"),
				Timestamp:       time.Now(),
				DeleteTimestamp: time.Time{},
				SubDomain:       "cloud",
				Type:            "A",
				Value:           "10.196.8.12",
				Operation:       "",
			},
			{
				Id:              utils.ObjectIdHex("68076ca406fd0087c078dfde"),
				Domain:          utils.ObjectIdHex("67b8a1d24866ba90e6c45b5b"),
				Resource:        primitive.ObjectID{},
				Deployment:      utils.ObjectIdHex("68076bb954e947708aa6d651"),
				Timestamp:       time.Now(),
				DeleteTimestamp: time.Time{},
				SubDomain:       "user.cloud",
				Type:            "A",
				Value:           "10.196.8.12",
				Operation:       "",
			},
			{
				Id:              utils.ObjectIdHex("6813705806fd0087c078dfe1"),
				Domain:          utils.ObjectIdHex("67b8a1d24866ba90e6c45b5b"),
				Resource:        primitive.ObjectID{},
				Deployment:      utils.ObjectIdHex("68136f7d43b4ac1351f54f0a"),
				Timestamp:       time.Now(),
				DeleteTimestamp: time.Time{},
				SubDomain:       "demo.cloud",
				Type:            "A",
				Value:           "10.196.8.46",
				Operation:       "",
			},
			{
				Id:              utils.ObjectIdHex("681e01394230fad44c6a5140"),
				Domain:          utils.ObjectIdHex("67b8a1d24866ba90e6c45b5b"),
				Resource:        primitive.ObjectID{},
				Deployment:      utils.ObjectIdHex("681e01308d67187e275a847a"),
				Timestamp:       time.Now(),
				DeleteTimestamp: time.Time{},
				SubDomain:       "forum",
				Type:            "AAAA",
				Value:           "2001:db8:85a3:42:d5c:82ca:9ed4:854b",
				Operation:       "",
			},
			{
				Id:              utils.ObjectIdHex("683e86d74230fad44c6a514d"),
				Domain:          utils.ObjectIdHex("67b8a1d24866ba90e6c45b5b"),
				Resource:        primitive.ObjectID{},
				Deployment:      utils.ObjectIdHex("683dcdf13249b43a9cc5ec70"),
				Timestamp:       time.Now(),
				DeleteTimestamp: time.Time{},
				SubDomain:       "docs",
				Type:            "CNAME",
				Value:           "docs.pritunl.dev",
				Operation:       "",
			},
		},
	},
}
