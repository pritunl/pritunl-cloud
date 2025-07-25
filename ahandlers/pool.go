package ahandlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/zone"

	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/utils"
)

type poolData struct {
	Id               primitive.ObjectID `json:"id"`
	Name             string             `json:"name"`
	Comment          string             `json:"comment"`
	DeleteProtection bool               `json:"delete_protection"`
	Zone             primitive.ObjectID `json:"zone"`
	Type             string             `json:"type"`
	VgName           string             `json:"vg_name"`
}

type poolsData struct {
	Pools []*pool.Pool `json:"pools"`
	Count int64        `json:"count"`
}

func poolPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &poolData{}

	poolId, ok := utils.ParseObjectId(c.Param("pool_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	pl, err := pool.Get(db, poolId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	pl.Name = data.Name
	pl.Comment = data.Comment
	pl.DeleteProtection = data.DeleteProtection
	pl.Type = data.Type

	fields := set.NewSet(
		"name",
		"comment",
		"delete_protection",
		"type",
	)

	errData, err := pl.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = pl.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "pool.change")

	c.JSON(200, pl)
}

func poolPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &poolData{
		Name: "New Pool",
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	zne, err := zone.Get(db, data.Zone)
	if err != nil {
		return
	}

	pl := &pool.Pool{
		Name:             data.Name,
		Comment:          data.Comment,
		DeleteProtection: data.DeleteProtection,
		Datacenter:       zne.Datacenter,
		Zone:             data.Zone,
		Type:             data.Type,
		VgName:           data.VgName,
	}

	errData, err := pl.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = pl.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "pool.change")

	c.JSON(200, pl)
}

func poolDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	poolId, ok := utils.ParseObjectId(c.Param("pool_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := pool.Remove(db, poolId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "pool.change")

	c.JSON(200, nil)
}

func poolsDelete(c *gin.Context) {
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

	err = pool.RemoveMulti(db, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "pool.change")

	c.JSON(200, nil)
}

func poolGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	poolId, ok := utils.ParseObjectId(c.Param("pool_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	pl, err := pool.Get(db, poolId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, pl)
}

func poolsGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	nodeNames, err := node.GetAllNamesMap(db, &bson.M{})
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{}

	poolId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = poolId
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

	pools, count, err := pool.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	for _, pl := range pools {
		pl.Json(nodeNames)
	}

	data := &poolsData{
		Pools: pools,
		Count: count,
	}

	c.JSON(200, data)
}
