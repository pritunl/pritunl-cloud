package instances

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/nodeport"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
	"github.com/sirupsen/logrus"
)

var diskTypeOptions = []widget.SelectOption{
	{Label: "QCOW", Value: disk.Qcow2},
	{Label: "LVM", Value: disk.Lvm},
}

// createData are the select options of the new instance form loaded
// once when the form opens.
type createData struct {
	orgs        []*database.Named
	datacenters []*datacenter.Datacenter
	zones       []*zone.Zone
	shapes      []*shape.Shape
	nodes       []*node.Node
	vpcs        []*vpc.Vpc
	pools       []*pool.Pool

	// images are keyed by datacenter, the image name query only returns
	// name, key, signed and firmware so the datacenter storages are part
	// of the query like the admin images handler
	images map[bson.ObjectID][]*image.Image
}

// loadDatacenterImages returns the images on the public and private
// storages of the datacenter.
func loadDatacenterImages(db *database.Database,
	dc *datacenter.Datacenter) (images []*image.Image, err error) {

	images = []*image.Image{}

	if len(dc.PublicStorages) > 0 {
		images, err = image.GetAllNames(db, &bson.M{
			"storage": &bson.M{
				"$in": dc.PublicStorages,
			},
		})
		if err != nil {
			return
		}
	}

	if !dc.PrivateStorage.IsZero() {
		private, e := image.GetAllNames(db, &bson.M{
			"storage": dc.PrivateStorage,
		})
		if e != nil {
			err = e
			return
		}
		images = append(images, private...)
	}

	return
}

func loadCreateData(db *database.Database) (dta *createData, err error) {
	dta = &createData{}

	dta.orgs, err = resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}

	dta.datacenters, err = datacenter.GetAll(db)
	if err != nil {
		return
	}

	dta.zones, err = zone.GetAll(db)
	if err != nil {
		return
	}

	dta.shapes, err = shape.GetAllNames(db, &bson.M{})
	if err != nil {
		return
	}

	dta.nodes, err = node.GetAllHypervisors(db, &bson.M{})
	if err != nil {
		return
	}

	dta.vpcs, err = vpc.GetAllNames(db, &bson.M{})
	if err != nil {
		return
	}

	dta.pools, err = pool.GetAllNames(db, &bson.M{})
	if err != nil {
		return
	}

	dta.images = map[bson.ObjectID][]*image.Image{}
	for _, dc := range dta.datacenters {
		dta.images[dc.Id], err = loadDatacenterImages(db, dc)
		if err != nil {
			return
		}
	}

	return
}

// creator is the new instance form mirroring the new instance panel of
// the web interface, Save inserts the instances.
type creator struct {
	data *createData
	form *form.Form

	name           *form.Text
	comment        *form.Area
	organization   *form.Select
	datacenter     *form.Select
	zone           *form.Select
	shape          *form.Select
	node           *form.Select
	vpc            *form.Select
	subnet         *form.Select
	cloudType      *form.Select
	cloudInterface *form.Select
	startup        *form.Toggle
	script         *form.Area
	start          *form.Toggle
	uefi           *form.Toggle
	secureBoot     *form.Toggle
	tpm            *form.Toggle
	publicIp4      *form.Toggle
	publicIp6      *form.Toggle
	hostAddr       *form.Toggle
	diskType       *form.Select
	diskPool       *form.Select
	hiddenImages   *form.Toggle
	imageBacking   *form.Toggle
	image          *form.Select
	roles          *form.Tokens
	diskSize       *form.Number

	// secureBootChanged is set once the user flips secure boot so an
	// image selection no longer resets it, like the web interface
	secureBootChanged bool
	memory            *form.Number
	processors        *form.Number
	count             *form.Number
	nodePorts         *form.Rows
	mounts            *form.Rows
}

func withSelect(label string, opts []widget.SelectOption) []widget.SelectOption {
	return append([]widget.SelectOption{{Label: label, Value: ""}}, opts...)
}

// imageLabel returns the image name, signed image names already carry
// the build date from the key.
func imageLabel(img *image.Image) string {
	return widget.Default(img.Name, img.Key)
}

func newCreator(dta *createData) *creator {
	c := &creator{
		data: dta,
	}

	c.name = form.NewText("Name", "Enter name", "new-instance")
	c.comment = form.NewArea("Comment", "Instance comment", "", 3)

	c.organization = form.NewSelect("Organization",
		withSelect("Select Organization", resource.SelectOptions(dta.orgs)),
		"")

	dcOpts := []widget.SelectOption{}
	for _, dc := range dta.datacenters {
		dcOpts = append(dcOpts, widget.SelectOption{
			Label: dc.Name,
			Value: dc.Id.Hex(),
		})
	}
	c.datacenter = form.NewSelect("Datacenter",
		withSelect("Select Datacenter", dcOpts), "")

	c.zone = form.NewSelect("Zone", nil, "")
	c.zone.Dynamic = func() []widget.SelectOption {
		opts := []widget.SelectOption{}
		dcId := c.datacenter.Value()
		for _, zne := range dta.zones {
			if zne.Datacenter.Hex() != dcId {
				continue
			}
			opts = append(opts, widget.SelectOption{
				Label: zne.Name,
				Value: zne.Id.Hex(),
			})
		}
		return withSelect("Select Zone", opts)
	}

	shapeOpts := []widget.SelectOption{}
	for _, shpe := range dta.shapes {
		shapeOpts = append(shapeOpts, widget.SelectOption{
			Label: shpe.Name,
			Value: shpe.Id.Hex(),
		})
	}
	c.shape = form.NewSelect("Shape", withSelect("No Shape", shapeOpts), "")
	c.shape.OnChange = c.applyShape

	c.node = form.NewSelect("Node", nil, "")
	c.node.Dynamic = func() []widget.SelectOption {
		opts := []widget.SelectOption{}
		zoneId := c.zone.Value()
		for _, nde := range dta.nodes {
			if nde.Zone.Hex() != zoneId {
				continue
			}
			opts = append(opts, widget.SelectOption{
				Label: nde.Name,
				Value: nde.Id.Hex(),
			})
		}
		return withSelect("Select Node", opts)
	}
	c.node.Hide = func() bool {
		return c.shape.Value() != ""
	}

	c.vpc = form.NewSelect("VPC", nil, "")
	c.vpc.Dynamic = func() []widget.SelectOption {
		opts := []widget.SelectOption{}
		orgId := c.organization.Value()
		dcId := c.datacenter.Value()
		for _, vc := range dta.vpcs {
			if vc.Organization.Hex() != orgId ||
				vc.Datacenter.Hex() != dcId {

				continue
			}
			opts = append(opts, widget.SelectOption{
				Label: vc.Name,
				Value: vc.Id.Hex(),
			})
		}
		return withSelect("Select VPC", opts)
	}

	c.subnet = form.NewSelect("Subnet", nil, "")
	c.subnet.Dynamic = func() []widget.SelectOption {
		opts := []widget.SelectOption{}
		vpcId := c.vpc.Value()
		for _, vc := range dta.vpcs {
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
		return withSelect("Select Subnet", opts)
	}

	c.cloudType = form.NewSelect("CloudInit Type", cloudTypeOptions,
		instance.Linux)
	c.cloudInterface = form.NewSelect("Guest Interface Mode",
		cloudInterfaceOptions, instance.Default)

	c.startup = form.NewToggle("Startup script", false)
	c.script = form.NewArea("Startup Script", "Startup script", "", 6)
	c.script.Hide = func() bool {
		return !c.startup.Value()
	}

	c.start = form.NewToggle("Start instance", true)
	c.uefi = form.NewToggle("UEFI", true)
	c.secureBoot = form.NewToggle("SecureBoot", true)
	c.secureBoot.Hide = func() bool {
		return !c.uefi.Value()
	}
	c.secureBoot.OnChange = func() {
		c.secureBootChanged = true
	}
	c.tpm = form.NewToggle("TPM", false)
	c.tpm.Hide = c.secureBoot.Hide

	c.publicIp4 = form.NewToggle("Public IPv4 address", true)
	c.publicIp6 = form.NewToggle("Public IPv6 address", true)
	c.hostAddr = form.NewToggle("Host address", true)

	c.diskType = form.NewSelect("Disk Type", diskTypeOptions, disk.Qcow2)
	c.diskType.Lock = c.shapeSelected
	c.diskPool = form.NewSelect("Pool", nil, "")
	c.diskPool.Dynamic = func() []widget.SelectOption {
		opts := []widget.SelectOption{}
		nde := c.selectedNode()
		for _, pl := range dta.pools {
			if nde == nil || !containsId(nde.Pools, pl.Id) {
				continue
			}
			opts = append(opts, widget.SelectOption{
				Label: pl.Name,
				Value: pl.Id.Hex(),
			})
		}
		return withSelect("Select Pool", opts)
	}
	// The server picks the pool of a shape, the pool is only chosen with
	// a node and LVM disks
	c.diskPool.Hide = func() bool {
		return c.diskType.Value() != disk.Lvm || c.shapeSelected()
	}

	c.hiddenImages = form.NewToggle("Show hidden images", false)
	c.imageBacking = form.NewToggle("Linked disk image", false)
	c.image = form.NewSelect("Image", nil, "")
	c.image.Dynamic = c.imageOptions
	c.image.OnChange = c.applyImage

	c.roles = form.NewTokens("Roles", "Add role", []string{})
	c.diskSize = form.NewNumber("Disk Size (GB)", "Disk size in gigabytes",
		10, 10, 0, 10)
	c.memory = form.NewNumber("Memory Size (MB)", "Memory in megabytes",
		1024, 256, 0, 1024)
	c.memory.Lock = c.shapeFixed
	c.processors = form.NewNumber("Processors", "Number of processors",
		1, 1, 0, 1)
	c.processors.Lock = c.shapeFixed
	c.count = form.NewNumber("Count", "Number of instances", 1, 1, 0, 1)

	c.nodePorts = form.NewRows("Node Ports", "No node ports",
		"Add Node Port",
		[]string{"Protocol", "External Port", "Internal Port"},
		newNodePortRow)
	c.mounts = form.NewRows("Host Paths", "No host paths",
		"Add Host Path", []string{"Host Path", "Name", "Instance Path"},
		newMountRow)

	c.form = form.New(
		c.name,
		c.comment,
		c.organization,
		c.datacenter,
		c.zone,
		c.shape,
		c.node,
		c.vpc,
		c.subnet,
		c.cloudType,
		c.cloudInterface,
		c.startup,
		c.script,
		c.start,
		c.uefi,
		c.secureBoot,
		c.tpm,
		c.publicIp4,
		c.publicIp6,
		c.hostAddr,
		c.diskType,
		c.diskPool,
		c.hiddenImages,
		c.imageBacking,
		c.image,
		c.roles,
		c.diskSize,
		c.memory,
		c.processors,
		c.count,
		c.nodePorts,
		c.mounts,
	)

	return c
}

func containsId(ids []bson.ObjectID, id bson.ObjectID) bool {
	for _, cur := range ids {
		if cur == id {
			return true
		}
	}
	return false
}

func (c *creator) selectedShape() *shape.Shape {
	shapeId := c.shape.Value()
	for _, shpe := range c.data.shapes {
		if shpe.Id.Hex() == shapeId {
			return shpe
		}
	}
	return nil
}

func (c *creator) shapeSelected() bool {
	return c.selectedShape() != nil
}

// shapeFixed returns true when a shape without flexible resources is
// selected, its memory and processors cannot be changed.
func (c *creator) shapeFixed() bool {
	shpe := c.selectedShape()
	return shpe != nil && !shpe.Flexible
}

// applyShape copies the disk type, pool, processors and memory of the
// selected shape like the web interface.
func (c *creator) applyShape() {
	shpe := c.selectedShape()
	if shpe == nil {
		return
	}

	c.diskType.SetValue(widget.Default(shpe.DiskType, disk.Qcow2))
	c.diskPool.SetValue(shpe.DiskPool.Hex())
	if shpe.DiskPool.IsZero() {
		c.diskPool.SetValue("")
	}
	if shpe.Processors > 0 {
		c.processors.Set(shpe.Processors)
	}
	if shpe.Memory > 0 {
		c.memory.Set(shpe.Memory)
	}
}

func (c *creator) selectedNode() *node.Node {
	nodeId := c.node.Value()
	for _, nde := range c.data.nodes {
		if nde.Id.Hex() == nodeId {
			return nde
		}
	}
	return nil
}

func (c *creator) selectedImage() *image.Image {
	dc := c.selectedDatacenter()
	if dc == nil {
		return nil
	}
	imgId := c.image.Value()
	for _, img := range c.data.images[dc.Id] {
		if img.Id.Hex() == imgId {
			return img
		}
	}
	return nil
}

// applyImage sets the secure boot and cloud init type of the selected
// image, BSD, Alpine and Arch images do not support secure boot and BSD
// images use the BSD cloud init. Secure boot is restored for other
// images unless the user changed it.
func (c *creator) applyImage() {
	img := c.selectedImage()
	if img == nil {
		return
	}

	name := strings.ToLower(img.Name)
	switch {
	case strings.Contains(name, "bsd"):
		c.secureBoot.SetValue(false)
		c.cloudType.SetValue(instance.BSD)
	case strings.Contains(name, "alpinelinux"),
		strings.Contains(name, "archlinux"):
		c.secureBoot.SetValue(false)
		c.cloudType.SetValue(instance.Linux)
	default:
		if !c.secureBootChanged {
			c.secureBoot.SetValue(true)
		}
		c.cloudType.SetValue(instance.Linux)
	}
}

func (c *creator) selectedDatacenter() *datacenter.Datacenter {
	dcId := c.datacenter.Value()
	for _, dc := range c.data.datacenters {
		if dc.Id.Hex() == dcId {
			return dc
		}
	}
	return nil
}

// imageVersion returns the release key and build version of a signed
// image key such as ubuntu2404_2409.qcow2, ok is false when the key has
// no version like the image menu of the web interface.
func imageVersion(key string) (release string, version int, ok bool) {
	parts := strings.Split(key, "_")
	if len(parts) < 2 {
		return
	}

	last := parts[len(parts)-1]
	if len(last) < 4 {
		return
	}

	version, err := strconv.Atoi(last[:4])
	if err != nil || version == 0 {
		return
	}

	return parts[0], version, true
}

// imageOptions lists the images of the selected datacenter matching the
// UEFI setting, only the latest build of each signed release is shown
// unless hidden images are enabled, like the image menu of the web
// interface.
func (c *creator) imageOptions() []widget.SelectOption {
	dc := c.selectedDatacenter()
	if dc == nil {
		return withSelect("Select Datacenter", nil)
	}

	uefi := c.uefi.Value()
	latest := map[string]*image.Image{}
	latestVer := map[string]int{}
	signed := []*image.Image{}
	others := []*image.Image{}

	for _, img := range c.data.images[dc.Id] {
		if uefi && img.Firmware == "bios" {
			continue
		}
		if !uefi && img.Firmware == "uefi" {
			continue
		}

		if !img.Signed {
			others = append(others, img)
			continue
		}

		if c.hiddenImages.Value() {
			signed = append(signed, img)
			continue
		}

		release, version, ok := imageVersion(img.Key)
		if !ok {
			signed = append(signed, img)
			continue
		}

		if cur, exists := latestVer[release]; !exists || version > cur {
			latest[release] = img
			latestVer[release] = version
		}
	}
	for _, img := range latest {
		signed = append(signed, img)
	}

	sort.Slice(signed, func(i, j int) bool {
		return imageLabel(signed[i]) < imageLabel(signed[j])
	})
	sort.Slice(others, func(i, j int) bool {
		return others[i].Name < others[j].Name
	})

	opts := []widget.SelectOption{}
	for _, img := range append(signed, others...) {
		opts = append(opts, widget.SelectOption{
			Label: imageLabel(img),
			Value: img.Id.Hex(),
		})
	}

	return withSelect("No Image", opts)
}

func (c *creator) Form() *form.Form {
	return c.form
}

func createError(message string) error {
	return &errortypes.ParseError{
		errors.New("instances: " + message),
	}
}

func (c *creator) buildNodePorts() []*nodeport.Mapping {
	mappings := []*nodeport.Mapping{}
	for _, row := range c.nodePorts.Rows() {
		mappings = append(mappings, &nodeport.Mapping{
			Protocol:     row.Cells[0].(*form.Select).Value(),
			ExternalPort: row.Cells[1].(*form.Number).Int(),
			InternalPort: row.Cells[2].(*form.Number).Int(),
		})
	}
	return mappings
}

func (c *creator) buildMounts() []*instance.Mount {
	mounts := []*instance.Mount{}
	for _, row := range c.mounts.Rows() {
		mounts = append(mounts, &instance.Mount{
			Type:     instance.HostPath,
			Path:     row.Cells[0].(*form.Text).Value(),
			Name:     row.Cells[1].(*form.Text).Value(),
			HostPath: row.Cells[2].(*form.Text).Value(),
		})
	}
	return mounts
}

// Save inserts the instances the same way as the admin instance create
// handler, the count creates several instances with the name formatted
// by index when it contains a percent verb.
func (c *creator) Save(db *database.Database) (err error) {
	zoneId, ok := utils.ParseObjectId(c.zone.Value())
	if !ok {
		err = createError("Missing required zone")
		return
	}

	zne, err := zone.Get(db, zoneId)
	if err != nil {
		return
	}

	orgId, _ := utils.ParseObjectId(c.organization.Value())
	shapeId, _ := utils.ParseObjectId(c.shape.Value())
	nodeId, _ := utils.ParseObjectId(c.node.Value())
	diskType := c.diskType.Value()
	diskPool, _ := utils.ParseObjectId(c.diskPool.Value())

	if !shapeId.IsZero() {
		nodeId = bson.NilObjectID
		diskType = ""
		diskPool = bson.NilObjectID
	} else {
		if nodeId.IsZero() {
			err = createError("Missing required node")
			return
		}

		nde, e := node.Get(db, nodeId)
		if e != nil {
			err = e
			return
		}

		if nde.Zone != zne.Id {
			err = createError("Node is not in the zone")
			return
		}

		if diskType == disk.Lvm && !containsId(nde.Pools, diskPool) {
			err = createError("Pool not found")
			return
		}
	}

	imageId, _ := utils.ParseObjectId(c.image.Value())
	if !imageId.IsZero() {
		img, e := image.GetOrgPublic(db, orgId, imageId)
		if e != nil {
			if _, ok := e.(*database.NotFoundError); ok {
				err = createError("Image not found")
			} else {
				err = e
			}
			return
		}

		stre, e := storage.Get(db, img.Storage)
		if e != nil {
			err = e
			return
		}

		available, e := data.ImageAvailable(stre, img)
		if e != nil {
			err = e
			return
		}
		if !available {
			err = createError("Image not restored from archive")
			return
		}
	}

	action := instance.Start
	if !c.start.Value() {
		action = instance.Stop
	}

	script := ""
	if c.startup.Value() {
		script = c.script.Value()
	}

	count := max(c.count.Int(), 1)
	baseName := c.name.Value()

	for i := 0; i < count; i++ {
		name := baseName
		if strings.Contains(baseName, "%") {
			name = fmt.Sprintf(baseName, i+1)
		}

		inst := &instance.Instance{
			Action:           action,
			Organization:     orgId,
			Zone:             zoneId,
			Vpc:              parseId(c.vpc.Value()),
			Subnet:           parseId(c.subnet.Value()),
			Shape:            shapeId,
			Node:             nodeId,
			DiskType:         diskType,
			DiskPool:         diskPool,
			Image:            imageId,
			ImageBacking:     c.imageBacking.Value(),
			Uefi:             c.uefi.Value(),
			SecureBoot:       c.secureBoot.Value(),
			Tpm:              c.tpm.Value(),
			CloudType:        c.cloudType.Value(),
			CloudScript:      script,
			CloudInterface:   c.cloudInterface.Value(),
			Name:             name,
			Comment:          c.comment.Value(),
			InitDiskSize:     c.diskSize.Int(),
			Memory:           c.memory.Int(),
			Processors:       c.processors.Int(),
			Roles:            c.roles.Values(),
			Mounts:           c.buildMounts(),
			NodePorts:        c.buildNodePorts(),
			NoPublicAddress:  !c.publicIp4.Value(),
			NoPublicAddress6: !c.publicIp6.Value(),
			NoHostAddress:    !c.hostAddr.Value(),
		}

		errData, e := inst.Validate(db)
		if e != nil {
			if _, ok := e.(*errortypes.NotFoundError); ok {
				err = createError("Failed to find available node")
			} else {
				err = e
			}
			return
		}
		if errData != nil {
			err = createError(errData.Message)
			return
		}

		err = inst.SyncNodePorts(db)
		if err != nil {
			return
		}

		err = inst.Insert(db)
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
		}).Info("tui: Instance created")
	}

	err = event.PublishDispatch(db, "instance.change")
	if err != nil {
		return
	}

	return
}

func parseId(hex string) bson.ObjectID {
	id, _ := utils.ParseObjectId(hex)
	return id
}

// NewEditor returns the new instance form with the web defaults.
func (p *Provider) NewEditor(db *database.Database) (
	editor resource.Editor, err error) {

	dta, err := loadCreateData(db)
	if err != nil {
		return
	}

	editor = newCreator(dta)
	return
}
