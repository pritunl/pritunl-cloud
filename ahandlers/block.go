package ahandlers

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/utils"
)

type blockData struct {
	Id       primitive.ObjectID `json:"id"`
	Name     string             `json:"name"`
	Comment  string             `json:"comment"`
	Type     string             `json:"type"`
	Subnets  []string           `json:"subnets"`
	Subnets6 []string           `json:"subnets6"`
	Excludes []string           `json:"excludes"`
	Netmask  string             `json:"netmask"`
	Gateway  string             `json:"gateway"`
	Gateway6 string             `json:"gateway6"`
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
	blck.Subnets = dta.Subnets
	blck.Subnets6 = dta.Subnets6
	blck.Excludes = dta.Excludes
	blck.Netmask = dta.Netmask
	blck.Gateway = dta.Gateway
	blck.Gateway6 = dta.Gateway6

	fields := set.NewSet(
		"name",
		"comment",
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

	err := block.Remove(db, blckId)
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

	blcks, err := block.GetAll(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, blcks)
}
