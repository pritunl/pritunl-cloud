package ahandlers

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
)

type blockData struct {
	Id        bson.ObjectId `json:"id"`
	Name      string        `json:"name"`
	Addresses []string      `json:"addresses"`
	Excludes  []string      `json:"excludes"`
	Netmask   string        `json:"netmask"`
	Gateway   string        `json:"gateway"`
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
		utils.AbortWithError(c, 500, err)
		return
	}

	blck, err := block.Get(db, blckId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	blck.Name = dta.Name
	blck.Addresses = dta.Addresses
	blck.Excludes = dta.Excludes
	blck.Netmask = dta.Netmask
	blck.Gateway = dta.Gateway

	fields := set.NewSet(
		"name",
		"addresses",
		"excludes",
		"netmask",
		"gateway",
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
		utils.AbortWithError(c, 500, err)
		return
	}

	blck := &block.Block{
		Name:      dta.Name,
		Addresses: dta.Addresses,
		Excludes:  dta.Excludes,
		Netmask:   dta.Netmask,
		Gateway:   dta.Gateway,
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
