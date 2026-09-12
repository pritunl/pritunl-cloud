package disks

import (
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
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/pritunl/pritunl-cloud/zone"
	"github.com/sirupsen/logrus"
)

var fileSystemOptions = []widget.SelectOption{
	{Label: "None", Value: ""},
	{Label: "XFS", Value: disk.Xfs},
	{Label: "ext4", Value: disk.Ext4},
	{Label: "LVM + XFS", Value: disk.LvmXfs},
	{Label: "LVM + ext4", Value: disk.LvmExt4},
}

// createData are the select options of the new disk form loaded once
// when the form opens.
type createData struct {
	orgs        []*database.Named
	datacenters []*datacenter.Datacenter
	zones       []*zone.Zone
	nodes       []*node.Node
	pools       []*pool.Pool

	// images are keyed by datacenter, the image name query only returns
	// name, key, signed and firmware so the datacenter storages are part
	// of the query like the admin images handler
	images map[bson.ObjectID][]*image.Image

	// instances are keyed by node like the instance query of the web
	// interface
	instances map[bson.ObjectID][]*nodeInstance
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

	dta.nodes, err = node.GetAllHypervisors(db, &bson.M{})
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

	nodeIds := []bson.ObjectID{}
	for _, nde := range dta.nodes {
		nodeIds = append(nodeIds, nde.Id)
	}
	dta.instances, err = loadNodeInstances(db, nodeIds)
	if err != nil {
		return
	}

	return
}

// creator is the new disk form mirroring the new disk panel of the web
// interface, Save inserts the disk.
type creator struct {
	data *createData
	form *form.Form

	name             *form.Text
	typ              *form.Select
	organization     *form.Select
	datacenter       *form.Select
	zone             *form.Select
	node             *form.Select
	pool             *form.Select
	deleteProtection *form.Toggle
	instance         *form.Select
	index            *form.Number
	image            *form.Select
	fileSystem       *form.Select
	hiddenImages     *form.Toggle
	backing          *form.Toggle
	lvSize           *form.Number
	size             *form.Number
}

func newCreator(dta *createData) *creator {
	c := &creator{
		data: dta,
	}

	c.name = form.NewText("Name", "Enter name", "new-disk")
	c.typ = form.NewSelect("Type", diskTypeOptions, disk.Qcow2)

	orgOpts := resource.SelectOptions(dta.orgs)
	if len(orgOpts) == 0 {
		orgOpts = resource.WithSelect("No Organizations", nil)
	}
	c.organization = form.NewSelect("Organization", orgOpts, "")

	dcOpts := []widget.SelectOption{}
	for _, dc := range dta.datacenters {
		dcOpts = append(dcOpts, widget.SelectOption{
			Label: dc.Name,
			Value: dc.Id.Hex(),
		})
	}
	if len(dcOpts) > 0 {
		dcOpts = resource.WithSelect("Select Datacenter", dcOpts)
	} else {
		dcOpts = resource.WithSelect("No Datacenters", nil)
	}
	c.datacenter = form.NewSelect("Datacenter", dcOpts, "")
	c.datacenter.OnChange = func() {
		// Changing the datacenter clears the node and image like the
		// web interface, the zone and instance depend on them
		c.zone.SetValue("")
		c.node.SetValue("")
		c.instance.SetValue("")
		c.image.SetValue("")
	}

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
		if len(opts) == 0 {
			return resource.WithSelect("No Zones", nil)
		}
		return resource.WithSelect("Select Zone", opts)
	}
	c.zone.OnChange = func() {
		c.node.SetValue("")
		c.instance.SetValue("")
	}

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
		if len(opts) == 0 {
			return resource.WithSelect("No Nodes", nil)
		}
		return resource.WithSelect("Select Node", opts)
	}
	c.node.Hide = func() bool {
		return c.typ.Value() == disk.Lvm
	}
	c.node.OnChange = func() {
		c.instance.SetValue("")
	}

	c.pool = form.NewSelect("Pool", nil, "")
	c.pool.Dynamic = func() []widget.SelectOption {
		opts := []widget.SelectOption{}
		zoneId := c.zone.Value()
		for _, pl := range dta.pools {
			if pl.Zone.Hex() != zoneId {
				continue
			}
			opts = append(opts, widget.SelectOption{
				Label: pl.Name,
				Value: pl.Id.Hex(),
			})
		}
		if len(opts) == 0 {
			return resource.WithSelect("No Pools", nil)
		}
		return resource.WithSelect("Select Pool", opts)
	}
	c.pool.Hide = func() bool {
		return c.typ.Value() != disk.Lvm
	}

	c.deleteProtection = form.NewToggle("Delete protection", false)

	c.instance = form.NewSelect("Instance", nil, "")
	c.instance.Dynamic = func() []widget.SelectOption {
		nodeId := resource.ParseId(c.node.Value())
		opts := instanceOptions(dta.instances[nodeId])
		if len(opts) == 0 {
			return resource.WithSelect("No Instances", nil)
		}
		return resource.WithSelect("Detached Disk", opts)
	}

	c.index = form.NewNumber("Index", "Disk index", 1, 0, 8, 1)
	c.index.Hide = func() bool {
		return c.instance.Value() == ""
	}

	// The image and file system are exclusive, picking one clears the
	// other like the web interface
	c.image = form.NewSelect("Image", nil, "")
	c.image.Dynamic = c.imageOptions
	c.image.OnChange = func() {
		if c.image.Value() != "" {
			c.fileSystem.SetValue("")
		}
	}

	c.fileSystem = form.NewSelect("File System", fileSystemOptions, "")
	c.fileSystem.OnChange = func() {
		if c.fileSystem.Value() != "" {
			c.image.SetValue("")
		}
	}

	c.hiddenImages = form.NewToggle("Show hidden images", false)
	c.backing = form.NewToggle("Linked disk image", false)

	c.lvSize = form.NewNumber("LV Size (GB)",
		"Logical volume size in gigabytes", 0, 10, 0, 1)
	c.lvSize.Hide = func() bool {
		return !strings.Contains(c.fileSystem.Value(), "lvm")
	}

	c.size = form.NewNumber("Size (GB)", "Disk size in gigabytes", 10, 10,
		0, 1)

	c.form = form.New(
		c.name,
		c.typ,
		c.organization,
		c.datacenter,
		c.zone,
		c.node,
		c.pool,
		c.deleteProtection,
		c.instance,
		c.index,
		c.image,
		c.fileSystem,
		c.hiddenImages,
		c.backing,
		c.lvSize,
		c.size,
	)

	return c
}

// imageLabel returns the image name, signed image names already carry
// the build date from the key.
func imageLabel(img *image.Image) string {
	return widget.Default(img.Name, img.Key)
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

// imageOptions lists the images of the selected datacenter, only the
// latest build of each signed release and firmware is shown unless
// hidden images are enabled, like the image menu of the new disk panel.
func (c *creator) imageOptions() []widget.SelectOption {
	dcId := resource.ParseId(c.datacenter.Value())
	if dcId.IsZero() {
		return resource.WithSelect("Blank Disk", nil)
	}

	latest := map[string]*image.Image{}
	latestVer := map[string]int{}
	signed := []*image.Image{}
	others := []*image.Image{}

	for _, img := range c.data.images[dcId] {
		if !img.Signed || c.hiddenImages.Value() {
			others = append(others, img)
			continue
		}

		release, version, ok := imageVersion(img.Key)
		if !ok {
			others = append(others, img)
			continue
		}

		key := release + "_" + img.Firmware
		if cur, exists := latestVer[key]; !exists || version > cur {
			latest[key] = img
			latestVer[key] = version
		}
	}
	for _, img := range latest {
		signed = append(signed, img)
	}

	sort.Slice(signed, func(i, j int) bool {
		return imageLabel(signed[i]) < imageLabel(signed[j])
	})
	sort.Slice(others, func(i, j int) bool {
		return imageLabel(others[i]) < imageLabel(others[j])
	})

	opts := []widget.SelectOption{}
	for _, img := range append(others, signed...) {
		opts = append(opts, widget.SelectOption{
			Label: imageLabel(img),
			Value: img.Id.Hex(),
		})
	}

	return resource.WithSelect("Blank Disk", opts)
}

func (c *creator) Form() *form.Form {
	return c.form
}

func createError(message string) error {
	return &errortypes.ParseError{
		errors.New("disks: " + message),
	}
}

// Save inserts the disk the same way as the admin disk create handler.
func (c *creator) Save(db *database.Database) (err error) {
	nodeId := resource.ParseId(c.node.Value())
	if nodeId.IsZero() {
		err = createError("Missing required node")
		return
	}

	nde, err := node.Get(db, nodeId)
	if err != nil {
		return
	}

	orgId := resource.ParseId(c.organization.Value())
	imageId := resource.ParseId(c.image.Value())

	imgSystemType := ""
	imgSystemKind := ""
	if !imageId.IsZero() {
		img, e := image.GetOrgPublic(db, orgId, imageId)
		if e != nil {
			err = e
			return
		}

		imgSystemType = img.GetSystemType()
		imgSystemKind = img.GetSystemKind()

		store, e := storage.Get(db, img.Storage)
		if e != nil {
			err = e
			return
		}

		available, e := data.ImageAvailable(store, img)
		if e != nil {
			err = e
			return
		}

		if !available {
			if store.IsOracle() {
				err = createError("Image not restored from archive")
			} else {
				err = createError("Image not restored from glacier")
			}
			return
		}
	}

	instId := resource.ParseId(c.instance.Value())
	index := ""
	if !instId.IsZero() {
		index = strconv.Itoa(c.index.Int())
	}

	fileSystem := c.fileSystem.Value()
	lvSize := 0
	if strings.Contains(fileSystem, "lvm") {
		lvSize = c.lvSize.Int()
	}

	dsk := &disk.Disk{
		Name:             c.name.Value(),
		Organization:     orgId,
		Instance:         instId,
		Datacenter:       nde.Datacenter,
		Zone:             nde.Zone,
		Index:            index,
		Type:             c.typ.Value(),
		SystemType:       imgSystemType,
		SystemKind:       imgSystemKind,
		Node:             nodeId,
		Pool:             resource.ParseId(c.pool.Value()),
		Image:            imageId,
		DeleteProtection: c.deleteProtection.Value(),
		FileSystem:       fileSystem,
		Backing:          c.backing.Value(),
		Size:             c.size.Int(),
		LvSize:           lvSize,
	}

	errData, err := dsk.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = createError(errData.Message)
		return
	}

	err = dsk.Insert(db)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"disk_id": dsk.Id.Hex(),
	}).Info("tui: Disk created")

	err = event.PublishDispatch(db, "disk.change")
	if err != nil {
		return
	}

	return
}

// NewEditor returns the new disk form with the web defaults.
func (p *Provider) NewEditor(db *database.Database) (
	editor resource.Editor, err error) {

	dta, err := loadCreateData(db)
	if err != nil {
		return
	}

	editor = newCreator(dta)
	return
}
