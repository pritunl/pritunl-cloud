package uhandlers

import (
	"fmt"
	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/authority"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
)

type authorityData struct {
	Id           bson.ObjectId `json:"id"`
	Name         string        `json:"name"`
	Type         string        `json:"type"`
	NetworkRoles []string      `json:"network_roles"`
	Key          string        `json:"key"`
	Roles        []string      `json:"roles"`
	Certificate  string        `json:"certificate"`
}

type authoritiesData struct {
	Authorities []*authority.Authority `json:"authorities"`
	Count       int                    `json:"count"`
}

func authorityPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectId)
	data := &authorityData{}

	authorityId, ok := utils.ParseObjectId(c.Param("authority_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	fire, err := authority.GetOrg(db, userOrg, authorityId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	fire.Name = data.Name
	fire.Type = data.Type
	fire.NetworkRoles = data.NetworkRoles
	fire.Key = data.Key
	fire.Roles = data.Roles
	fire.Certificate = data.Certificate

	fields := set.NewSet(
		"state",
		"name",
		"type",
		"organization",
		"network_roles",
		"key",
		"roles",
		"certificate",
	)

	errData, err := fire.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = fire.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "authority.change")

	c.JSON(200, fire)
}

func authorityPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectId)
	data := &authorityData{
		Name: "New Authority",
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	fire := &authority.Authority{
		Name:         data.Name,
		Type:         data.Type,
		Organization: userOrg,
		NetworkRoles: data.NetworkRoles,
		Key:          data.Key,
		Roles:        data.Roles,
		Certificate:  data.Certificate,
	}

	errData, err := fire.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = fire.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "authority.change")

	c.JSON(200, fire)
}

func authorityDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectId)

	authorityId, ok := utils.ParseObjectId(c.Param("authority_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := authority.RemoveOrg(db, userOrg, authorityId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "authority.change")

	c.JSON(200, nil)
}

func authoritiesDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectId)
	data := []bson.ObjectId{}

	err := c.Bind(&data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	err = authority.RemoveMultiOrg(db, userOrg, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "authority.change")

	c.JSON(200, nil)
}

func authorityGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectId)

	authorityId, ok := utils.ParseObjectId(c.Param("authority_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	fire, err := authority.GetOrg(db, userOrg, authorityId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, fire)
}

func authoritiesGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectId)

	page, _ := strconv.Atoi(c.Query("page"))
	pageCount, _ := strconv.Atoi(c.Query("page_count"))

	query := bson.M{
		"organization": userOrg,
	}

	authrId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = authrId
	}

	name := strings.TrimSpace(c.Query("name"))
	if name != "" {
		query["name"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", name),
			"$options": "i",
		}
	}

	role := strings.TrimSpace(c.Query("role"))
	if role != "" {
		query["roles"] = role
	}

	networkRole := strings.TrimSpace(c.Query("network_role"))
	if networkRole != "" {
		query["network_roles"] = networkRole
	}

	authorities, count, err := authority.GetAllPaged(
		db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	data := &authoritiesData{
		Authorities: authorities,
		Count:       count,
	}

	c.JSON(200, data)
}
