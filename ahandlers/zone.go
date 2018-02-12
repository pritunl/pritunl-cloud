package ahandlers

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/zone"
	"gopkg.in/mgo.v2/bson"
)

type zoneData struct {
	Id            bson.ObjectId   `json:"id"`
	Datacenter    bson.ObjectId   `json:"datacenter"`
	Organizations []bson.ObjectId `json:"organizations"`
	Name          string          `json:"name"`
}

func zonePut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &zoneData{}

	zoneId, ok := utils.ParseObjectId(c.Param("zone_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	zone, err := zone.Get(db, zoneId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	zone.Name = data.Name
	zone.Organizations = data.Organizations

	fields := set.NewSet(
		"name",
		"organizations",
	)

	errData, err := zone.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = zone.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "zone.change")

	c.JSON(200, zone)
}

func zonePost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &zoneData{
		Name: "New Zone",
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	zone := &zone.Zone{
		Datacenter:    data.Datacenter,
		Organizations: data.Organizations,
		Name:          data.Name,
	}

	errData, err := zone.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = zone.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "zone.change")

	c.JSON(200, zone)
}

func zoneDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	zoneId, ok := utils.ParseObjectId(c.Param("zone_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := zone.Remove(db, zoneId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "zone.change")

	c.JSON(200, nil)
}

func zoneGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	zoneId, ok := utils.ParseObjectId(c.Param("zone_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	zone, err := zone.Get(db, zoneId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, zone)
}

func zonesGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	zones, err := zone.GetAll(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, zones)
}
