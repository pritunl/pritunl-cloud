package ahandlers

import (
	"fmt"
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
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
)

type instanceData struct {
	Id               primitive.ObjectID `json:"id"`
	Organization     primitive.ObjectID `json:"organization"`
	Zone             primitive.ObjectID `json:"zone"`
	Vpc              primitive.ObjectID `json:"vpc"`
	Node             primitive.ObjectID `json:"node"`
	Image            primitive.ObjectID `json:"image"`
	ImageBacking     bool               `json:"image_backing"`
	Domain           primitive.ObjectID `json:"domain"`
	Name             string             `json:"name"`
	State            string             `json:"state"`
	DeleteProtection bool               `json:"delete_protection"`
	InitDiskSize     int                `json:"init_disk_size"`
	Memory           int                `json:"memory"`
	Processors       int                `json:"processors"`
	NetworkRoles     []string           `json:"network_roles"`
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
	inst.Vpc = dta.Vpc
	if dta.State != "" {
		inst.State = dta.State
	}
	inst.DeleteProtection = dta.DeleteProtection
	inst.Memory = dta.Memory
	inst.Processors = dta.Processors
	inst.NetworkRoles = dta.NetworkRoles
	inst.Vnc = dta.Vnc
	inst.Domain = dta.Domain
	inst.NoPublicAddress = dta.NoPublicAddress
	inst.NoHostAddress = dta.NoHostAddress

	fields := set.NewSet(
		"name",
		"vpc",
		"state",
		"restart",
		"restart_block_ip",
		"delete_protection",
		"memory",
		"processors",
		"network_roles",
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
	dta := &instanceData{
		Name: "New Instance",
	}

	err := c.Bind(dta)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

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
			Organization:     dta.Organization,
			Zone:             dta.Zone,
			Vpc:              dta.Vpc,
			Node:             dta.Node,
			Image:            dta.Image,
			ImageBacking:     dta.ImageBacking,
			DeleteProtection: dta.DeleteProtection,
			Name:             name,
			InitDiskSize:     dta.InitDiskSize,
			Memory:           dta.Memory,
			Processors:       dta.Processors,
			NetworkRoles:     dta.NetworkRoles,
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
	dta := &instanceMultiData{}

	err := c.Bind(dta)
	if err != nil {
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
	}

	c.JSON(200, inst)
}

func instancesGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	nde, _ := utils.ParseObjectId(c.Query("node_names"))
	if !nde.IsZero() {
		query := &bson.M{
			"node": nde,
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
				"$regex":   fmt.Sprintf(".*%s.*", name),
				"$options": "i",
			}
		}

		networkRole := strings.TrimSpace(c.Query("network_role"))
		if networkRole != "" {
			query["network_roles"] = networkRole
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

		organization, ok := utils.ParseObjectId(c.Query("organization"))
		if ok {
			query["organization"] = organization
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
			}
		}

		dta := &instancesData{
			Instances: instances,
			Count:     count,
		}

		c.JSON(200, dta)
	}
}
