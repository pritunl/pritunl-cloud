package ahandlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
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
	"github.com/pritunl/pritunl-cloud/pci"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/pritunl/pritunl-cloud/usb"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/zone"
)

type instanceData struct {
	Id                  primitive.ObjectID `json:"id"`
	Organization        primitive.ObjectID `json:"organization"`
	Zone                primitive.ObjectID `json:"zone"`
	Vpc                 primitive.ObjectID `json:"vpc"`
	Subnet              primitive.ObjectID `json:"subnet"`
	OracleSubnet        string             `json:"oracle_subnet"`
	Shape               primitive.ObjectID `json:"shape"`
	Node                primitive.ObjectID `json:"node"`
	DiskType            string             `json:"disk_type"`
	DiskPool            primitive.ObjectID `json:"disk_pool"`
	Image               primitive.ObjectID `json:"image"`
	ImageBacking        bool               `json:"image_backing"`
	Name                string             `json:"name"`
	Comment             string             `json:"comment"`
	Action              string             `json:"action"`
	RootEnabled         bool               `json:"root_enabled"`
	Uefi                bool               `json:"uefi"`
	SecureBoot          bool               `json:"secure_boot"`
	Tpm                 bool               `json:"tpm"`
	DhcpServer          bool               `json:"dhcp_server"`
	CloudType           string             `json:"cloud_type"`
	CloudScript         string             `json:"cloud_script"`
	DeleteProtection    bool               `json:"delete_protection"`
	SkipSourceDestCheck bool               `json:"skip_source_dest_check"`
	InitDiskSize        int                `json:"init_disk_size"`
	Memory              int                `json:"memory"`
	Processors          int                `json:"processors"`
	NetworkRoles        []string           `json:"network_roles"`
	Isos                []*iso.Iso         `json:"isos"`
	UsbDevices          []*usb.Device      `json:"usb_devices"`
	PciDevices          []*pci.Device      `json:"pci_devices"`
	DriveDevices        []*drive.Device    `json:"drive_devices"`
	IscsiDevices        []*iscsi.Device    `json:"iscsi_devices"`
	Vnc                 bool               `json:"vnc"`
	Spice               bool               `json:"spice"`
	Gui                 bool               `json:"gui"`
	NoPublicAddress     bool               `json:"no_public_address"`
	NoPublicAddress6    bool               `json:"no_public_address6"`
	NoHostAddress       bool               `json:"no_host_address"`
	Count               int                `json:"count"`
}

type instanceMultiData struct {
	Ids    []primitive.ObjectID `json:"ids"`
	Action string               `json:"action"`
}

type instancesData struct {
	Instances []*aggregate.InstanceAggregate `json:"instances"`
	Count     int64                          `json:"count"`
}

func instancePut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	dta := &instanceData{}

	instanceId, ok := utils.ParseObjectId(c.Param("instance_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(dta)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	inst, err := instance.Get(db, instanceId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	inst.PreCommit()

	inst.Name = dta.Name
	inst.Comment = dta.Comment
	inst.Vpc = dta.Vpc
	inst.Subnet = dta.Subnet
	inst.OracleSubnet = dta.OracleSubnet
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
	inst.NetworkRoles = dta.NetworkRoles
	inst.Isos = dta.Isos
	inst.UsbDevices = dta.UsbDevices
	inst.PciDevices = dta.PciDevices
	inst.DriveDevices = dta.DriveDevices
	inst.IscsiDevices = dta.IscsiDevices
	inst.RootEnabled = dta.RootEnabled
	inst.Vnc = dta.Vnc
	inst.Spice = dta.Spice
	inst.Gui = dta.Gui
	inst.NoPublicAddress = dta.NoPublicAddress
	inst.NoPublicAddress6 = dta.NoPublicAddress6
	inst.NoHostAddress = dta.NoHostAddress

	fields := set.NewSet(
		"unix_id",
		"name",
		"comment",
		"vpc",
		"subnet",
		"oracle_subnet",
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
		"network_roles",
		"isos",
		"usb_devices",
		"pci_devices",
		"drive_devices",
		"iscsi_devices",
		"root_enabled",
		"root_passwd",
		"vnc",
		"vnc_display",
		"vnc_password",
		"spice",
		"spice_port",
		"spice_password",
		"gui",
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
		utils.AbortWithError(c, 500, err)
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
	dta := &instanceData{
		Name: "New Instance",
	}

	err := c.Bind(dta)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	zne, err := zone.Get(db, dta.Zone)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if !dta.Shape.IsZero() {
		dta.Node = primitive.NilObjectID
		dta.DiskType = ""
		dta.DiskPool = primitive.NilObjectID
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

	if !dta.Image.IsZero() {
		img, err := image.GetOrgPublic(db, dta.Organization, dta.Image)
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
			Organization:        dta.Organization,
			Zone:                dta.Zone,
			Vpc:                 dta.Vpc,
			Subnet:              dta.Subnet,
			OracleSubnet:        dta.OracleSubnet,
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
			NetworkRoles:        dta.NetworkRoles,
			Isos:                dta.Isos,
			UsbDevices:          dta.UsbDevices,
			PciDevices:          dta.PciDevices,
			DriveDevices:        dta.DriveDevices,
			IscsiDevices:        dta.IscsiDevices,
			RootEnabled:         dta.RootEnabled,
			Vnc:                 dta.Vnc,
			Spice:               dta.Spice,
			Gui:                 dta.Gui,
			NoPublicAddress:     dta.NoPublicAddress,
			NoPublicAddress6:    dta.NoPublicAddress6,
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

		err = inst.Insert(db)
		if err != nil {
			utils.AbortWithError(c, 500, err)
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
	dta := &instanceMultiData{}

	err := c.Bind(dta)
	if err != nil {
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

	err = instance.UpdateMulti(db, dta.Ids, &doc)
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

	err = instance.Delete(db, instanceId)
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
	dta := []primitive.ObjectID{}

	err := c.Bind(&dta)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	force := c.Query("force")
	if force == "true" {
		for _, instId := range dta {
			err = instance.Remove(db, instId)
			if err != nil {
				utils.AbortWithError(c, 500, err)
				return
			}
		}
	} else {
		err = instance.DeleteMulti(db, dta)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
	}

	event.PublishDispatch(db, "instance.change")

	c.JSON(200, nil)
}

func instanceGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	instanceId, ok := utils.ParseObjectId(c.Param("instance_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	inst, err := instance.Get(db, instanceId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if demo.IsDemo() {
		inst.State = instance.Active
		inst.Action = instance.Start
		inst.VirtState = vm.Running
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
	db := c.MustGet("db").(*database.Database)

	ndeId, _ := utils.ParseObjectId(c.Query("node_names"))
	plId, _ := utils.ParseObjectId(c.Query("pool_names"))
	if !ndeId.IsZero() {
		query := &bson.M{
			"node": ndeId,
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

		ndeIds := []primitive.ObjectID{}

		for _, nde := range nodes {
			ndeIds = append(ndeIds, nde.Id)
		}

		query := &bson.M{
			"node": &bson.M{
				"$in": ndeIds,
			},
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

		query := bson.M{}

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

		networkRole := strings.TrimSpace(c.Query("network_role"))
		if networkRole != "" {
			if strings.HasPrefix(networkRole, "!") {
				networkRole = strings.TrimLeft(networkRole, "!")
				query["network_roles"] = &bson.M{
					"$ne": networkRole,
				}
			} else {
				query["network_roles"] = networkRole
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

		organization, ok := utils.ParseObjectId(c.Query("organization"))
		if ok {
			query["organization"] = organization
		}

		comment := strings.TrimSpace(c.Query("comment"))
		if comment != "" {
			query["comment"] = &bson.M{
				"$regex":   fmt.Sprintf(".*%s.*", comment),
				"$options": "i",
			}
		}

		instances, count, err := aggregate.GetInstancePaged(
			db, &query, page, pageCount)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		for _, inst := range instances {
			inst.Json(false)

			if demo.IsDemo() {
				inst.State = instance.Active
				inst.Action = instance.Start
				inst.VirtState = vm.Running
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

	instanceId, ok := utils.ParseObjectId(c.Param("instance_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	inst, err := instance.Get(db, instanceId)
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
