package ahandlers

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/zone"
)

type zoneData struct {
	Id         primitive.ObjectID `json:"id"`
	Datacenter primitive.ObjectID `json:"datacenter"`
	Name       string             `json:"name"`
	Comment    string             `json:"comment"`
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

	zne, err := zone.Get(db, zoneId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	zne.Name = data.Name
	zne.Comment = data.Comment

	fields := set.NewSet(
		"name",
		"comment",
	)

	errData, err := zne.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = zne.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "zone.change")

	c.JSON(200, zne)
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

	zne := &zone.Zone{
		Datacenter: data.Datacenter,
		Name:       data.Name,
		Comment:    data.Comment,
	}

	errData, err := zne.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = zne.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "zone.change")

	c.JSON(200, zne)
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

	zne, err := zone.Get(db, zoneId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, zne)
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
