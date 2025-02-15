package node

import (
	"container/list"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/bridges"
	"github.com/pritunl/pritunl-cloud/cloud"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/drive"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/iso"
	"github.com/pritunl/pritunl-cloud/lvm"
	"github.com/pritunl/pritunl-cloud/pci"
	"github.com/pritunl/pritunl-cloud/render"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/usb"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/zone"
	"github.com/pritunl/webauthn/webauthn"
	"github.com/sirupsen/logrus"
)

var (
	Self *Node
)

type Node struct {
	Id                      primitive.ObjectID   `bson:"_id" json:"id"`
	Zone                    primitive.ObjectID   `bson:"zone,omitempty" json:"zone"`
	Name                    string               `bson:"name" json:"name"`
	Comment                 string               `bson:"comment" json:"comment"`
	Types                   []string             `bson:"types" json:"types"`
	Timestamp               time.Time            `bson:"timestamp" json:"timestamp"`
	Port                    int                  `bson:"port" json:"port"`
	NoRedirectServer        bool                 `bson:"no_redirect_server" json:"no_redirect_server"`
	Protocol                string               `bson:"protocol" json:"protocol"`
	Hypervisor              string               `bson:"hypervisor" json:"hypervisor"`
	Vga                     string               `bson:"vga" json:"vga"`
	VgaRender               string               `bson:"vga_render" json:"vga_render"`
	AvailableRenders        []string             `bson:"available_renders" json:"available_renders"`
	Gui                     bool                 `bson:"gui" json:"gui"`
	GuiUser                 string               `bson:"gui_user" json:"gui_user"`
	GuiMode                 string               `bson:"gui_mode" json:"gui_mode"`
	Certificate             primitive.ObjectID   `bson:"certificate" json:"certificate"`
	Certificates            []primitive.ObjectID `bson:"certificates" json:"certificates"`
	SelfCertificate         string               `bson:"self_certificate_key" json:"-"`
	SelfCertificateKey      string               `bson:"self_certificate" json:"-"`
	AdminDomain             string               `bson:"admin_domain" json:"admin_domain"`
	UserDomain              string               `bson:"user_domain" json:"user_domain"`
	WebauthnDomain          string               `bson:"webauthn_domain" json:"webauthn_domain"`
	RequestsMin             int64                `bson:"requests_min" json:"requests_min"`
	ForwardedForHeader      string               `bson:"forwarded_for_header" json:"forwarded_for_header"`
	ForwardedProtoHeader    string               `bson:"forwarded_proto_header" json:"forwarded_proto_header"`
	ExternalInterface       string               `bson:"external_interface" json:"external_interface"`
	InternalInterface       string               `bson:"internal_interface" json:"internal_interface"`
	ExternalInterfaces      []string             `bson:"external_interfaces" json:"external_interfaces"`
	ExternalInterfaces6     []string             `bson:"external_interfaces6" json:"external_interfaces6"`
	InternalInterfaces      []string             `bson:"internal_interfaces" json:"internal_interfaces"`
	AvailableInterfaces     []string             `bson:"available_interfaces" json:"available_interfaces"`
	AvailableBridges        []string             `bson:"available_bridges" json:"available_bridges"`
	AvailableVpcs           []*cloud.Vpc         `bson:"available_vpcs" json:"available_vpcs"`
	OracleSubnets           []string             `bson:"oracle_subnets" json:"oracle_subnets"`
	DefaultInterface        string               `bson:"default_interface" json:"default_interface"`
	NetworkMode             string               `bson:"network_mode" json:"network_mode"`
	NetworkMode6            string               `bson:"network_mode6" json:"network_mode6"`
	Blocks                  []*BlockAttachment   `bson:"blocks" json:"blocks"`
	Blocks6                 []*BlockAttachment   `bson:"blocks6" json:"blocks6"`
	Pools                   []primitive.ObjectID `bson:"pools" json:"pools"`
	AvailableDrives         []*drive.Device      `bson:"available_drives" json:"available_drives"`
	InstanceDrives          []*drive.Device      `bson:"instance_drives" json:"instance_drives"`
	NoHostNetwork           bool                 `bson:"no_host_network" json:"no_host_network"`
	HostNat                 bool                 `bson:"host_nat" json:"host_nat"`
	DefaultNoPublicAddress  bool                 `bson:"default_no_public_address" json:"default_no_public_address"`
	DefaultNoPublicAddress6 bool                 `bson:"default_no_public_address6" json:"default_no_public_address6"`
	JumboFrames             bool                 `bson:"jumbo_frames" json:"jumbo_frames"`
	JumboFramesInternal     bool                 `bson:"jumbo_frames_internal" json:"jumbo_frames_internal"`
	Iscsi                   bool                 `bson:"iscsi" json:"iscsi"`
	LocalIsos               []*iso.Iso           `bson:"local_isos" json:"local_isos"`
	UsbPassthrough          bool                 `bson:"usb_passthrough" json:"usb_passthrough"`
	UsbDevices              []*usb.Device        `bson:"usb_devices" json:"usb_devices"`
	PciPassthrough          bool                 `bson:"pci_passthrough" json:"pci_passthrough"`
	PciDevices              []*pci.Device        `bson:"pci_devices" json:"pci_devices"`
	Hugepages               bool                 `bson:"hugepages" json:"hugepages"`
	HugepagesSize           int                  `bson:"hugepages_size" json:"hugepages_size"`
	Firewall                bool                 `bson:"firewall" json:"firewall"`
	NetworkRoles            []string             `bson:"network_roles" json:"network_roles"`
	Memory                  float64              `bson:"memory" json:"memory"`
	HugePagesUsed           float64              `bson:"hugepages_used" json:"hugepages_used"`
	Load1                   float64              `bson:"load1" json:"load1"`
	Load5                   float64              `bson:"load5" json:"load5"`
	Load15                  float64              `bson:"load15" json:"load15"`
	CpuUnits                int                  `bson:"cpu_units" json:"cpu_units"`
	MemoryUnits             float64              `bson:"memory_units" json:"memory_units"`
	CpuUnitsRes             int                  `bson:"cpu_units_res" json:"cpu_units_res"`
	MemoryUnitsRes          float64              `bson:"memory_units_res" json:"memory_units_res"`
	PublicIps               []string             `bson:"public_ips" json:"public_ips"`
	PublicIps6              []string             `bson:"public_ips6" json:"public_ips6"`
	PrivateIps              map[string]string    `bson:"private_ips" json:"private_ips"`
	SoftwareVersion         string               `bson:"software_version" json:"software_version"`
	Hostname                string               `bson:"hostname" json:"hostname"`
	Version                 int                  `bson:"version" json:"-"`
	VirtPath                string               `bson:"virt_path" json:"virt_path"`
	CachePath               string               `bson:"cache_path" json:"cache_path"`
	TempPath                string               `bson:"temp_path" json:"temp_path"`
	OracleUser              string               `bson:"oracle_user" json:"oracle_user"`
	OraclePrivateKey        string               `bson:"oracle_private_key" json:"-"`
	OraclePublicKey         string               `bson:"oracle_public_key" json:"oracle_public_key"`
	Operation               string               `bson:"operation" json:"operation"`
	oracleSubnetsNamed      []*OracleSubnet      `bson:"-" json:"-"`
	reqLock                 sync.Mutex           `bson:"-" json:"-"`
	reqCount                *list.List           `bson:"-" json:"-"`
	dcId                    primitive.ObjectID   `bson:"-" json:"-"`
	dcZoneId                primitive.ObjectID   `bson:"-" json:"-"`
}

type OracleSubnet struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (n *Node) Copy() *Node {
	nde := &Node{
		Id:                      n.Id,
		Zone:                    n.Zone,
		Name:                    n.Name,
		Comment:                 n.Comment,
		Types:                   n.Types,
		Timestamp:               n.Timestamp,
		Port:                    n.Port,
		NoRedirectServer:        n.NoRedirectServer,
		Protocol:                n.Protocol,
		Hypervisor:              n.Hypervisor,
		Vga:                     n.Vga,
		VgaRender:               n.VgaRender,
		Gui:                     n.Gui,
		GuiUser:                 n.GuiUser,
		GuiMode:                 n.GuiMode,
		AvailableRenders:        n.AvailableRenders,
		Certificate:             n.Certificate,
		Certificates:            n.Certificates,
		SelfCertificate:         n.SelfCertificate,
		SelfCertificateKey:      n.SelfCertificateKey,
		AdminDomain:             n.AdminDomain,
		UserDomain:              n.UserDomain,
		RequestsMin:             n.RequestsMin,
		ForwardedForHeader:      n.ForwardedForHeader,
		ForwardedProtoHeader:    n.ForwardedProtoHeader,
		ExternalInterface:       n.ExternalInterface,
		InternalInterface:       n.InternalInterface,
		ExternalInterfaces:      n.ExternalInterfaces,
		ExternalInterfaces6:     n.ExternalInterfaces6,
		InternalInterfaces:      n.InternalInterfaces,
		AvailableInterfaces:     n.AvailableInterfaces,
		AvailableBridges:        n.AvailableBridges,
		AvailableVpcs:           n.AvailableVpcs,
		OracleSubnets:           n.OracleSubnets,
		DefaultInterface:        n.DefaultInterface,
		NetworkMode:             n.NetworkMode,
		NetworkMode6:            n.NetworkMode6,
		Blocks:                  n.Blocks,
		Blocks6:                 n.Blocks6,
		Pools:                   n.Pools,
		AvailableDrives:         n.AvailableDrives,
		InstanceDrives:          n.InstanceDrives,
		NoHostNetwork:           n.NoHostNetwork,
		HostNat:                 n.HostNat,
		DefaultNoPublicAddress:  n.DefaultNoPublicAddress,
		DefaultNoPublicAddress6: n.DefaultNoPublicAddress6,
		JumboFrames:             n.JumboFrames,
		JumboFramesInternal:     n.JumboFramesInternal,
		Iscsi:                   n.Iscsi,
		LocalIsos:               n.LocalIsos,
		UsbPassthrough:          n.UsbPassthrough,
		UsbDevices:              n.UsbDevices,
		PciPassthrough:          n.PciPassthrough,
		PciDevices:              n.PciDevices,
		Hugepages:               n.Hugepages,
		HugepagesSize:           n.HugepagesSize,
		Firewall:                n.Firewall,
		NetworkRoles:            n.NetworkRoles,
		Memory:                  n.Memory,
		Load1:                   n.Load1,
		Load5:                   n.Load5,
		Load15:                  n.Load15,
		CpuUnits:                n.CpuUnits,
		MemoryUnits:             n.MemoryUnits,
		CpuUnitsRes:             n.CpuUnitsRes,
		MemoryUnitsRes:          n.MemoryUnitsRes,
		PublicIps:               n.PublicIps,
		PublicIps6:              n.PublicIps6,
		PrivateIps:              n.PrivateIps,
		SoftwareVersion:         n.SoftwareVersion,
		Hostname:                n.Hostname,
		Version:                 n.Version,
		VirtPath:                n.VirtPath,
		CachePath:               n.CachePath,
		TempPath:                n.TempPath,
		OracleUser:              n.OracleUser,
		OraclePrivateKey:        n.OraclePrivateKey,
		OraclePublicKey:         n.OraclePublicKey,
		Operation:               n.Operation,
		dcId:                    n.dcId,
		dcZoneId:                n.dcZoneId,
	}

	return nde
}

func (n *Node) AddRequest() {
	n.reqLock.Lock()
	back := n.reqCount.Back()
	back.Value = back.Value.(int) + 1
	n.reqLock.Unlock()
}

func (n *Node) GetVirtPath() string {
	if n.VirtPath == "" {
		return constants.DefaultRoot
	}
	return n.VirtPath
}

func (n *Node) GetCachePath() string {
	if n.CachePath == "" {
		return constants.DefaultCache
	}
	return n.CachePath
}

func (n *Node) GetTempPath() string {
	if n.TempPath == "" {
		return constants.DefaultTemp
	}
	return n.TempPath
}

func (n *Node) GetDatacenter(db *database.Database) (
	dcId primitive.ObjectID, err error) {

	if n.Zone == n.dcZoneId {
		dcId = n.dcId
		return
	}

	zne, err := zone.Get(db, n.Zone)
	if err != nil {
		return
	}

	dcId = zne.Datacenter
	n.dcId = zne.Datacenter
	n.dcZoneId = n.Zone

	return
}

func (n *Node) GetOracleSubnetsName() (subnets []*OracleSubnet) {
	if n.oracleSubnetsNamed != nil {
		subnets = n.oracleSubnetsNamed
		return
	}

	names := map[string]string{}

	if n.AvailableVpcs != nil {
		for _, vpc := range n.AvailableVpcs {
			for _, subnet := range vpc.Subnets {
				names[subnet.Id] = fmt.Sprintf(
					"%s - %s", vpc.Name, subnet.Name)
			}
		}
	}

	subnets = []*OracleSubnet{}

	if n.OracleSubnets != nil {
		for _, subnetId := range n.OracleSubnets {
			name := names[subnetId]
			if name == "" {
				name = subnetId
			}

			subnets = append(subnets, &OracleSubnet{
				Id:   subnetId,
				Name: name,
			})
		}
	}

	n.oracleSubnetsNamed = subnets

	return
}

func (n *Node) IsAdmin() bool {
	for _, typ := range n.Types {
		if typ == Admin {
			return true
		}
	}
	return false
}

func (n *Node) IsUser() bool {
	for _, typ := range n.Types {
		if typ == User {
			return true
		}
	}
	return false
}

func (n *Node) IsBalancer() bool {
	for _, typ := range n.Types {
		if typ == Balancer {
			return true
		}
	}
	return false
}

func (n *Node) IsHypervisor() bool {
	for _, typ := range n.Types {
		if typ == Hypervisor {
			return true
		}
	}
	return false
}

func (n *Node) IsOnline() bool {
	if time.Since(n.Timestamp) > time.Duration(
		settings.System.NodeTimestampTtl)*time.Second {

		return false
	}
	return true
}

func (n *Node) Usage() int {
	memoryUsage := float64(n.MemoryUnitsRes) / float64(n.MemoryUnits)
	if memoryUsage > 1.0 {
		memoryUsage = 1.0
	}

	cpuUsage := float64(n.CpuUnitsRes) / float64(n.CpuUnits)
	if cpuUsage > 1.0 {
		cpuUsage = 1.0
	}

	totalUsage := (memoryUsage * 0.75) + (cpuUsage * 0.25)
	if totalUsage > 1.0 {
		totalUsage = 1.0
	}

	return int(totalUsage * 100)
}

func (n *Node) SizeResource(memory, processors int) bool {
	memoryUnits := float64(memory) / float64(1024)

	if memoryUnits+n.MemoryUnitsRes > n.MemoryUnits {
		return false
	}

	if processors+n.CpuUnitsRes > n.CpuUnits*2 {
		return false
	}

	return true
}

func (n *Node) GetOracleAuthProvider() (pv *NodeOracleAuthProvider) {
	pv = &NodeOracleAuthProvider{
		nde: n,
	}
	return
}

func (n *Node) GetWebauthn(origin string, strict bool) (
	web *webauthn.WebAuthn, err error) {

	webauthnDomain := n.WebauthnDomain
	if webauthnDomain == "" {
		if strict {
			err = &errortypes.ReadError{
				errors.New("node: Webauthn domain not configured"),
			}
			return
		} else {
			userN := strings.Count(n.UserDomain, ".")
			adminN := strings.Count(n.AdminDomain, ".")

			if userN <= adminN {
				webauthnDomain = n.UserDomain
			} else {
				webauthnDomain = n.AdminDomain
			}
		}
	}

	web, err = webauthn.New(&webauthn.Config{
		RPDisplayName: "Pritunl Cloud",
		RPID:          webauthnDomain,
		RPOrigin:      origin,
	})
	if err != nil {
		err = utils.ParseWebauthnError(err)
		return
	}

	return
}

func (n *Node) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	n.Name = utils.FilterName(n.Name)

	if n.Hypervisor == "" {
		n.Hypervisor = Kvm
	}

	switch n.Vga {
	case Std, Vmware, Virtio:
		n.VgaRender = ""
		break
	case VirtioEgl, VirtioEglVulkan:
		if n.VgaRender != "" {
			found := false
			for _, rendr := range n.AvailableRenders {
				if n.VgaRender == rendr {
					found = true
					break
				}
			}

			if !found {
				errData = &errortypes.ErrorData{
					Error:   "node_vga_render_invalid",
					Message: "Invalid EGL render",
				}
				return
			}
		}
		break
	case "":
		n.Vga = Virtio
		n.VgaRender = ""
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "node_vga_invalid",
			Message: "Invalid VGA type",
		}
		return
	}

	if n.Gui {
		if n.GuiUser == "" {
			errData = &errortypes.ErrorData{
				Error:   "gui_user_missing",
				Message: "Desktop GUI user must be set",
			}
			return
		}

		switch n.GuiMode {
		case Sdl, "":
			n.GuiMode = Sdl
			break
		case Gtk:
			break
		default:
			errData = &errortypes.ErrorData{
				Error:   "gui_mode_invalid",
				Message: "Invalid desktop GUI mode",
			}
			return
		}
	} else {
		n.GuiUser = ""
		n.GuiMode = ""
	}

	if n.Protocol != "http" && n.Protocol != "https" {
		errData = &errortypes.ErrorData{
			Error:   "node_protocol_invalid",
			Message: "Invalid node server protocol",
		}
		return
	}

	if n.Port < 1 || n.Port > 65535 {
		errData = &errortypes.ErrorData{
			Error:   "node_port_invalid",
			Message: "Invalid node server port",
		}
		return
	}

	if n.Certificates == nil || n.Protocol != "https" {
		n.Certificates = []primitive.ObjectID{}
	}

	if n.Types == nil {
		n.Types = []string{}
	}

	for _, typ := range n.Types {
		switch typ {
		case Admin, User, Balancer, Hypervisor:
			break
		default:
			errData = &errortypes.ErrorData{
				Error:   "type_invalid",
				Message: "Invalid node type",
			}
			return
		}
	}

	if !n.IsBalancer() && ((n.IsAdmin() && !n.IsUser()) ||
		(n.IsUser() && !n.IsAdmin())) {

		n.AdminDomain = ""
		n.UserDomain = ""
	} else {
		if !n.IsAdmin() {
			n.AdminDomain = ""
		}
		if !n.IsUser() {
			n.UserDomain = ""
		}
	}

	if !n.Zone.IsZero() {
		coll := db.Zones()
		count, e := coll.CountDocuments(db, &bson.M{
			"_id": n.Zone,
		})
		if e != nil {
			err = database.ParseError(e)
			return
		}

		if count == 0 {
			n.Zone = primitive.NilObjectID
		}
	}

	if n.VirtPath == "" {
		n.VirtPath = constants.DefaultRoot
	}
	if n.CachePath == "" {
		n.CachePath = constants.DefaultCache
	}

	if n.NetworkRoles == nil || !n.Firewall {
		n.NetworkRoles = []string{}
	}

	if n.Firewall && len(n.NetworkRoles) == 0 {
		errData = &errortypes.ErrorData{
			Error:   "firewall_empty_roles",
			Message: "Cannot enable firewall without network roles",
		}
		return
	}

	if n.ExternalInterfaces == nil {
		n.ExternalInterfaces = []string{}
	}
	if n.InternalInterfaces == nil {
		n.InternalInterfaces = []string{}
	}
	if n.Blocks == nil {
		n.Blocks = []*BlockAttachment{}
	}
	if n.ExternalInterfaces6 == nil {
		n.ExternalInterfaces6 = []string{}
	}
	if n.Blocks6 == nil {
		n.Blocks6 = []*BlockAttachment{}
	}

	if n.OracleSubnets == nil {
		n.OracleSubnets = []string{}
	}

	instanceDrives := []*drive.Device{}
	if n.InstanceDrives != nil {
		for _, device := range n.InstanceDrives {
			device.Id = utils.FilterPath(device.Id, 512)
			instanceDrives = append(instanceDrives, device)
		}
	}
	n.InstanceDrives = instanceDrives

	switch n.NetworkMode {
	case Static:
		for _, blckAttch := range n.Blocks {
			blck, e := block.Get(db, blckAttch.Block)
			if e != nil {
				err = e
				if _, ok := err.(*database.NotFoundError); ok {
					err = nil
				} else {
					return
				}
			}

			if blck == nil || blck.Type != block.IPv4 {
				errData = &errortypes.ErrorData{
					Error:   "invalid_block",
					Message: "External IPv4 block invalid",
				}
				return
			}
		}

		break
	case Oracle:
		n.Blocks = []*BlockAttachment{}
		break
	case Dhcp, "":
		n.NetworkMode = Dhcp
		n.Blocks = []*BlockAttachment{}
		break
	case Disabled:
		n.Blocks = []*BlockAttachment{}
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "invalid_network_mode",
			Message: "Network mode invalid",
		}
		return
	}

	switch n.NetworkMode6 {
	case Static:
		for _, blckAttch := range n.Blocks6 {
			blck, e := block.Get(db, blckAttch.Block)
			if e != nil {
				err = e
				if _, ok := err.(*database.NotFoundError); ok {
					err = nil
				} else {
					return
				}
			}

			if blck == nil || blck.Type != block.IPv6 {
				errData = &errortypes.ErrorData{
					Error:   "invalid_block6",
					Message: "External IPv6 block invalid",
				}
				return
			}
		}

		break
	case Oracle:
		n.Blocks6 = []*BlockAttachment{}
		break
	case Dhcp, "":
		n.NetworkMode6 = Dhcp
		n.Blocks6 = []*BlockAttachment{}
		break
	case Disabled:
		n.Blocks6 = []*BlockAttachment{}
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "invalid_network_mode6",
			Message: "Network mode6 invalid",
		}
		return
	}

	if n.NetworkMode == Static && n.NetworkMode6 == Static ||
		n.NetworkMode == Disabled && n.NetworkMode6 == Disabled {

		n.ExternalInterfaces = []string{}
	}

	if n.NetworkMode == Oracle || n.NetworkMode6 == Oracle {
		if n.OracleUser == "" {
			errData = &errortypes.ErrorData{
				Error:   "missing_oracle_user",
				Message: "Oracle user OCID required for host routing",
			}
			return
		}
	} else {
		n.OracleUser = ""
		n.OracleSubnets = []string{}
		n.AvailableVpcs = []*cloud.Vpc{}
	}

	n.Format()

	return
}

func (n *Node) Format() {
	sort.Strings(n.Types)
	utils.SortObjectIds(n.Certificates)
}

func (n *Node) JsonHypervisor() {
	vpcs := []*cloud.Vpc{}

	oracleSubnets := set.NewSet()
	for _, subnet := range n.OracleSubnets {
		oracleSubnets.Add(subnet)
	}

	for _, vpc := range n.AvailableVpcs {
		subnets := []*cloud.Subnet{}
		for _, subnet := range vpc.Subnets {
			if oracleSubnets.Contains(subnet.Id) {
				subnets = append(subnets, subnet)
			}
		}

		if len(subnets) > 0 {
			vpcs = append(vpcs, vpc)
		}
	}

	n.AvailableVpcs = vpcs

	return
}

func (n *Node) SetActive() {
	if time.Since(n.Timestamp) > 30*time.Second {
		n.RequestsMin = 0
		n.Memory = 0
		n.Load1 = 0
		n.Load5 = 0
		n.Load15 = 0
		n.CpuUnits = 0
		n.CpuUnitsRes = 0
		n.MemoryUnits = 0
		n.MemoryUnitsRes = 0
	}
}

func (n *Node) Commit(db *database.Database) (err error) {
	coll := db.Nodes()

	err = coll.Commit(n.Id, n)
	if err != nil {
		return
	}

	return
}

func (n *Node) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Nodes()

	err = coll.CommitFields(n.Id, n, fields)
	if err != nil {
		return
	}

	return
}

func (n *Node) GetStaticAddr(db *database.Database,
	instId primitive.ObjectID) (blck *block.Block, ip net.IP, iface string,
	err error) {

	blck, blckIp, err := block.GetInstanceIp(db, instId, block.External)
	if err != nil {
		return
	}

	if blckIp != nil {
		for _, blckAttch := range n.Blocks {
			if blckAttch.Block == blck.Id {
				ip = blckIp.GetIp()
				iface = blckAttch.Interface
				return
			}
		}

		err = block.RemoveIp(db, blckIp.Id)
		if err != nil {
			return
		}
	}

	for _, blckAttch := range n.Blocks {
		blck, err = block.Get(db, blckAttch.Block)
		if err != nil {
			return
		}

		iface = blckAttch.Interface

		ip, err = blck.GetIp(db, instId, block.External)
		if err != nil {
			if _, ok := err.(*block.BlockFull); ok {
				err = nil
				continue
			} else {
				return
			}
		}

		break
	}

	if ip == nil {
		err = &errortypes.NotFoundError{
			errors.New("node: No external block addresses available"),
		}
		return
	}

	return
}

func (n *Node) GetStaticAddr6(db *database.Database,
	instId primitive.ObjectID, vlan int, matchIface string) (
	blck *block.Block, ip net.IP, cidr int, iface string, err error) {

	mismatch := false

	if matchIface != "" {
		for _, blckAttch := range n.Blocks6 {
			if blckAttch.Interface != matchIface {
				mismatch = true
				continue
			}

			blck, err = block.Get(db, blckAttch.Block)
			if err != nil {
				return
			}

			iface = blckAttch.Interface

			ip, cidr, err = blck.GetIp6(db, instId, vlan)
			if err != nil {
				if _, ok := err.(*block.BlockFull); ok {
					err = nil
					continue
				} else {
					return
				}
			}

			break
		}
	} else {
		for _, blckAttch := range n.Blocks6 {
			blck, err = block.Get(db, blckAttch.Block)
			if err != nil {
				return
			}

			iface = blckAttch.Interface

			ip, cidr, err = blck.GetIp6(db, instId, vlan)
			if err != nil {
				if _, ok := err.(*block.BlockFull); ok {
					err = nil
					continue
				} else {
					return
				}
			}

			break
		}
	}

	if ip == nil {
		if mismatch {
			err = &errortypes.NotFoundError{
				errors.New("node: No external block6 with matching " +
					"block interface available"),
			}
		} else {
			err = &errortypes.NotFoundError{
				errors.New("node: No external block6 addresses available"),
			}
		}
		return
	}

	return
}

func (n *Node) GetStaticHostAddr(db *database.Database,
	instId primitive.ObjectID) (blck *block.Block, ip net.IP, err error) {

	blck, err = block.GetNodeBlock(n.Id)
	if err != nil {
		return
	}

	blckIp, err := block.GetInstanceHostIp(db, instId)
	if err != nil {
		return
	}

	if blckIp != nil {
		contains, e := blck.Contains(blckIp)
		if e != nil {
			err = e
			return
		}

		if contains {
			ip = blckIp.GetIp()
			return
		}

		err = block.RemoveIp(db, blckIp.Id)
		if err != nil {
			return
		}
	}

	ip, err = blck.GetIp(db, instId, block.Host)
	if err != nil {
		if _, ok := err.(*block.BlockFull); ok {
			err = nil
		} else {
			return
		}
	}

	if ip == nil {
		err = &errortypes.NotFoundError{
			errors.New("node: No host block addresses available"),
		}
		return
	}

	return
}

func (n *Node) GetRemoteAddr(r *http.Request) (addr string) {
	if n.ForwardedForHeader != "" {
		addr = strings.TrimSpace(
			strings.SplitN(r.Header.Get(n.ForwardedForHeader), ",", 2)[0])
		if addr != "" {
			return
		}
	}

	addr = utils.StripPort(r.RemoteAddr)
	return
}

func (n *Node) update(db *database.Database) (err error) {
	coll := db.Nodes()

	nde := &Node{}
	opts := &options.FindOneAndUpdateOptions{}
	opts.SetReturnDocument(options.After)

	err = coll.FindOneAndUpdate(
		db,
		&bson.M{
			"_id": n.Id,
		},
		&bson.M{
			"$set": &bson.M{
				"timestamp":            n.Timestamp,
				"requests_min":         n.RequestsMin,
				"memory":               n.Memory,
				"hugepages_used":       n.HugePagesUsed,
				"load1":                n.Load1,
				"load5":                n.Load5,
				"load15":               n.Load15,
				"cpu_units":            n.CpuUnits,
				"memory_units":         n.MemoryUnits,
				"cpu_units_res":        n.CpuUnitsRes,
				"memory_units_res":     n.MemoryUnitsRes,
				"public_ips":           n.PublicIps,
				"public_ips6":          n.PublicIps6,
				"private_ips":          n.PrivateIps,
				"hostname":             n.Hostname,
				"local_isos":           n.LocalIsos,
				"usb_devices":          n.UsbDevices,
				"pci_devices":          n.PciDevices,
				"available_renders":    n.AvailableRenders,
				"available_interfaces": n.AvailableInterfaces,
				"available_bridges":    n.AvailableBridges,
				"available_vpcs":       n.AvailableVpcs,
				"default_interface":    n.DefaultInterface,
				"pools":                n.Pools,
				"available_drives":     n.AvailableDrives,
			},
		},
		opts,
	).Decode(nde)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	n.Id = nde.Id
	n.Zone = nde.Zone
	n.Name = nde.Name
	n.Comment = nde.Comment
	n.Types = nde.Types
	n.Port = nde.Port
	n.NoRedirectServer = nde.NoRedirectServer
	n.Protocol = nde.Protocol
	n.Hypervisor = nde.Hypervisor
	n.Vga = nde.Vga
	n.VgaRender = nde.VgaRender
	n.Gui = nde.Gui
	n.GuiUser = nde.GuiUser
	n.GuiMode = nde.GuiMode
	n.Certificates = nde.Certificates
	n.SelfCertificate = nde.SelfCertificate
	n.SelfCertificateKey = nde.SelfCertificateKey
	n.AdminDomain = nde.AdminDomain
	n.UserDomain = nde.UserDomain
	n.WebauthnDomain = nde.WebauthnDomain
	n.ForwardedForHeader = nde.ForwardedForHeader
	n.ForwardedProtoHeader = nde.ForwardedProtoHeader
	n.ExternalInterface = nde.ExternalInterface
	n.InternalInterface = nde.InternalInterface
	n.ExternalInterfaces = nde.ExternalInterfaces
	n.ExternalInterfaces6 = nde.ExternalInterfaces6
	n.InternalInterfaces = nde.InternalInterfaces
	n.OracleSubnets = nde.OracleSubnets
	n.NetworkMode = nde.NetworkMode
	n.NetworkMode6 = nde.NetworkMode6
	n.Blocks = nde.Blocks
	n.Blocks6 = nde.Blocks6
	n.InstanceDrives = nde.InstanceDrives
	n.NoHostNetwork = nde.NoHostNetwork
	n.HostNat = nde.HostNat
	n.DefaultNoPublicAddress = nde.DefaultNoPublicAddress
	n.DefaultNoPublicAddress6 = nde.DefaultNoPublicAddress6
	n.JumboFrames = nde.JumboFrames
	n.JumboFramesInternal = nde.JumboFramesInternal
	n.Iscsi = nde.Iscsi
	n.UsbPassthrough = nde.UsbPassthrough
	n.PciPassthrough = nde.PciPassthrough
	n.Hugepages = nde.Hugepages
	n.HugepagesSize = nde.HugepagesSize
	n.Firewall = nde.Firewall
	n.NetworkRoles = nde.NetworkRoles
	n.VirtPath = nde.VirtPath
	n.CachePath = nde.CachePath
	n.TempPath = nde.TempPath
	n.OracleUser = nde.OracleUser
	n.OraclePrivateKey = nde.OraclePrivateKey
	n.OraclePublicKey = nde.OraclePublicKey
	n.Operation = nde.Operation

	return
}

func (n *Node) sync() {
	db := database.GetDatabase()
	defer db.Close()

	n.Timestamp = time.Now()

	mem, err := utils.GetMemInfo()
	if err != nil {
		n.Memory = 0
		n.HugePagesUsed = 0
		n.MemoryUnits = 0

		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("node: Failed to get memory")
	} else {
		n.Memory = utils.ToFixed(mem.UsedPercent, 2)
		n.HugePagesUsed = utils.ToFixed(mem.HugePagesUsedPercent, 2)
		n.MemoryUnits = utils.ToFixed(
			float64(mem.Total)/float64(1048576), 2)
	}

	load, err := utils.LoadAverage()
	if err != nil {
		n.CpuUnits = 0
		n.Load1 = 0
		n.Load5 = 0
		n.Load15 = 0

		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("node: Failed to get load")
	} else {
		n.CpuUnits = load.CpuUnits
		n.Load1 = load.Load1
		n.Load5 = load.Load5
		n.Load15 = load.Load15
	}

	defaultIface, err := getDefaultIface()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"default_interface": defaultIface,
			"error":             err,
		}).Error("node: Failed to get public address")
	}

	if defaultIface != "" {
		n.DefaultInterface = defaultIface

		pubAddr, pubAddr6, err := bridges.GetIpAddrs(defaultIface)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"default_interface": defaultIface,
				"error":             err,
			}).Error("node: Failed to get public address")
		}

		if pubAddr != "" {
			n.PublicIps = []string{
				pubAddr,
			}
		}

		if pubAddr6 != "" {
			n.PublicIps6 = []string{
				pubAddr6,
			}
		}
	}

	privateIps := map[string]string{}
	internalInterfaces := n.InternalInterfaces
	if internalInterfaces != nil {
		for _, iface := range internalInterfaces {
			addr, _, err := bridges.GetIpAddrs(iface)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"internal_interface": iface,
					"error":              err,
				}).Error("node: Failed to get private address")
			}

			if addr != "" {
				privateIps[iface] = addr
			}
		}
	}
	n.PrivateIps = privateIps

	ifaces, err := GetInterfaces()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("node: Failed to get interfaces")
	}

	if ifaces != nil {
		n.AvailableInterfaces = ifaces
	} else {
		n.AvailableInterfaces = []string{}
	}

	brdgs, err := bridges.GetBridges()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("node: Failed to get bridge interfaces")
	}

	if brdgs != nil {
		n.AvailableBridges = brdgs
	} else {
		n.AvailableBridges = []string{}
	}

	if n.JumboFrames {
		n.JumboFramesInternal = true
	}

	if n.NetworkMode == Oracle || n.NetworkMode6 == Oracle {
		oracleVpcs, e := cloud.GetOracleVpcs(n.GetOracleAuthProvider())
		if e != nil {
			logrus.WithFields(logrus.Fields{
				"error": e,
			}).Error("node: Failed to get oracle vpcs")
		}

		if oracleVpcs != nil {
			n.AvailableVpcs = oracleVpcs
		} else {
			n.AvailableVpcs = []*cloud.Vpc{}
		}
	} else {
		n.AvailableVpcs = []*cloud.Vpc{}
	}

	pools, err := lvm.GetAvailablePools(db, n.Zone)
	if err != nil {
		return
	}

	poolIds := []primitive.ObjectID{}
	for _, pl := range pools {
		poolIds = append(poolIds, pl.Id)
	}
	n.Pools = poolIds

	drives, err := drive.GetDevices()
	if err != nil {
		return
	}
	n.AvailableDrives = drives

	renders, err := render.GetRenders()
	if err != nil {
		return
	}
	n.AvailableRenders = renders

	hostname, err := os.Hostname()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "node: Failed to get hostname"),
		}
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("node: Failed to get hostname")
	}
	n.Hostname = hostname

	isos, err := iso.GetIsos(path.Join(n.GetVirtPath(), "isos"))
	if err != nil {
		return
	}
	n.LocalIsos = isos

	if n.UsbPassthrough {
		devices, e := usb.GetDevices()
		if e != nil {
			err = e
			return
		}

		n.UsbDevices = devices
	} else {
		n.UsbDevices = []*usb.Device{}
	}

	if n.PciPassthrough {
		pciDevices, e := pci.GetVfioAll()
		if err != nil {
			err = e
			return
		}
		n.PciDevices = pciDevices
	} else {
		n.PciDevices = []*pci.Device{}
	}

	err = n.update(db)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("node: Failed to update node")
	}

	if n.Operation == Restart {
		logrus.Info("node: Restarting node")

		n.Operation = ""
		err = n.CommitFields(db, set.NewSet("operation"))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("node: Failed to commit node operation")
		} else {
			cmd := exec.Command("systemctl", "restart", "pritunl-cloud")
			err = cmd.Start()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("node: Failed to start node restart")
			}
		}
	}
}

func (n *Node) keepalive() {
	for {
		if constants.Shutdown {
			return
		}

		n.sync()
		time.Sleep(1 * time.Second)
	}
}

func (n *Node) reqInit() {
	n.reqLock.Lock()
	n.reqCount = list.New()
	for i := 0; i < 60; i++ {
		n.reqCount.PushBack(0)
	}
	n.reqLock.Unlock()
}

func (n *Node) reqSync() {
	for {
		time.Sleep(1 * time.Second)

		if constants.Shutdown {
			return
		}

		n.reqLock.Lock()

		var count int64
		for elm := n.reqCount.Front(); elm != nil; elm = elm.Next() {
			count += int64(elm.Value.(int))
		}
		n.RequestsMin = count

		n.reqCount.Remove(n.reqCount.Front())
		n.reqCount.PushBack(0)

		n.reqLock.Unlock()
	}
}

func (n *Node) Init() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	coll := db.Nodes()

	err = coll.FindOneId(n.Id, n)
	if err != nil {
		switch err.(type) {
		case *database.NotFoundError:
			err = nil
		default:
			return
		}
	}

	n.SoftwareVersion = constants.Version

	if n.Name == "" {
		n.Name = utils.RandName()
	}

	if n.Types == nil {
		n.Types = []string{Admin, Hypervisor}
	}

	if n.Protocol == "" {
		n.Protocol = "https"
	}

	if n.Port == 0 {
		n.Port = 443
	}

	if n.Hypervisor == "" {
		n.Hypervisor = Kvm
	}

	bsonSet := bson.M{
		"_id":              n.Id,
		"name":             n.Name,
		"types":            n.Types,
		"timestamp":        time.Now(),
		"protocol":         n.Protocol,
		"port":             n.Port,
		"hypervisor":       n.Hypervisor,
		"vga":              n.Vga,
		"software_version": n.SoftwareVersion,
	}

	if n.OraclePublicKey == "" || n.OraclePrivateKey == "" {
		privKey, pubKey, e := utils.GenerateRsaKey()
		if e != nil {
			err = e
			return
		}

		bsonSet["oracle_public_key"] = strings.TrimSpace(string(pubKey))
		bsonSet["oracle_private_key"] = strings.TrimSpace(string(privKey))
	}

	// Database upgrade
	if n.InternalInterfaces == nil {
		ifaces := []string{}
		iface := n.InternalInterface
		if iface != "" {
			ifaces = append(ifaces, iface)
		}
		n.InternalInterfaces = ifaces
		bsonSet["internal_interfaces"] = ifaces
	}

	// Database upgrade
	if n.ExternalInterfaces == nil {
		ifaces := []string{}
		iface := n.ExternalInterface
		if iface != "" {
			ifaces = append(ifaces, iface)
		}
		n.ExternalInterfaces = ifaces
		bsonSet["external_interfaces"] = ifaces
	}

	opts := &options.UpdateOptions{}
	opts.SetUpsert(true)

	_, err = coll.UpdateOne(
		db,
		&bson.M{
			"_id": n.Id,
		},
		&bson.M{
			"$set": bsonSet,
		},
		opts,
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	n.reqInit()

	n.sync()

	event.PublishDispatch(db, "node.change")

	Self = n

	go n.keepalive()
	go n.reqSync()

	return
}
