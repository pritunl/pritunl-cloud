package ahandlers

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
)

type datacenterData struct {
	Id                 bson.ObjectId   `json:"id"`
	Name               string          `json:"name"`
	MatchOrganizations bool            `json:"match_organizations"`
	Organizations      []bson.ObjectId `json:"organizations"`
	PublicStorages     []bson.ObjectId `json:"public_storages"`
	PrivateStorage     bson.ObjectId   `json:"private_storage"`
}

func datacenterPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &datacenterData{}

	dcId, ok := utils.ParseObjectId(c.Param("dc_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	dc, err := datacenter.Get(db, dcId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	dc.Name = data.Name
	dc.MatchOrganizations = data.MatchOrganizations
	dc.Organizations = data.Organizations
	dc.PublicStorages = data.PublicStorages
	dc.PrivateStorage = data.PrivateStorage

	fields := set.NewSet(
		"name",
		"match_organizations",
		"organizations",
		"public_storages",
		"private_storage",
	)

	errData, err := dc.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = dc.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "datacenter.change")

	c.JSON(200, dc)
}

func datacenterPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &datacenterData{
		Name: "New Datacenter",
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	dc := &datacenter.Datacenter{
		Name:               data.Name,
		MatchOrganizations: data.MatchOrganizations,
		Organizations:      data.Organizations,
		PublicStorages:     data.PublicStorages,
		PrivateStorage:     data.PrivateStorage,
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

	event.PublishDispatch(db, "datacenter.change")

	c.JSON(200, dc)
}

func datacenterDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	dcId, ok := utils.ParseObjectId(c.Param("dc_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := datacenter.Remove(db, dcId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "datacenter.change")

	c.JSON(200, nil)
}

func datacenterGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	dcId, ok := utils.ParseObjectId(c.Param("dc_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	dc, err := datacenter.Get(db, dcId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, dc)
}

func datacentersGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	dcs, err := datacenter.GetAll(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, dcs)
}
