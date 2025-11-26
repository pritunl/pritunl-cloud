package uhandlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/drive"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/iscsi"
	"github.com/pritunl/pritunl-cloud/iso"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/nodeport"
	"github.com/pritunl/pritunl-cloud/pci"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/pritunl/pritunl-cloud/usb"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
)

type instanceData struct {
	Id                  bson.ObjectID       `json:"id"`
	Zone                bson.ObjectID       `json:"zone"`
	Vpc                 bson.ObjectID       `json:"vpc"`
	Subnet              bson.ObjectID       `json:"subnet"`
	CloudSubnet         string              `json:"cloud_subnet"`
	Shape               bson.ObjectID       `json:"shape"`
	Node                bson.ObjectID       `json:"node"`
	DiskType            string              `json:"disk_type"`
	DiskPool            bson.ObjectID       `json:"disk_pool"`
	Image               bson.ObjectID       `json:"image"`
	ImageBacking        bool                `json:"image_backing"`
	Name                string              `json:"name"`
	Comment             string              `json:"comment"`
	Action              string              `json:"action"`
	RootEnabled         bool                `json:"root_enabled"`
	Uefi                bool                `json:"uefi"`
	SecureBoot          bool                `json:"secure_boot"`
	Tpm                 bool                `json:"tpm"`
	DhcpServer          bool                `json:"dhcp_server"`
	CloudType           string              `json:"cloud_type"`
	CloudScript         string              `json:"cloud_script"`
	DeleteProtection    bool                `json:"delete_protection"`
	SkipSourceDestCheck bool                `json:"skip_source_dest_check"`
	InitDiskSize        int                 `json:"init_disk_size"`
	Memory              int                 `json:"memory"`
	Processors          int                 `json:"processors"`
	Roles               []string            `json:"roles"`
	Isos                []*iso.Iso          `json:"isos"`
	UsbDevices          []*usb.Device       `json:"usb_devices"`
	PciDevices          []*pci.Device       `json:"pci_devices"`
	DriveDevices        []*drive.Device     `json:"drive_devices"`
	IscsiDevices        []*iscsi.Device     `json:"iscsi_devices"`
	Mounts              []*instance.Mount   `json:"mounts"`
	Vnc                 bool                `json:"vnc"`
	Spice               bool                `json:"spice"`
	Gui                 bool                `json:"gui"`
	NodePorts           []*nodeport.Mapping `json:"node_ports"`
	NoPublicAddress     bool                `json:"no_public_address"`
	NoPublicAddress6    bool                `json:"no_public_address6"`
	NoHostAddress       bool                `json:"no_host_address"`
	Count               int                 `json:"count"`
}

type instanceMultiData struct {
	Ids    []bson.ObjectID `json:"ids"`
	Action string          `json:"action"`
}

type instancesData struct {
	Instances []*instance.Instance `json:"instances"`
	Count     int64                `json:"count"`
}

func instancePut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)
	dta := &instanceData{}

	instanceId, ok := utils.ParseObjectId(c.Param("instance_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(dta)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	inst, err := instance.GetOrg(db, userOrg, instanceId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	exists, err := vpc.ExistsOrg(db, userOrg, dta.Vpc)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}
	if !exists {
		utils.AbortWithStatus(c, 405)
		return
	}

	inst.PreCommit()

	inst.Name = dta.Name
	inst.Comment = dta.Comment
	inst.Vpc = dta.Vpc
	inst.Subnet = dta.Subnet
	inst.CloudSubnet = dta.CloudSubnet
	if dta.Action != "" {
		inst.Action = dta.Action
	}
	inst.Uefi = dta.Uefi
	inst.SecureBoot = dta.SecureBoot
	inst.Tpm = dta.Tpm
	inst.DhcpServer = dta.DhcpServer
	inst.CloudType = dta.CloudType
	inst.CloudScript = dta.CloudScript
	inst.DeleteProtection = dta.DeleteProtection
	inst.SkipSourceDestCheck = dta.SkipSourceDestCheck
	inst.Memory = dta.Memory
	inst.Processors = dta.Processors
	inst.Roles = dta.Roles
	inst.Isos = dta.Isos
	inst.UsbDevices = dta.UsbDevices
	inst.PciDevices = dta.PciDevices
	inst.DriveDevices = dta.DriveDevices
	inst.IscsiDevices = dta.IscsiDevices
	inst.Mounts = dta.Mounts
	inst.RootEnabled = dta.RootEnabled
	inst.Vnc = dta.Vnc
	inst.Spice = dta.Spice
	inst.Gui = dta.Gui
	inst.NodePorts = dta.NodePorts
	inst.NoPublicAddress = dta.NoPublicAddress
	inst.NoPublicAddress6 = dta.NoPublicAddress6
	inst.NoHostAddress = dta.NoHostAddress

	fields := set.NewSet(
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

	errData, err := inst.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	dskChange, err := inst.PostCommit(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	err = inst.CommitFields(db, fields)
	if err != nil {
		_ = inst.Cleanup(db)

		utils.AbortWithError(c, 500, err)
		return
	}

	err = inst.Cleanup(db)
	if err != nil {
		return
	}

	event.PublishDispatch(db, "instance.change")
	if dskChange {
		event.PublishDispatch(db, "disk.change")
	}

	c.JSON(200, inst)
}

func instancePost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)
	dta := &instanceData{
		Name: "new-instance",
	}

	err := c.Bind(dta)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	zne, err := zone.Get(db, dta.Zone)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	exists, err := datacenter.ExistsOrg(db, userOrg, zne.Datacenter)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}
	if !exists {
		utils.AbortWithStatus(c, 405)
		return
	}

	if !dta.Shape.IsZero() {
		dta.Node = bson.NilObjectID
		dta.DiskType = ""
		dta.DiskPool = bson.NilObjectID
	} else {
		nde, err := node.Get(db, dta.Node)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		if nde.Zone != zne.Id {
			utils.AbortWithStatus(c, 405)
			return
		}

		if dta.DiskType == disk.Lvm {
			poolMatch := false
			for _, plId := range nde.Pools {
				if plId == dta.DiskPool {
					poolMatch = true
				}
			}

			if !poolMatch {
				errData := &errortypes.ErrorData{
					Error:   "pool_not_found",
					Message: "Pool not found",
				}
				c.JSON(400, errData)
				return
			}
		}
	}

	exists, err = vpc.ExistsOrg(db, userOrg, dta.Vpc)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}
	if !exists {
		utils.AbortWithStatus(c, 405)
		return
	}

	if !dta.Image.IsZero() {
		img, err := image.GetOrgPublic(db, userOrg, dta.Image)
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				errData := &errortypes.ErrorData{
					Error:   "image_not_found",
					Message: "Image not found",
				}
				c.JSON(400, errData)
			} else {
				utils.AbortWithError(c, 500, err)
			}
			return
		}

		stre, err := storage.Get(db, img.Storage)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		available, err := data.ImageAvailable(stre, img)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
		if !available {
			if stre.IsOracle() {
				errData := &errortypes.ErrorData{
					Error:   "image_not_available",
					Message: "Image not restored from archive",
				}
				c.JSON(400, errData)
			} else {
				errData := &errortypes.ErrorData{
					Error:   "image_not_available",
					Message: "Image not restored from glacier",
				}
				c.JSON(400, errData)
			}

			return
		}
	}

	insts := []*instance.Instance{}

	if dta.Count == 0 {
		dta.Count = 1
	}

	for i := 0; i < dta.Count; i++ {
		name := ""
		if strings.Contains(dta.Name, "%") {
			name = fmt.Sprintf(dta.Name, i+1)
		} else {
			name = dta.Name
		}

		inst := &instance.Instance{
			Action:              dta.Action,
			Organization:        userOrg,
			Zone:                dta.Zone,
			Vpc:                 dta.Vpc,
			Subnet:              dta.Subnet,
			CloudSubnet:         dta.CloudSubnet,
			Shape:               dta.Shape,
			Node:                dta.Node,
			DiskType:            dta.DiskType,
			DiskPool:            dta.DiskPool,
			Image:               dta.Image,
			ImageBacking:        dta.ImageBacking,
			Uefi:                dta.Uefi,
			SecureBoot:          dta.SecureBoot,
			Tpm:                 dta.Tpm,
			DhcpServer:          dta.DhcpServer,
			CloudType:           dta.CloudType,
			CloudScript:         dta.CloudScript,
			DeleteProtection:    dta.DeleteProtection,
			SkipSourceDestCheck: dta.SkipSourceDestCheck,
			Name:                name,
			Comment:             dta.Comment,
			InitDiskSize:        dta.InitDiskSize,
			Memory:              dta.Memory,
			Processors:          dta.Processors,
			Roles:               dta.Roles,
			Isos:                dta.Isos,
			UsbDevices:          dta.UsbDevices,
			PciDevices:          dta.PciDevices,
			DriveDevices:        dta.DriveDevices,
			IscsiDevices:        dta.IscsiDevices,
			Mounts:              dta.Mounts,
			RootEnabled:         dta.RootEnabled,
			Vnc:                 dta.Vnc,
			Spice:               dta.Spice,
			Gui:                 dta.Gui,
			NodePorts:           dta.NodePorts,
			NoPublicAddress:     dta.NoPublicAddress,
			NoHostAddress:       dta.NoHostAddress,
		}

		errData, err := inst.Validate(db)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		if errData != nil {
			c.JSON(400, errData)
			return
		}

		err = inst.SyncNodePorts(db)
		if err != nil {
			return
		}

		err = inst.Insert(db)
		if err != nil {
			_ = inst.Cleanup(db)

			utils.AbortWithError(c, 500, err)
			return
		}

		err = inst.Cleanup(db)
		if err != nil {
			return
		}

		insts = append(insts, inst)
	}

	event.PublishDispatch(db, "instance.change")

	if len(insts) == 1 {
		c.JSON(200, insts[0])
	} else {
		c.JSON(200, insts)
	}
}

func instancesPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)
	dta := &instanceMultiData{}

	err := c.Bind(dta)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	if !instance.ValidActions.Contains(dta.Action) {
		errData := &errortypes.ErrorData{
			Error:   "invalid_action",
			Message: "Invalid instance action",
		}
		c.JSON(400, errData)
		return
	}

	doc := bson.M{
		"action": dta.Action,
	}

	if dta.Action != instance.Start {
		doc["restart"] = false
		doc["restart_block_ip"] = false
	}

	err = instance.UpdateMultiOrg(db, userOrg, dta.Ids, &doc)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "instance.change")

	c.JSON(200, nil)
}

func instanceDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)

	instanceId, ok := utils.ParseObjectId(c.Param("instance_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	inst, err := instance.Get(db, instanceId)
	if err != nil {
		return
	}

	if inst.DeleteProtection {
		errData := &errortypes.ErrorData{
			Error:   "delete_protection",
			Message: "Cannot delete instance with delete protection",
		}

		c.JSON(400, errData)
		return
	}

	err = instance.DeleteOrg(db, userOrg, instanceId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "instance.change")

	c.JSON(200, nil)
}

func instancesDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)
	dta := []bson.ObjectID{}

	err := c.Bind(&dta)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	err = instance.DeleteMultiOrg(db, userOrg, dta)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "instance.change")

	c.JSON(200, nil)
}

func instanceGet(c *gin.Context) {
	if demo.IsDemo() {
		inst := demo.Instances[0]
		inst.Guest.Timestamp = time.Now()
		inst.Guest.Heartbeat = time.Now()
		c.JSON(200, inst)
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)

	instanceId, ok := utils.ParseObjectId(c.Param("instance_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	inst, err := instance.GetOrg(db, userOrg, instanceId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if demo.IsDemo() {
		inst.State = vm.Running
		inst.Action = instance.Start
		inst.State = vm.Running
		inst.Status = "Running"
		inst.PublicIps = []string{
			demo.RandIp(inst.Id),
		}
		inst.PublicIps6 = []string{
			demo.RandIp6(inst.Id),
		}
		inst.PrivateIps = []string{
			demo.RandPrivateIp(inst.Id),
		}
		inst.PrivateIps6 = []string{
			demo.RandPrivateIp6(inst.Id),
		}
		inst.NetworkNamespace = vm.GetNamespace(inst.Id, 0)
	}

	c.JSON(200, inst)
}

func instancesGet(c *gin.Context) {
	if demo.IsDemo() {
		for _, inst := range demo.Instances {
			inst.Guest.Timestamp = time.Now()
			inst.Guest.Heartbeat = time.Now()
		}

		data := &instancesData{
			Instances: demo.Instances,
			Count:     int64(len(demo.Instances)),
		}

		c.JSON(200, data)
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)

	ndeId, _ := utils.ParseObjectId(c.Query("node_names"))
	plId, _ := utils.ParseObjectId(c.Query("pool_names"))
	if !ndeId.IsZero() {
		query := &bson.M{
			"node":         ndeId,
			"organization": userOrg,
		}

		insts, err := instance.GetAllName(db, query)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		c.JSON(200, insts)
	} else if !plId.IsZero() {
		nodes, err := node.GetAllPool(db, plId)
		if err != nil {
			return
		}

		ndeIds := []bson.ObjectID{}

		for _, nde := range nodes {
			ndeIds = append(ndeIds, nde.Id)
		}

		query := &bson.M{
			"node": &bson.M{
				"$in": ndeIds,
			},
			"organization": userOrg,
		}

		insts, err := instance.GetAllName(db, query)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		c.JSON(200, insts)
	} else {
		page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
		pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

		query := bson.M{
			"organization": userOrg,
		}

		instId, ok := utils.ParseObjectId(c.Query("id"))
		if ok {
			query["_id"] = instId
		}

		name := strings.TrimSpace(c.Query("name"))
		if name != "" {
			query["name"] = &bson.M{
				"$regex":   fmt.Sprintf(".*%s.*", regexp.QuoteMeta(name)),
				"$options": "i",
			}
		}

		role := strings.TrimSpace(c.Query("role"))
		if role != "" {
			if strings.HasPrefix(role, "~") {
				role := role[1:]
				if strings.HasPrefix(role, "!") {
					query["roles"] = &bson.M{
						"$not": &bson.M{
							"$regex": fmt.Sprintf(".*%s.*",
								regexp.QuoteMeta(role[1:])),
							"$options": "i",
						},
					}
				} else {
					query["$or"] = []*bson.M{
						&bson.M{
							"roles": &bson.M{
								"$regex": fmt.Sprintf(".*%s.*",
									regexp.QuoteMeta(role)),
								"$options": "i",
							},
						},
					}
				}
			} else {
				if strings.HasPrefix(role, "!") {
					role = strings.TrimLeft(role, "!")
					query["roles"] = &bson.M{
						"$ne": role,
					}
				} else {
					query["roles"] = role
				}
			}
		}

		networkNamespace := strings.TrimSpace(c.Query("network_namespace"))
		if networkNamespace != "" {
			query["network_namespace"] = networkNamespace
		}

		nodeId, ok := utils.ParseObjectId(c.Query("node"))
		if ok {
			query["node"] = nodeId
		}

		zoneId, ok := utils.ParseObjectId(c.Query("zone"))
		if ok {
			query["zone"] = zoneId
		}

		vpcId, ok := utils.ParseObjectId(c.Query("vpc"))
		if ok {
			query["vpc"] = vpcId
		}

		subnetId, ok := utils.ParseObjectId(c.Query("subnet"))
		if ok {
			query["subnet"] = subnetId
		}

		comment := strings.TrimSpace(c.Query("comment"))
		if comment != "" {
			query["comment"] = &bson.M{
				"$regex":   fmt.Sprintf(".*%s.*", comment),
				"$options": "i",
			}
		}

		instances, count, err := instance.GetAllPaged(
			db, &query, page, pageCount)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		for _, inst := range instances {
			inst.Json(false)

			if demo.IsDemo() {
				inst.State = vm.Running
				inst.Action = instance.Start
				inst.Status = "Running"
				inst.PublicIps = []string{
					demo.RandIp(inst.Id),
				}
				inst.PublicIps6 = []string{
					demo.RandIp6(inst.Id),
				}
				inst.PrivateIps = []string{
					demo.RandPrivateIp(inst.Id),
				}
				inst.PrivateIps6 = []string{
					demo.RandPrivateIp6(inst.Id),
				}
				inst.NetworkNamespace = vm.GetNamespace(inst.Id, 0)
			}
		}

		dta := &instancesData{
			Instances: instances,
			Count:     count,
		}

		c.JSON(200, dta)
	}
}

func instanceVncGet(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)

	instanceId, ok := utils.ParseObjectId(c.Param("instance_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	inst, err := instance.GetOrg(db, userOrg, instanceId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	err = inst.VncConnect(db, c.Writer, c.Request)
	if err != nil {
		if _, ok := err.(*instance.VncDialError); ok {
			utils.AbortWithStatus(c, 504)
		} else {
			utils.AbortWithError(c, 500, err)
		}
		return
	}
}
