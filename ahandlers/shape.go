package ahandlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/relations"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/utils"
)

type shapeData struct {
	Id               primitive.ObjectID `json:"id"`
	Name             string             `json:"name"`
	Comment          string             `json:"comment"`
	Type             string             `json:"type"`
	DeleteProtection bool               `json:"delete_protection"`
	Datacenter       primitive.ObjectID `json:"datacenter"`
	Roles            []string           `json:"roles"`
	Flexible         bool               `json:"flexible"`
	DiskType         string             `json:"disk_type"`
	DiskPool         primitive.ObjectID `json:"disk_pool"`
	Memory           int                `json:"memory"`
	Processors       int                `json:"processors"`
}

type shapesData struct {
	Shapes []*shape.Shape `json:"shapes"`
	Count  int64          `json:"count"`
}

func shapePut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &shapeData{}

	shapeId, ok := utils.ParseObjectId(c.Param("shape_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	shpe, err := shape.Get(db, shapeId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	shpe.Name = data.Name
	shpe.Type = data.Type
	shpe.Comment = data.Comment
	shpe.DeleteProtection = data.DeleteProtection
	shpe.Datacenter = data.Datacenter
	shpe.Roles = data.Roles
	shpe.Flexible = data.Flexible
	shpe.DiskType = data.DiskType
	shpe.DiskPool = data.DiskPool
	shpe.Memory = data.Memory
	shpe.Processors = data.Processors

	fields := set.NewSet(
		"name",
		"type",
		"comment",
		"delete_protection",
		"datacenter",
		"roles",
		"flexible",
		"disk_type",
		"disk_pool",
		"memory",
		"processors",
	)

	errData, err := shpe.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = shpe.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "shape.change")

	c.JSON(200, shpe)
}

func shapePost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &shapeData{
		Name: "New Shape",
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	shpe := &shape.Shape{
		Name:             data.Name,
		Comment:          data.Comment,
		DeleteProtection: data.DeleteProtection,
		Datacenter:       data.Datacenter,
		Roles:            data.Roles,
		Flexible:         data.Flexible,
		DiskType:         data.DiskType,
		DiskPool:         data.DiskPool,
		Memory:           data.Memory,
		Processors:       data.Processors,
	}

	errData, err := shpe.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = shpe.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "shape.change")

	c.JSON(200, shpe)
}

func shapeDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	shapeId, ok := utils.ParseObjectId(c.Param("shape_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	errData, err := relations.CanDelete(db, "shape", shapeId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = shape.Remove(db, shapeId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "shape.change")

	c.JSON(200, nil)
}

func shapesDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := []primitive.ObjectID{}

	err := c.Bind(&data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	errData, err := relations.CanDeleteAll(db, "shape", data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = shape.RemoveMulti(db, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "shape.change")

	c.JSON(200, nil)
}

func shapeGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	shapeId, ok := utils.ParseObjectId(c.Param("shape_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	shpe, err := shape.Get(db, shapeId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, shpe)
}

func shapesGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{}

	shapeId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = shapeId
	}

	name := strings.TrimSpace(c.Query("name"))
	if name != "" {
		query["name"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", regexp.QuoteMeta(name)),
			"$options": "i",
		}
	}

	role := strings.TrimSpace(c.Query("role"))
	if role != "" {
		query["roles"] = role
	}

	organization, ok := utils.ParseObjectId(c.Query("organization"))
	if ok {
		query["organization"] = organization
	}

	comment := strings.TrimSpace(c.Query("comment"))
	if comment != "" {
		query["comment"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", comment),
			"$options": "i",
		}
	}

	shapes, count, err := aggregate.GetShapePaged(db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	data := &shapesData{
		Shapes: shapes,
		Count:  count,
	}

	c.JSON(200, data)
}
