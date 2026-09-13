package instances

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/nodeport"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/sirupsen/logrus"
)

// commitFields are the fields written by the admin instance handler.
var commitFields = set.NewSet(
	"unix_id",
	"name",
	"comment",
	"datacenter",
	"vpc",
	"subnet",
	"dhcp_ip",
	"dhcp_ip6",
	"cloud_subnet",
	"state",
	"restart",
	"restart_block_ip",
	"uefi",
	"secure_boot",
	"tpm",
	"tpm_secret",
	"dhcp_server",
	"cloud_type",
	"cloud_script",
	"cloud_interface",
	"delete_protection",
	"skip_source_dest_check",
	"memory",
	"processors",
	"roles",
	"isos",
	"usb_devices",
	"pci_devices",
	"drive_devices",
	"iscsi_devices",
	"mounts",
	"root_enabled",
	"root_passwd",
	"vnc",
	"vnc_display",
	"vnc_password",
	"spice",
	"spice_port",
	"spice_password",
	"gui",
	"node_ports",
	"no_public_address",
	"no_public_address6",
	"no_host_address",
)

var (
	cloudTypeOptions = []widget.SelectOption{
		{Label: "Linux", Value: instance.Linux},
		{Label: "BSD", Value: instance.BSD},
		{Label: "Linux Legacy", Value: instance.LinuxLegacy},
	}
	cloudInterfaceOptions = []widget.SelectOption{
		{Label: "Default", Value: instance.Default},
		{Label: "Bridge", Value: instance.Bridge},
	}
	protocolOptions = []widget.SelectOption{
		{Label: "TCP", Value: nodeport.Tcp},
		{Label: "UDP", Value: nodeport.Udp},
	}
)

// editor is the instance settings form mirroring the inputs of the
// detailed instance view in the web interface.
type editor struct {
	inst *instance.Instance
	vpcs []*vpc.Vpc
	form *form.Form

	name           *form.Text
	comment        *form.Area
	memory         *form.Number
	processors     *form.Number
	roles          *form.Tokens
	vpc            *form.Select
	subnet         *form.Select
	cloudType      *form.Select
	cloudInterface *form.Select
	startup        *form.Toggle
	script         *form.Area
	uefi           *form.Toggle
	secureBoot     *form.Toggle
	tpm            *form.Toggle
	publicIp4      *form.Toggle
	publicIp6      *form.Toggle
	hostAddr       *form.Toggle
	skipSrcDst     *form.Toggle
	dhcp           *form.Toggle
	deleteProt     *form.Toggle
	nodePorts      *form.Rows
	mounts         *form.Rows
	vnc            *form.Toggle
	spice          *form.Toggle
	gui            *form.Toggle
	root           *form.Toggle
}

// nodePortRow creates the protocol, external port and internal port
// inputs of a node port.
func nodePortRow(mapping *nodeport.Mapping) *form.Row {
	protocol := form.NewSelect("", protocolOptions, mapping.Protocol)
	protocol.Compact = true

	return &form.Row{
		Cells: []form.Input{
			protocol,
			form.NewNumberCell("External Port", mapping.ExternalPort),
			form.NewNumberCell("Internal Port", mapping.InternalPort),
		},
		Tag: mapping,
	}
}

func newNodePortRow() *form.Row {
	row := nodePortRow(&nodeport.Mapping{
		Protocol: nodeport.Tcp,
	})
	row.Tag = nil
	return row
}

// mountRow creates the host path, name and instance path inputs of a
// host path mount.
func mountRow(mount *instance.Mount) *form.Row {
	return &form.Row{
		Cells: []form.Input{
			form.NewCell("Enter host path", mount.Path),
			form.NewCell("Enter name", mount.Name),
			form.NewCell("Enter instance path", mount.HostPath),
		},
		Tag: mount,
	}
}

func newMountRow() *form.Row {
	row := mountRow(&instance.Mount{
		Type: instance.HostPath,
	})
	row.Tag = nil
	return row
}

func newEditor(inst *instance.Instance, vpcs []*vpc.Vpc) *editor {
	e := &editor{
		inst: inst,
		vpcs: vpcs,
	}

	e.name = form.NewText("Name", "Enter name", inst.Name)
	e.comment = form.NewArea("Comment", "Instance comment", inst.Comment, 4)
	e.memory = form.NewNumber("Memory Size (MB)", "Memory in megabytes",
		inst.Memory, 256, 0, 1024)
	e.processors = form.NewNumber("Processors", "Number of processors",
		inst.Processors, 1, 0, 1)
	e.roles = form.NewTokens("Roles", "Add role", inst.Roles)

	vpcOpts := []widget.SelectOption{}
	for _, vc := range vpcs {
		if !inst.Organization.IsZero() &&
			vc.Organization != inst.Organization {

			continue
		}
		vpcOpts = append(vpcOpts, widget.SelectOption{
			Label: vc.Name,
			Value: vc.Id.Hex(),
		})
	}
	e.vpc = form.NewSelect("VPC", vpcOpts, inst.Vpc.Hex())

	e.subnet = form.NewSelect("Subnet", nil, inst.Subnet.Hex())
	e.subnet.Dynamic = func() []widget.SelectOption {
		opts := []widget.SelectOption{}
		vpcId := e.vpc.Value()
		for _, vc := range e.vpcs {
			if vc.Id.Hex() != vpcId {
				continue
			}
			for _, sub := range vc.Subnets {
				opts = append(opts, widget.SelectOption{
					Label: sub.Name + " - " + sub.Network,
					Value: sub.Id.Hex(),
				})
			}
		}
		return opts
	}

	e.cloudType = form.NewSelect("CloudInit Type", cloudTypeOptions,
		inst.CloudType)
	e.cloudInterface = form.NewSelect("Guest Interface Mode",
		cloudInterfaceOptions, inst.CloudInterface)

	e.startup = form.NewToggle("Startup script", inst.CloudScript != "")
	e.script = form.NewArea("Startup Script", "Startup script",
		inst.CloudScript, 6)
	e.script.Hide = func() bool {
		return !e.startup.Value()
	}

	e.uefi = form.NewToggle("UEFI", inst.Uefi)
	e.secureBoot = form.NewToggle("SecureBoot", inst.SecureBoot)
	e.secureBoot.Hide = func() bool {
		return !e.uefi.Value()
	}
	e.tpm = form.NewToggle("TPM", inst.Tpm)
	e.tpm.Hide = e.secureBoot.Hide

	e.publicIp4 = form.NewToggle("Public IPv4 address",
		!inst.NoPublicAddress)
	e.publicIp6 = form.NewToggle("Public IPv6 address",
		!inst.NoPublicAddress6)
	e.hostAddr = form.NewToggle("Host address", !inst.NoHostAddress)
	e.skipSrcDst = form.NewToggle("Skip source/destination check",
		inst.SkipSourceDestCheck)
	e.dhcp = form.NewToggle("DHCP server", inst.DhcpServer)
	e.deleteProt = form.NewToggle("Delete protection",
		inst.DeleteProtection)

	e.nodePorts = form.NewRows("Node Ports", "No node ports",
		"Add Node Port",
		[]string{"Protocol", "External Port", "Internal Port"},
		newNodePortRow)
	for _, mapping := range inst.NodePorts {
		e.nodePorts.Add(nodePortRow(mapping))
	}

	e.mounts = form.NewRows("Host Paths", "No host paths",
		"Add Host Path", []string{"Host Path", "Name", "Instance Path"},
		newMountRow)
	for _, mount := range inst.Mounts {
		e.mounts.Add(mountRow(mount))
	}

	e.vnc = form.NewToggle("VNC server", inst.Vnc)
	e.spice = form.NewToggle("Spice server", inst.Spice)
	e.gui = form.NewToggle("Desktop GUI", inst.Gui)
	e.root = form.NewToggle("Root enabled", inst.RootEnabled)

	e.form = form.New(
		e.name,
		e.comment,
		e.memory,
		e.processors,
		e.roles,
		e.vpc,
		e.subnet,
		e.cloudType,
		e.cloudInterface,
		e.startup,
		e.script,
		e.uefi,
		e.secureBoot,
		e.tpm,
		e.publicIp4,
		e.publicIp6,
		e.hostAddr,
		e.skipSrcDst,
		e.dhcp,
		e.deleteProt,
		e.nodePorts,
		e.mounts,
		e.vnc,
		e.spice,
		e.gui,
		e.root,
	)

	return e
}

func (e *editor) Form() *form.Form {
	return e.form
}

// buildNodePorts converts the rows to mappings, removed mappings are
// sent with the delete flag so the node port is released unless the
// external port is reused, matching the web interface.
func (e *editor) buildNodePorts() []*nodeport.Mapping {
	mappings := []*nodeport.Mapping{}
	ports := map[int]bool{}

	for _, row := range e.nodePorts.Rows() {
		mapping := &nodeport.Mapping{
			Protocol:     row.Cells[0].(*form.Select).Value(),
			ExternalPort: row.Cells[1].(*form.Number).Int(),
			InternalPort: row.Cells[2].(*form.Number).Int(),
		}

		// Keep the node port of unchanged external ports, a changed
		// external port allocates a new node port
		if orig, ok := row.Tag.(*nodeport.Mapping); ok &&
			orig.ExternalPort == mapping.ExternalPort {

			mapping.NodePort = orig.NodePort
		}

		mappings = append(mappings, mapping)
		ports[mapping.ExternalPort] = true
	}

	for _, tag := range e.nodePorts.Removed() {
		orig, ok := tag.(*nodeport.Mapping)
		if !ok || ports[orig.ExternalPort] {
			continue
		}

		removed := *orig
		removed.Delete = true
		mappings = append(mappings, &removed)
		ports[orig.ExternalPort] = true
	}

	return mappings
}

func (e *editor) buildMounts() []*instance.Mount {
	mounts := []*instance.Mount{}
	for _, row := range e.mounts.Rows() {
		mounts = append(mounts, &instance.Mount{
			Type:     instance.HostPath,
			Path:     row.Cells[0].(*form.Text).Value(),
			Name:     row.Cells[1].(*form.Text).Value(),
			HostPath: row.Cells[2].(*form.Text).Value(),
		})
	}
	return mounts
}

// Save applies the form to the current instance the same way as the
// admin instance handler.
func (e *editor) Save(db *database.Database) (err error) {
	inst, err := instance.Get(db, e.inst.Id)
	if err != nil {
		return
	}

	inst.PreCommit()

	inst.Name = e.name.Value()
	inst.Comment = e.comment.Value()
	inst.Memory = e.memory.Int()
	inst.Processors = e.processors.Int()
	inst.Roles = e.roles.Values()
	inst.Vpc, _ = utils.ParseObjectId(e.vpc.Value())
	inst.Subnet, _ = utils.ParseObjectId(e.subnet.Value())
	inst.CloudType = e.cloudType.Value()
	inst.CloudInterface = e.cloudInterface.Value()
	if e.startup.Value() {
		inst.CloudScript = e.script.Value()
	} else {
		inst.CloudScript = ""
	}
	inst.Uefi = e.uefi.Value()
	inst.SecureBoot = e.secureBoot.Value()
	inst.Tpm = e.tpm.Value()
	inst.NoPublicAddress = !e.publicIp4.Value()
	inst.NoPublicAddress6 = !e.publicIp6.Value()
	inst.NoHostAddress = !e.hostAddr.Value()
	inst.SkipSourceDestCheck = e.skipSrcDst.Value()
	inst.DhcpServer = e.dhcp.Value()
	inst.DeleteProtection = e.deleteProt.Value()
	inst.NodePorts = e.buildNodePorts()
	inst.Mounts = e.buildMounts()
	inst.Vnc = e.vnc.Value()
	inst.Spice = e.spice.Value()
	inst.Gui = e.gui.Value()
	inst.RootEnabled = e.root.Value()

	errData, err := inst.Validate(db)
	if err != nil {
		if _, ok := err.(*errortypes.NotFoundError); ok {
			err = &errortypes.NotFoundError{
				errors.New("instances: Failed to find available node"),
			}
		}
		return
	}
	if errData != nil {
		err = &errortypes.ParseError{
			errors.New("instances: " + errData.Message),
		}
		return
	}

	dskChange, err := inst.PostCommit(db)
	if err != nil {
		return
	}

	err = inst.CommitFields(db, commitFields)
	if err != nil {
		_ = inst.Cleanup(db)
		return
	}

	err = inst.Cleanup(db)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"instance_id": inst.Id.Hex(),
	}).Info("tui: Instance settings saved")

	err = event.PublishDispatch(db, "instance.change")
	if err != nil {
		return
	}

	if dskChange {
		err = event.PublishDispatch(db, "disk.change")
		if err != nil {
			return
		}
	}

	return
}

// Editor returns a new settings form for the instance.
func (i *Item) Editor() resource.Editor {
	return newEditor(i.inst, i.vpcs)
}
