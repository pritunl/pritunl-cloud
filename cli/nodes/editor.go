package nodes

import (
	"sort"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/drive"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/subscription"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/zone"
	"github.com/sirupsen/logrus"
)

// commitFields are the fields written by the admin node handler.
var commitFields = []string{
	"name",
	"comment",
	"zone",
	"types",
	"port",
	"http2",
	"no_redirect_server",
	"protocol",
	"hypervisor",
	"vga",
	"vga_render",
	"gui",
	"gui_user",
	"gui_mode",
	"certificates",
	"admin_domain",
	"user_domain",
	"webauthn_domain",
	"external_interfaces",
	"external_interfaces6",
	"internal_interfaces",
	"cloud_subnets",
	"network_mode",
	"network_mode6",
	"blocks",
	"blocks6",
	"shares",
	"instance_drives",
	"no_host_network",
	"no_node_port_network",
	"host_nat",
	"default_no_public_address",
	"default_no_public_address6",
	"advertise_address",
	"jumbo_frames",
	"jumbo_frames_internal",
	"iscsi",
	"usb_passthrough",
	"pci_passthrough",
	"hugepages",
	"hugepages_size",
	"forwarded_for_header",
	"forwarded_proto_header",
	"firewall",
	"roles",
	"oracle_user",
	"oracle_tenancy",
}

var (
	protocolOptions = []widget.SelectOption{
		{Label: "HTTP", Value: "http"},
		{Label: "HTTPS", Value: "https"},
	}
	networkModeOptions = []widget.SelectOption{
		{Label: "Disabled", Value: node.Disabled},
		{Label: "DHCP", Value: node.Dhcp},
		{Label: "Static", Value: node.Static},
		{Label: "Oracle Cloud", Value: node.Cloud},
	}
	guiModeOptions = []widget.SelectOption{
		{Label: "SDL", Value: node.Sdl},
		{Label: "GTK", Value: node.Gtk},
	}
	hypervisorOptions = []widget.SelectOption{
		{Label: "QEMU", Value: node.Qemu},
		{Label: "KVM", Value: node.Kvm},
	}
	vgaOptions = []widget.SelectOption{
		{Label: "Standard", Value: node.Std},
		{Label: "VMware", Value: node.Vmware},
		{Label: "Virtio", Value: node.Virtio},
		{Label: "Virtio GPU PCI", Value: node.VirtioPci},
		{Label: "Virtio VGA OpenGL", Value: node.VirtioVgaGl},
		{Label: "Virtio VGA Vulkan", Value: node.VirtioVgaGlVulkan},
		{Label: "Virtio GPU OpenGL", Value: node.VirtioGl},
		{Label: "Virtio GPU Vulkan", Value: node.VirtioGlVulkan},
		{Label: "Virtio GPU PCI OpenGL", Value: node.VirtioPciGl},
		{Label: "Virtio GPU PCI Vulkan", Value: node.VirtioPciGlVulkan},
		{Label: "Virtio GPU PCI Prime", Value: node.VirtioPciPrime},
		{Label: "Virtio VGA OpenGL Prime", Value: node.VirtioVgaGlPrime},
		{Label: "Virtio VGA Vulkan Prime", Value: node.VirtioVgaGlVulkanPrime},
		{Label: "Virtio GPU OpenGL Prime", Value: node.VirtioGlPrime},
		{Label: "Virtio GPU Vulkan Prime", Value: node.VirtioGlVulkanPrime},
		{Label: "Virtio GPU PCI OpenGL Prime", Value: node.VirtioPciGlPrime},
		{Label: "Virtio GPU PCI Vulkan Prime",
			Value: node.VirtioPciGlVulkanPrime},
	}
)

// editor is the node settings form mirroring the inputs of the detailed
// node view in the web interface.
type editor struct {
	nde   *node.Node
	names *names
	form  *form.Form

	name    *form.Text
	comment *form.Area

	admin      *form.Toggle
	user       *form.Toggle
	balancer   *form.Toggle
	hypervisor *form.Toggle

	adminDomain    *form.Text
	userDomain     *form.Text
	webauthnDomain *form.Text
	protocol       *form.Select
	port           *form.Number
	http2          *form.Toggle
	redirectServer *form.Toggle
	zone           *form.Select

	networkMode  *form.Select
	networkMode6 *form.Select
	externalIfs  *form.Tokens
	internalIfs  *form.Tokens
	blocks       *form.Rows
	blocks6      *form.Rows
	cloudSubnets *form.Tokens
	hostNetwork  *form.Toggle
	hostNat      *form.Toggle
	nodePortNet  *form.Toggle
	oracleUser   *form.Text
	oracleTenant *form.Text
	publicIp4    *form.Toggle
	publicIp6    *form.Toggle
	jumboExt     *form.Toggle
	jumboInt     *form.Toggle
	iscsi        *form.Toggle
	pci          *form.Toggle
	advertise    *form.Text
	usb          *form.Toggle
	hugepages    *form.Toggle
	hugepagesSz  *form.Number
	firewall     *form.Toggle
	gui          *form.Toggle
	guiMode      *form.Select
	guiUser      *form.Text

	hypervisorMode *form.Select
	vga            *form.Select
	vgaRender      *form.Select
	drives         *form.Tokens
	shares         *form.Rows
	roles          *form.Tokens
	certificates   *form.Tokens
}

func hasType(types []string, typ string) bool {
	for _, cur := range types {
		if cur == typ {
			return true
		}
	}
	return false
}

// interfaceOptions lists the available interfaces and bridges of the
// node with their address.
func interfaceOptions(nde *node.Node) []widget.SelectOption {
	opts := []widget.SelectOption{}
	seen := map[string]bool{}

	add := func(name, address string) {
		if seen[name] {
			return
		}
		seen[name] = true
		label := name
		if address != "" {
			label += " (" + address + ")"
		}
		opts = append(opts, widget.SelectOption{
			Label: label,
			Value: name,
		})
	}

	for _, iface := range nde.AvailableInterfaces {
		add(iface.Name, iface.Address)
	}
	for _, iface := range nde.AvailableBridges {
		add(iface.Name, iface.Address)
	}

	return opts
}

func subnetOptions(nde *node.Node) []widget.SelectOption {
	opts := []widget.SelectOption{}
	for _, vc := range nde.AvailableVpcs {
		for _, subnet := range vc.Subnets {
			opts = append(opts, widget.SelectOption{
				Label: vc.Name + " - " + subnet.Name,
				Value: subnet.Id,
			})
		}
	}
	return opts
}

func driveOptions(nde *node.Node) []widget.SelectOption {
	opts := []widget.SelectOption{}
	for _, device := range nde.AvailableDrives {
		opts = append(opts, widget.SelectOption{
			Label: device.Id,
			Value: device.Id,
		})
	}
	return opts
}

func renderOptions(nde *node.Node) []widget.SelectOption {
	opts := []widget.SelectOption{}
	for _, render := range nde.AvailableRenders {
		opts = append(opts, widget.SelectOption{
			Label: render,
			Value: render,
		})
	}
	return opts
}

func driveIds(devices []*drive.Device) []string {
	ids := []string{}
	for _, device := range devices {
		ids = append(ids, device.Id)
	}
	return ids
}

func objectIdHexes(ids []bson.ObjectID) []string {
	hexes := []string{}
	for _, id := range ids {
		hexes = append(hexes, id.Hex())
	}
	return hexes
}

// blockRow creates the interface and block selects of a block
// attachment.
func blockRow(ifaces, blocks []widget.SelectOption,
	attachment *node.BlockAttachment) *form.Row {

	iface := form.NewSelect("", ifaces, attachment.Interface)
	iface.Compact = true

	blck := form.NewSelect("", blocks, attachment.Block.Hex())
	blck.Compact = true
	if attachment.Block.IsZero() {
		blck = form.NewSelect("", blocks, "")
		blck.Compact = true
	}

	return &form.Row{
		Cells: []form.Input{iface, blck},
		Tag:   attachment,
	}
}

// shareRow creates the path and roles inputs of a share.
func shareRow(share *node.Share) *form.Row {
	return &form.Row{
		Cells: []form.Input{
			form.NewCell("Enter share path", share.Path),
			form.NewTokens("", "Add role", share.Roles),
		},
		Tag: share,
	}
}

func newEditor(nde *node.Node, nms *names) *editor {
	e := &editor{
		nde:   nde,
		names: nms,
	}

	e.name = form.NewText("Name", "Enter name", nde.Name)
	e.comment = form.NewArea("Comment", "Node comment", nde.Comment, 4)

	e.admin = form.NewToggle("Admin", hasType(nde.Types, node.Admin))
	e.user = form.NewToggle("User", hasType(nde.Types, node.User))
	e.balancer = form.NewToggle("Load Balancer",
		hasType(nde.Types, node.Balancer))
	e.hypervisor = form.NewToggle("Hypervisor",
		hasType(nde.Types, node.Hypervisor))

	// Domains are needed to route the admin and user consoles on a
	// balancer or a node serving both
	e.adminDomain = form.NewText("Admin Domain", "Enter admin domain",
		nde.AdminDomain)
	e.adminDomain.Hide = func() bool {
		return !e.balancer.Value() && !(e.admin.Value() && e.user.Value())
	}
	e.userDomain = form.NewText("User Domain", "Enter user domain",
		nde.UserDomain)
	e.userDomain.Hide = e.adminDomain.Hide
	e.webauthnDomain = form.NewText("WebAuthn Domain",
		"Enter WebAuthn domain", nde.WebauthnDomain)
	e.webauthnDomain.Hide = func() bool {
		return !e.admin.Value() && !e.user.Value()
	}

	e.protocol = form.NewSelect("Protocol", protocolOptions,
		widget.Default(nde.Protocol, "https"))
	e.port = form.NewNumber("Port", "443", nde.Port, 1, 65535, 1)
	e.http2 = form.NewToggle("Web server HTTP/2 support", nde.Http2)
	e.redirectServer = form.NewToggle("Web redirect server",
		!nde.NoRedirectServer)

	// Zone can only be set once
	e.zone = form.NewSelect("Zone", append(
		[]widget.SelectOption{{Label: "Select Zone", Value: ""}},
		resource.SelectOptions(nms.zones)...), "")
	e.zone.Hide = func() bool {
		return !nde.Zone.IsZero()
	}

	e.networkMode = form.NewSelect("Network IPv4 Mode", networkModeOptions,
		widget.Default(nde.NetworkMode, node.Disabled))
	e.networkMode6 = form.NewSelect("Network IPv6 Mode",
		networkModeOptions, widget.Default(nde.NetworkMode6, node.Disabled))

	ifaces := interfaceOptions(nde)
	e.externalIfs = form.NewTokensSelect("External Interfaces", ifaces,
		nde.ExternalInterfaces)
	e.externalIfs.Hide = func() bool {
		mode := e.networkMode.Value()
		mode6 := e.networkMode6.Value()
		return mode != node.Dhcp && mode != "" &&
			mode6 != node.Dhcp && mode6 != ""
	}
	e.internalIfs = form.NewTokensSelect("Internal Interfaces", ifaces,
		nde.InternalInterfaces)

	blocks := resource.SelectOptions(nms.blocks)
	e.blocks = form.NewRows("External IPv4 Block Attachments",
		"No block attachments", "Add Block Attachment",
		[]string{"Interface", "Block"}, func() *form.Row {
			row := blockRow(ifaces, blocks, &node.BlockAttachment{})
			row.Tag = nil
			return row
		})
	for _, attachment := range nde.Blocks {
		e.blocks.Add(blockRow(ifaces, blocks, attachment))
	}
	e.blocks.Hide = func() bool {
		return e.networkMode.Value() != node.Static
	}

	e.blocks6 = form.NewRows("External IPv6 Block Attachments",
		"No block attachments", "Add Block Attachment",
		[]string{"Interface", "Block"}, func() *form.Row {
			row := blockRow(ifaces, blocks, &node.BlockAttachment{})
			row.Tag = nil
			return row
		})
	for _, attachment := range nde.Blocks6 {
		e.blocks6.Add(blockRow(ifaces, blocks, attachment))
	}
	e.blocks6.Hide = func() bool {
		return e.networkMode6.Value() != node.Static
	}

	e.cloudSubnets = form.NewTokensSelect("Oracle Cloud Subnets",
		subnetOptions(nde), nde.CloudSubnets)
	e.cloudSubnets.Hide = func() bool {
		return e.networkMode.Value() != node.Cloud
	}

	e.hostNetwork = form.NewToggle("Host Network", !nde.NoHostNetwork)
	e.hostNat = form.NewToggle("Host Network NAT", nde.HostNat)
	e.hostNat.Hide = func() bool {
		return !e.hostNetwork.Value()
	}
	e.nodePortNet = form.NewToggle("Node Port Network",
		!nde.NoNodePortNetwork)

	e.oracleUser = form.NewText("Oracle Cloud User OCID",
		"Enter user OCID", nde.OracleUser)
	e.oracleUser.Hide = func() bool {
		return e.networkMode.Value() != node.Cloud &&
			e.networkMode6.Value() != node.Cloud
	}
	e.oracleTenant = form.NewText("Oracle Cloud User Tenancy",
		"Enter user tenancy", nde.OracleTenancy)
	e.oracleTenant.Hide = e.oracleUser.Hide

	e.publicIp4 = form.NewToggle("Default instance public IPv4 address",
		!nde.DefaultNoPublicAddress)
	e.publicIp6 = form.NewToggle("Default instance public IPv6 address",
		!nde.DefaultNoPublicAddress6)
	e.jumboExt = form.NewToggle("Jumbo frames external", nde.JumboFrames)
	e.jumboInt = form.NewToggle("Jumbo frames internal",
		nde.JumboFramesInternal)
	e.iscsi = form.NewToggle("Instance iSCSI support", nde.Iscsi)
	e.pci = form.NewToggle("PCI Passthrough", nde.PciPassthrough)
	e.advertise = form.NewText("Advertise Address",
		"Enter advertise address", nde.AdvertiseAddress)
	e.usb = form.NewToggle("USB Passthrough", nde.UsbPassthrough)
	e.hugepages = form.NewToggle("HugePages", nde.Hugepages)
	e.hugepagesSz = form.NewNumber("HugePages Size", "Size in megabytes",
		nde.HugepagesSize, 0, 0, 1024)
	e.hugepagesSz.Hide = func() bool {
		return !e.hugepages.Value()
	}
	e.firewall = form.NewToggle("Firewall", nde.Firewall)

	e.gui = form.NewToggle("Desktop GUI", nde.Gui)
	e.guiMode = form.NewSelect("Desktop GUI Mode", guiModeOptions,
		widget.Default(nde.GuiMode, node.Sdl))
	e.guiMode.Hide = func() bool {
		return !e.gui.Value()
	}
	e.guiUser = form.NewText("Desktop GUI User", "Enter GUI user",
		nde.GuiUser)
	e.guiUser.Hide = e.guiMode.Hide

	e.hypervisorMode = form.NewSelect("Hypervisor Mode", hypervisorOptions,
		widget.Default(nde.Hypervisor, node.Kvm))
	e.hypervisorMode.Hide = func() bool {
		return !e.hypervisor.Value()
	}
	e.vga = form.NewSelect("Hypervisor VGA Type", vgaOptions,
		widget.Default(nde.Vga, node.Virtio))
	e.vga.Hide = e.hypervisorMode.Hide
	e.vgaRender = form.NewSelect("Hypervisor EGL Render",
		renderOptions(nde), nde.VgaRender)
	e.vgaRender.Hide = func() bool {
		return !e.hypervisor.Value() ||
			!node.VgaRenderModes.Contains(e.vga.Value())
	}

	e.drives = form.NewTokensSelect("Instance Passthrough Disks",
		driveOptions(nde), driveIds(nde.InstanceDrives))

	e.shares = form.NewRows("Share Paths", "No share paths",
		"Add Share Path", []string{"Path", "Roles"}, func() *form.Row {
			row := shareRow(&node.Share{Type: node.HostPath})
			row.Tag = nil
			return row
		})
	for _, share := range nde.Shares {
		e.shares.Add(shareRow(share))
	}

	e.roles = form.NewTokens("Roles", "Add role", nde.Roles)

	e.certificates = form.NewTokensSelect("Certificates",
		resource.SelectOptions(nms.certificates),
		objectIdHexes(nde.Certificates))
	e.certificates.Hide = func() bool {
		return e.protocol.Value() == "http"
	}

	e.form = form.New(
		e.name,
		e.comment,
		e.admin,
		e.user,
		e.balancer,
		e.hypervisor,
		e.adminDomain,
		e.userDomain,
		e.webauthnDomain,
		e.protocol,
		e.port,
		e.http2,
		e.redirectServer,
		e.zone,
		e.networkMode,
		e.networkMode6,
		e.externalIfs,
		e.internalIfs,
		e.blocks,
		e.blocks6,
		e.cloudSubnets,
		e.hostNetwork,
		e.hostNat,
		e.nodePortNet,
		e.oracleUser,
		e.oracleTenant,
		e.publicIp4,
		e.publicIp6,
		e.jumboExt,
		e.jumboInt,
		e.iscsi,
		e.pci,
		e.advertise,
		e.usb,
		e.hugepages,
		e.hugepagesSz,
		e.firewall,
		e.gui,
		e.guiMode,
		e.guiUser,
		e.hypervisorMode,
		e.vga,
		e.vgaRender,
		e.drives,
		e.shares,
		e.roles,
		e.certificates,
	)

	return e
}

func (e *editor) Form() *form.Form {
	return e.form
}

func (e *editor) buildTypes() []string {
	types := []string{}
	if e.admin.Value() {
		types = append(types, node.Admin)
	}
	if e.user.Value() {
		types = append(types, node.User)
	}
	if e.balancer.Value() {
		types = append(types, node.Balancer)
	}
	if e.hypervisor.Value() {
		types = append(types, node.Hypervisor)
	}
	sort.Strings(types)
	return types
}

func buildBlocks(rows *form.Rows) []*node.BlockAttachment {
	attachments := []*node.BlockAttachment{}
	for _, row := range rows.Rows() {
		blockId, _ := utils.ParseObjectId(row.Cells[1].(*form.Select).Value())
		attachments = append(attachments, &node.BlockAttachment{
			Interface: row.Cells[0].(*form.Select).Value(),
			Block:     blockId,
		})
	}
	return attachments
}

func (e *editor) buildShares() []*node.Share {
	shares := []*node.Share{}
	for _, row := range e.shares.Rows() {
		shares = append(shares, &node.Share{
			Type:  node.HostPath,
			Path:  row.Cells[0].(*form.Text).Value(),
			Roles: row.Cells[1].(*form.Tokens).Values(),
		})
	}
	return shares
}

func (e *editor) buildDrives() []*drive.Device {
	devices := []*drive.Device{}
	for _, id := range e.drives.Values() {
		devices = append(devices, &drive.Device{
			Id: id,
		})
	}
	return devices
}

func (e *editor) buildCertificates() []bson.ObjectID {
	certs := []bson.ObjectID{}
	for _, hex := range e.certificates.Values() {
		certId, ok := utils.ParseObjectId(hex)
		if ok {
			certs = append(certs, certId)
		}
	}
	return certs
}

// Save applies the form to the current node the same way as the admin
// node handler.
func (e *editor) Save(db *database.Database) (err error) {
	nde, err := node.Get(db, e.nde.Id)
	if err != nil {
		return
	}

	types := e.buildTypes()

	// Multi-tenant user nodes require a subscription
	if hasType(types, node.User) && !hasType(nde.Types, node.User) &&
		!subscription.Sub.Active {

		err = &errortypes.ParseError{
			errors.New("nodes: Subscription required for multi-tenant"),
		}
		return
	}

	nde.Name = e.name.Value()
	nde.Comment = e.comment.Value()
	nde.Types = types
	nde.Port = e.port.Int()
	nde.Http2 = e.http2.Value()
	nde.NoRedirectServer = !e.redirectServer.Value()
	nde.Protocol = e.protocol.Value()
	nde.Hypervisor = e.hypervisorMode.Value()
	nde.Vga = e.vga.Value()
	nde.VgaRender = e.vgaRender.Value()
	nde.Gui = e.gui.Value()
	nde.GuiUser = e.guiUser.Value()
	nde.GuiMode = e.guiMode.Value()
	nde.Certificates = e.buildCertificates()
	nde.AdminDomain = e.adminDomain.Value()
	nde.UserDomain = e.userDomain.Value()
	nde.WebauthnDomain = e.webauthnDomain.Value()
	nde.ExternalInterfaces = e.externalIfs.Values()
	nde.InternalInterfaces = e.internalIfs.Values()
	nde.CloudSubnets = e.cloudSubnets.Values()
	nde.NetworkMode = e.networkMode.Value()
	nde.NetworkMode6 = e.networkMode6.Value()
	nde.Blocks = buildBlocks(e.blocks)
	nde.Blocks6 = buildBlocks(e.blocks6)
	nde.Shares = e.buildShares()
	nde.InstanceDrives = e.buildDrives()
	nde.NoHostNetwork = !e.hostNetwork.Value()
	nde.NoNodePortNetwork = !e.nodePortNet.Value()
	nde.HostNat = e.hostNat.Value()
	nde.DefaultNoPublicAddress = !e.publicIp4.Value()
	nde.DefaultNoPublicAddress6 = !e.publicIp6.Value()
	nde.AdvertiseAddress = e.advertise.Value()
	nde.JumboFrames = e.jumboExt.Value()
	nde.JumboFramesInternal = e.jumboInt.Value()
	nde.Iscsi = e.iscsi.Value()
	nde.UsbPassthrough = e.usb.Value()
	nde.PciPassthrough = e.pci.Value()
	nde.Hugepages = e.hugepages.Value()
	nde.HugepagesSize = e.hugepagesSz.Int()
	nde.Firewall = e.firewall.Value()
	nde.Roles = e.roles.Values()
	nde.OracleUser = e.oracleUser.Value()
	nde.OracleTenancy = e.oracleTenant.Value()

	fields := set.NewSet()
	for _, field := range commitFields {
		fields.Add(field)
	}

	// Zone can only be set once, the datacenter follows the zone
	zoneId, ok := utils.ParseObjectId(e.zone.Value())
	if ok && zoneId != nde.Zone {
		if !nde.Zone.IsZero() {
			err = &errortypes.ParseError{
				errors.New("nodes: Cannot modify zone once set"),
			}
			return
		}

		zne, e := zone.Get(db, zoneId)
		if e != nil {
			err = e
			return
		}

		nde.Zone = zoneId
		nde.Datacenter = zne.Datacenter
		fields.Add("datacenter")
	}

	errData, err := nde.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = &errortypes.ParseError{
			errors.New("nodes: " + errData.Message),
		}
		return
	}

	err = nde.CommitFields(db, fields)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"node_id": nde.Id.Hex(),
	}).Info("tui: Node settings saved")

	err = event.PublishDispatch(db, "node.change")
	if err != nil {
		return
	}

	return
}
