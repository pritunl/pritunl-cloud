package uhandlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/domain"
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
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
)

type instanceData struct {
	Id               primitive.ObjectID `json:"id"`
	Zone             primitive.ObjectID `json:"zone"`
	Vpc              primitive.ObjectID `json:"vpc"`
	Subnet           primitive.ObjectID `json:"subnet"`
	OracleSubnet     string             `json:"oracle_subnet"`
	Node             primitive.ObjectID `json:"node"`
	Image            primitive.ObjectID `json:"image"`
	ImageBacking     bool               `json:"image_backing"`
	Domain           primitive.ObjectID `json:"domain"`
	Name             string             `json:"name"`
	Comment          string             `json:"comment"`
	State            string             `json:"state"`
	Uefi             bool               `json:"uefi"`
	SecureBoot       bool               `json:"secure_boot"`
	DeleteProtection bool               `json:"delete_protection"`
	InitDiskSize     int                `json:"init_disk_size"`
	Memory           int                `json:"memory"`
	Processors       int                `json:"processors"`
	NetworkRoles     []string           `json:"network_roles"`
	Isos             []*iso.Iso         `json:"isos"`
	UsbDevices       []*usb.Device      `json:"usb_devices"`
	PciDevices       []*pci.Device      `json:"pci_devices"`
	DriveDevices     []*drive.Device    `json:"drive_devices"`
	IscsiDevices     []*iscsi.Device    `json:"iscsi_devices"`
	Vnc              bool               `json:"vnc"`
	NoPublicAddress  bool               `json:"no_public_address"`
	NoHostAddress    bool               `json:"no_host_address"`
	Count            int                `json:"count"`
}

type instanceMultiData struct {
	Ids   []primitive.ObjectID `json:"ids"`
	State string               `json:"state"`
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
	userOrg := c.MustGet("organization").(primitive.ObjectID)
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

	if !dta.Domain.IsZero() {
		exists, err := domain.ExistsOrg(db, userOrg, dta.Domain)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
		if !exists {
			utils.AbortWithStatus(c, 405)
			return
		}
	}

	inst.PreCommit()

	inst.Name = dta.Name
	inst.Comment = dta.Comment
	inst.Vpc = dta.Vpc
	inst.Subnet = dta.Subnet
	inst.OracleSubnet = dta.OracleSubnet
	if dta.State != "" {
		inst.State = dta.State
	}
	inst.Uefi = dta.Uefi
	inst.SecureBoot = dta.SecureBoot
	inst.DeleteProtection = dta.DeleteProtection
	inst.Memory = dta.Memory
	inst.Processors = dta.Processors
	inst.NetworkRoles = dta.NetworkRoles
	inst.Isos = dta.Isos
	inst.UsbDevices = dta.UsbDevices
	inst.PciDevices = dta.PciDevices
	inst.DriveDevices = dta.DriveDevices
	inst.IscsiDevices = dta.IscsiDevices
	inst.Vnc = dta.Vnc
	inst.Domain = dta.Domain
	inst.NoPublicAddress = dta.NoPublicAddress
	inst.NoHostAddress = dta.NoHostAddress

	fields := set.NewSet(
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
		"delete_protection",
		"memory",
		"processors",
		"network_roles",
		"isos",
		"usb_devices",
		"pci_devices",
		"drive_devices",
		"iscsi_devices",
		"vnc",
		"vnc_display",
		"vnc_password",
		"domain",
		"no_public_address",
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
	userOrg := c.MustGet("organization").(primitive.ObjectID)
	dta := &instanceData{
		Name: "New Instance",
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

	nde, err := node.Get(db, dta.Node)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if nde.Zone != zne.Id {
		utils.AbortWithStatus(c, 405)
		return
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

	if !dta.Domain.IsZero() {
		exists, err := domain.ExistsOrg(db, userOrg, dta.Domain)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
		if !exists {
			utils.AbortWithStatus(c, 405)
			return
		}
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

		store, err := storage.Get(db, img.Storage)
		if err != nil {
			return
		}

		available, err := data.ImageAvailable(store, img)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
		if !available {
			if store.IsOracle() {
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
			State:            dta.State,
			Organization:     userOrg,
			Zone:             dta.Zone,
			Vpc:              dta.Vpc,
			Subnet:           dta.Subnet,
			OracleSubnet:     dta.OracleSubnet,
			Node:             dta.Node,
			Image:            dta.Image,
			ImageBacking:     dta.ImageBacking,
			DeleteProtection: dta.DeleteProtection,
			Name:             name,
			Comment:          dta.Comment,
			InitDiskSize:     dta.InitDiskSize,
			Memory:           dta.Memory,
			Processors:       dta.Processors,
			NetworkRoles:     dta.NetworkRoles,
			Isos:             dta.Isos,
			UsbDevices:       dta.UsbDevices,
			PciDevices:       dta.PciDevices,
			DriveDevices:     dta.DriveDevices,
			IscsiDevices:     dta.IscsiDevices,
			Vnc:              dta.Vnc,
			Domain:           dta.Domain,
			NoPublicAddress:  dta.NoPublicAddress,
			NoHostAddress:    dta.NoHostAddress,
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
	userOrg := c.MustGet("organization").(primitive.ObjectID)
	dta := &instanceMultiData{}

	err := c.Bind(dta)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	doc := bson.M{
		"state": dta.State,
	}

	if dta.State != instance.Start {
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
	userOrg := c.MustGet("organization").(primitive.ObjectID)

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
	userOrg := c.MustGet("organization").(primitive.ObjectID)
	dta := []primitive.ObjectID{}

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
	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)

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
		inst.State = instance.Start
		inst.VmState = vm.Running
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
	userOrg := c.MustGet("organization").(primitive.ObjectID)

	nde, _ := utils.ParseObjectId(c.Query("node_names"))
	if !nde.IsZero() {
		query := &bson.M{
			"node":         nde,
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
				"$regex":   fmt.Sprintf(".*%s.*", name),
				"$options": "i",
			}
		}

		networkRole := strings.TrimSpace(c.Query("network_role"))
		if networkRole != "" {
			query["network_roles"] = networkRole
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

		instances, count, err := aggregate.GetInstancePaged(
			db, &query, page, pageCount)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		for _, inst := range instances {
			inst.Json()

			if demo.IsDemo() {
				inst.State = instance.Start
				inst.VmState = vm.Running
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
	userOrg := c.MustGet("organization").(primitive.ObjectID)

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
