package ahandlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/relations"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/zone"
)

type zoneData struct {
	Id         bson.ObjectID `json:"id"`
	Datacenter bson.ObjectID `json:"datacenter"`
	Name       string        `json:"name"`
	Comment    string        `json:"comment"`
}

type zonesData struct {
	Zones []*zone.Zone `json:"zones"`
	Count int64        `json:"count"`
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
		Name: "new-zone",
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

	errData, err := relations.CanDelete(db, "zone", zoneId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = zone.Remove(db, zoneId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "zone.change")

	c.JSON(200, nil)
}

func zonesDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := []bson.ObjectID{}

	err := c.Bind(&data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	errData, err := relations.CanDeleteAll(db, "zone", data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = zone.RemoveMulti(db, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "zone.change")

	c.JSON(200, nil)
}

func zoneGet(c *gin.Context) {
	if demo.IsDemo() {
		zne := demo.Zones[0]
		c.JSON(200, zne)
		return
	}

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
	if demo.IsDemo() {
		data := &zonesData{
			Zones: demo.Zones,
			Count: int64(len(demo.Zones)),
		}

		c.JSON(200, data)
		return
	}

	db := c.MustGet("db").(*database.Database)

	if c.Query("names") == "true" {
		dcs, err := zone.GetAllNames(db, &bson.M{})
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		c.JSON(200, dcs)
		return
	}

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{}

	zoneId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = zoneId
	}

	name := strings.TrimSpace(c.Query("name"))
	if name != "" {
		query["name"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", regexp.QuoteMeta(name)),
			"$options": "i",
		}
	}

	organization, ok := utils.ParseObjectId(c.Query("organization"))
	if ok {
		query["organization"] = organization
	}

	datacenter, ok := utils.ParseObjectId(c.Query("datacenter"))
	if ok {
		query["datacenter"] = datacenter
	}

	comment := strings.TrimSpace(c.Query("comment"))
	if comment != "" {
		query["comment"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", comment),
			"$options": "i",
		}
	}

	znes, count, err := zone.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	data := &zonesData{
		Zones: znes,
		Count: count,
	}

	c.JSON(200, data)
}
