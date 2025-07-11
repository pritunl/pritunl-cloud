package ahandlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/relations"
	"github.com/pritunl/pritunl-cloud/utils"
)

type blockData struct {
	Id       primitive.ObjectID `json:"id"`
	Name     string             `json:"name"`
	Comment  string             `json:"comment"`
	Vlan     int                `json:"vlan"`
	Type     string             `json:"type"`
	Subnets  []string           `json:"subnets"`
	Subnets6 []string           `json:"subnets6"`
	Excludes []string           `json:"excludes"`
	Netmask  string             `json:"netmask"`
	Gateway  string             `json:"gateway"`
	Gateway6 string             `json:"gateway6"`
}

type blocksData struct {
	Blocks []*aggregate.BlockAggregate `json:"blocks"`
	Count  int64                       `json:"count"`
}

func blockPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	dta := &blockData{}

	blckId, ok := utils.ParseObjectId(c.Param("block_id"))
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

	blck, err := block.Get(db, blckId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	blck.Name = dta.Name
	blck.Comment = dta.Comment
	blck.Vlan = dta.Vlan
	blck.Subnets = dta.Subnets
	blck.Subnets6 = dta.Subnets6
	blck.Excludes = dta.Excludes
	blck.Netmask = dta.Netmask
	blck.Gateway = dta.Gateway
	blck.Gateway6 = dta.Gateway6

	fields := set.NewSet(
		"name",
		"comment",
		"vlan",
		"subnets",
		"subnets6",
		"excludes",
		"netmask",
		"gateway",
		"gateway6",
	)

	errData, err := blck.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = blck.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "block.change")

	c.JSON(200, blck)
}

func blockPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	dta := &blockData{
		Name: "New IP Block",
	}

	err := c.Bind(dta)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	blck := &block.Block{
		Name:     dta.Name,
		Comment:  dta.Comment,
		Vlan:     dta.Vlan,
		Type:     dta.Type,
		Subnets:  dta.Subnets,
		Subnets6: dta.Subnets6,
		Excludes: dta.Excludes,
		Netmask:  dta.Netmask,
		Gateway:  dta.Gateway,
		Gateway6: dta.Gateway6,
	}

	errData, err := blck.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = blck.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "block.change")

	c.JSON(200, blck)
}

func blockDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	blckId, ok := utils.ParseObjectId(c.Param("block_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	errData, err := relations.CanDelete(db, "block", blckId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = block.Remove(db, blckId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "block.change")

	c.JSON(200, nil)
}

func blocksDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := []primitive.ObjectID{}

	err := c.Bind(&data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	errData, err := relations.CanDeleteAll(db, "block", data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = block.RemoveMulti(db, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "block.change")

	c.JSON(200, nil)
}

func blockGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	blckId, ok := utils.ParseObjectId(c.Param("block_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	blck, err := block.Get(db, blckId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, blck)
}

func blocksGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{}

	blockId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = blockId
	}

	name := strings.TrimSpace(c.Query("name"))
	if name != "" {
		query["name"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", regexp.QuoteMeta(name)),
			"$options": "i",
		}
	}

	comment := strings.TrimSpace(c.Query("comment"))
	if comment != "" {
		query["comment"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", comment),
			"$options": "i",
		}
	}

	blcks, count, err := aggregate.GetBlockPaged(db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	data := &blocksData{
		Blocks: blcks,
		Count:  count,
	}

	c.JSON(200, data)
}
