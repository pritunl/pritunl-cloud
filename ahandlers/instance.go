package ahandlers

import (
	"fmt"
	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
)

type instanceData struct {
	Id           bson.ObjectId `json:"id"`
	Organization bson.ObjectId `json:"organization"`
	Zone         bson.ObjectId `json:"zone"`
	Node         bson.ObjectId `json:"node"`
	Name         string        `json:"name"`
	Memory       int           `json:"memory"`
	Processors   int           `json:"processors"`
}

type instancesData struct {
	Instances []*instance.Instance `json:"instances"`
	Count     int                  `json:"count"`
}

func instancePut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &instanceData{}

	instanceId, ok := utils.ParseObjectId(c.Param("instance_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	nde, err := instance.Get(db, instanceId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	nde.Name = data.Name
	nde.Memory = data.Memory
	nde.Processors = data.Processors

	fields := set.NewSet(
		"name",
		"memory",
		"processors",
	)

	errData, err := nde.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = nde.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "instance.change")

	c.JSON(200, nde)
}

func instancePost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &instanceData{
		Name: "New Instance",
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	dc := &instance.Instance{
		Organization: data.Organization,
		Zone:         data.Zone,
		Node:         data.Node,
		Name:         data.Name,
		Memory:       data.Memory,
		Processors:   data.Processors,
	}

	errData, err := dc.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = dc.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "instance.change")

	c.JSON(200, dc)
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

	err := instance.Remove(db, instanceId)
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
	data := []bson.ObjectId{}

	err := c.Bind(&data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	err = instance.RemoveMulti(db, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
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

	nde, err := instance.Get(db, instanceId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, nde)
}

func instancesGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	page, _ := strconv.Atoi(c.Query("page"))
	pageCount, _ := strconv.Atoi(c.Query("page_count"))

	query := bson.M{}

	name := strings.TrimSpace(c.Query("name"))
	if name != "" {
		query["name"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", name),
			"$options": "i",
		}
	}

	instances, count, err := instance.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	data := &instancesData{
		Instances: instances,
		Count:     count,
	}

	c.JSON(200, data)
}
