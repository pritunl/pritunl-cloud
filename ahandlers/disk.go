package ahandlers

import (
	"fmt"
	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
)

type diskData struct {
	Id           bson.ObjectId `json:"id"`
	Name         string        `json:"name"`
	Organization bson.ObjectId `json:"organization"`
	Instance     bson.ObjectId `json:"instance"`
	Index        string        `json:"index"`
	Node         bson.ObjectId `json:"node"`
	Image        bson.ObjectId `json:"image"`
	Size         int           `json:"size"`
}

type disksData struct {
	Disks []*disk.Disk `json:"disks"`
	Count int          `json:"count"`
}

func diskPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	dta := &diskData{}

	diskId, ok := utils.ParseObjectId(c.Param("disk_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(dta)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	dsk, err := disk.Get(db, diskId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	dsk.Name = dta.Name
	dsk.Organization = dta.Organization
	dsk.Instance = dta.Instance
	dsk.Index = dta.Index

	fields := set.NewSet(
		"state",
		"name",
		"organization",
		"instance",
		"index",
	)

	errData, err := dsk.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = dsk.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "disk.change")

	c.JSON(200, dsk)
}

func diskPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	dta := &diskData{
		Name: "New Disk",
	}

	err := c.Bind(dta)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	dsk := &disk.Disk{
		Name:         dta.Name,
		Organization: dta.Organization,
		Instance:     dta.Instance,
		Index:        dta.Index,
		Node:         dta.Node,
		Image:        dta.Image,
		Size:         dta.Size,
	}

	errData, err := dsk.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = dsk.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "disk.change")

	c.JSON(200, dsk)
}

func diskDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	diskId, ok := utils.ParseObjectId(c.Param("disk_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := disk.Delete(db, diskId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "disk.change")

	c.JSON(200, nil)
}

func disksDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	dta := []bson.ObjectId{}

	err := c.Bind(&dta)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	err = disk.DeleteMulti(db, dta)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "disk.change")

	c.JSON(200, nil)
}

func diskGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	diskId, ok := utils.ParseObjectId(c.Param("disk_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	dsk, err := disk.Get(db, diskId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, dsk)
}

func disksGet(c *gin.Context) {
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

	disks, count, err := disk.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	dta := &disksData{
		Disks: disks,
		Count: count,
	}

	c.JSON(200, dta)
}
