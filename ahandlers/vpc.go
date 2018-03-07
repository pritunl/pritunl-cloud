package ahandlers

import (
	"fmt"
	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vpc"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
)

type vpcData struct {
	Id           bson.ObjectId `json:"id"`
	Name         string        `json:"name"`
	Network      string        `json:"network"`
	Organization bson.ObjectId `json:"organization"`
	Datacenter   bson.ObjectId `json:"datacenter"`
}

type vpcsData struct {
	Vpcs  []*vpc.Vpc `json:"vpcs"`
	Count int        `json:"count"`
}

func vpcPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &vpcData{}

	vpcId, ok := utils.ParseObjectId(c.Param("vpc_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	vc, err := vpc.Get(db, vpcId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	vc.Name = data.Name
	vc.Network = data.Network
	vc.Organization = data.Organization
	vc.Datacenter = data.Datacenter

	fields := set.NewSet(
		"state",
		"name",
		"network",
		"organization",
		"datacenter",
	)

	errData, err := vc.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = vc.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "vpc.change")

	c.JSON(200, vc)
}

func vpcPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &vpcData{
		Name: "New Vpc",
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	vc := &vpc.Vpc{
		Name:         data.Name,
		Network:      data.Network,
		Organization: data.Organization,
		Datacenter:   data.Datacenter,
	}

	vc.GenerateVpcId()

	errData, err := vc.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = vc.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "vpc.change")

	c.JSON(200, vc)
}

func vpcDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	vpcId, ok := utils.ParseObjectId(c.Param("vpc_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := vpc.Remove(db, vpcId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "vpc.change")

	c.JSON(200, nil)
}

func vpcsDelete(c *gin.Context) {
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

	err = vpc.RemoveMulti(db, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "vpc.change")

	c.JSON(200, nil)
}

func vpcGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	vpcId, ok := utils.ParseObjectId(c.Param("vpc_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	vc, err := vpc.Get(db, vpcId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, vc)
}

func vpcsGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	if c.Query("names") == "true" {
		query := &bson.M{}

		vpcs, err := vpc.GetAllNames(db, query)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		c.JSON(200, vpcs)
	} else {
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

		vpcs, count, err := vpc.GetAllPaged(db, &query, page, pageCount)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		data := &vpcsData{
			Vpcs:  vpcs,
			Count: count,
		}

		c.JSON(200, data)
	}
}
