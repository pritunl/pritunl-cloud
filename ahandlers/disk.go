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
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/pritunl/pritunl-cloud/utils"
)

type diskData struct {
	Id               bson.ObjectID `json:"id"`
	Name             string        `json:"name"`
	Comment          string        `json:"comment"`
	Organization     bson.ObjectID `json:"organization"`
	Instance         bson.ObjectID `json:"instance"`
	Index            string        `json:"index"`
	Type             string        `json:"type"`
	Node             bson.ObjectID `json:"node"`
	Pool             bson.ObjectID `json:"pool"`
	DeleteProtection bool          `json:"delete_protection"`
	FileSystem       string        `json:"file_system"`
	Image            bson.ObjectID `json:"image"`
	RestoreImage     bson.ObjectID `json:"restore_image"`
	Backing          bool          `json:"backing"`
	Action           string        `json:"action"`
	Size             int           `json:"size"`
	LvSize           int           `json:"lv_size"`
	NewSize          int           `json:"new_size"`
	Backup           bool          `json:"backup"`
}

type disksMultiData struct {
	Ids    []bson.ObjectID `json:"ids"`
	Action string          `json:"action"`
}

type disksData struct {
	Disks []*aggregate.DiskAggregate `json:"disks"`
	Count int64                      `json:"count"`
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
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	dsk, err := disk.Get(db, diskId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	fields := set.NewSet(
		"name",
		"comment",
		"type",
		"instance",
		"delete_protection",
		"index",
		"backup",
		"new_size",
	)

	dsk.PreCommit()

	dsk.Name = dta.Name
	dsk.Comment = dta.Comment
	dsk.Instance = dta.Instance
	dsk.DeleteProtection = dta.DeleteProtection
	dsk.Index = dta.Index
	dsk.Backup = dta.Backup

	if dta.Action != "" && dsk.Action != "" {
		errData := &errortypes.ErrorData{
			Error:   "disk_actin_active",
			Message: "Disk action already active",
		}

		c.JSON(400, errData)
		return
	}

	if dsk.IsActive() && dta.Action == disk.Snapshot {
		dsk.Action = disk.Snapshot
		fields.Add("action")
	} else if dsk.IsActive() && dta.Action == disk.Backup {
		dsk.Action = disk.Backup
		fields.Add("action")
	} else if dsk.IsActive() && dta.Action == disk.Expand {
		dsk.Action = disk.Expand
		dsk.NewSize = dta.NewSize
		fields.Add("action")
	} else if dsk.IsActive() && dta.Action == disk.Restore {
		img, err := image.Get(db, dta.RestoreImage)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		if img.Disk != dsk.Id {
			errData := &errortypes.ErrorData{
				Error:   "invalid_restore_image",
				Message: "Invalid restore image",
			}

			c.JSON(400, errData)
			return
		}

		dsk.Action = disk.Restore
		dsk.RestoreImage = img.Id

		fields.Add("action")
		fields.Add("restore_image")
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
		Name: "new-disk",
	}

	err := c.Bind(dta)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	nde, err := node.Get(db, dta.Node)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	imgSystemType := ""
	imgSystemKind := ""
	if !dta.Image.IsZero() {
		img, err := image.GetOrgPublic(db, dta.Organization, dta.Image)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		imgSystemType = img.GetSystemType()
		imgSystemKind = img.GetSystemKind()

		store, err := storage.Get(db, img.Storage)
		if err != nil {
			utils.AbortWithError(c, 500, err)
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

	dsk := &disk.Disk{
		Name:             dta.Name,
		Comment:          dta.Comment,
		Organization:     dta.Organization,
		Instance:         dta.Instance,
		Datacenter:       nde.Datacenter,
		Zone:             nde.Zone,
		Index:            dta.Index,
		Type:             dta.Type,
		SystemType:       imgSystemType,
		SystemKind:       imgSystemKind,
		Node:             dta.Node,
		Pool:             dta.Pool,
		Image:            dta.Image,
		DeleteProtection: dta.DeleteProtection,
		FileSystem:       dta.FileSystem,
		Backing:          dta.Backing,
		Size:             dta.Size,
		LvSize:           dta.LvSize,
		Backup:           dta.Backup,
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

func disksPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &disksMultiData{}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	if data.Action != disk.Snapshot && data.Action != disk.Backup {
		errData := &errortypes.ErrorData{
			Error:   "invalid_action",
			Message: "Invalid disk action",
		}

		c.JSON(400, errData)
		return
	}

	doc := bson.M{
		"action": data.Action,
	}

	err = disk.UpdateMulti(db, data.Ids, &doc)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "disk.change")

	c.JSON(200, nil)
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

	dsk, err := disk.Get(db, diskId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if dsk.DeleteProtection {
		errData := &errortypes.ErrorData{
			Error:   "delete_protection",
			Message: "Cannot delete disk with delete protection",
		}

		c.JSON(400, errData)
		return
	}

	if !dsk.Instance.IsZero() {
		inst, e := instance.Get(db, dsk.Instance)
		if e != nil {
			err = e
			return
		}

		if inst.DeleteProtection {
			errData := &errortypes.ErrorData{
				Error: "instance_delete_protection",
				Message: "Cannot delete disk attached to " +
					"instance with delete protection",
			}

			c.JSON(400, errData)
			return
		}
	}

	err = disk.Delete(db, diskId)
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
	dta := []bson.ObjectID{}

	err := c.Bind(&dta)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	force := c.Query("force")
	if force == "true" {
		for _, diskId := range dta {
			err = disk.Remove(db, diskId)
			if err != nil {
				utils.AbortWithError(c, 500, err)
				return
			}
		}
	} else {
		err = disk.DeleteMulti(db, dta)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
	}

	event.PublishDispatch(db, "disk.change")

	c.JSON(200, nil)
}

func diskGet(c *gin.Context) {
	if demo.IsDemo() {
		dsk := demo.Disks[0]
		c.JSON(200, dsk)
		return
	}

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
	if demo.IsDemo() {
		data := &disksData{
			Disks: demo.Disks,
			Count: int64(len(demo.Disks)),
		}

		c.JSON(200, data)
		return
	}

	db := c.MustGet("db").(*database.Database)

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{}

	diskId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = diskId
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

	inst, ok := utils.ParseObjectId(c.Query("instance"))
	if ok {
		query["instance"] = inst
	}

	nodeId, ok := utils.ParseObjectId(c.Query("node"))
	if ok {
		query["node"] = nodeId
	}

	comment := strings.TrimSpace(c.Query("comment"))
	if comment != "" {
		query["comment"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", comment),
			"$options": "i",
		}
	}

	disks, count, err := aggregate.GetDiskPaged(db, &query, page, pageCount)
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
