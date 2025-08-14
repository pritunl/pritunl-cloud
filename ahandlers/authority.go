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
	"github.com/pritunl/pritunl-cloud/authority"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/utils"
)

type authorityData struct {
	Id           primitive.ObjectID `json:"id"`
	Name         string             `json:"name"`
	Comment      string             `json:"comment"`
	Type         string             `json:"type"`
	Organization primitive.ObjectID `json:"organization"`
	Roles        []string           `json:"roles"`
	Key          string             `json:"key"`
	Principals   []string           `json:"principals"`
	Certificate  string             `json:"certificate"`
}

type authoritiesData struct {
	Authorities []*authority.Authority `json:"authorities"`
	Count       int64                  `json:"count"`
}

func authorityPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &authorityData{}

	authorityId, ok := utils.ParseObjectId(c.Param("authority_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	authr, err := authority.Get(db, authorityId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	authr.Name = data.Name
	authr.Comment = data.Comment
	authr.Type = data.Type
	authr.Organization = data.Organization
	authr.Roles = data.Roles
	authr.Key = data.Key
	authr.Principals = data.Principals
	authr.Certificate = data.Certificate

	fields := set.NewSet(
		"name",
		"comment",
		"type",
		"organization",
		"roles",
		"key",
		"principals",
		"certificate",
	)

	errData, err := authr.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = authr.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "authority.change")

	c.JSON(200, authr)
}

func authorityPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &authorityData{
		Name: "New Authority",
	}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	authr := &authority.Authority{
		Name:         data.Name,
		Comment:      data.Comment,
		Type:         data.Type,
		Organization: data.Organization,
		Roles:        data.Roles,
		Key:          data.Key,
		Principals:   data.Principals,
		Certificate:  data.Certificate,
	}

	errData, err := authr.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = authr.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "authority.change")

	c.JSON(200, authr)
}

func authorityDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	authorityId, ok := utils.ParseObjectId(c.Param("authority_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := authority.Remove(db, authorityId)
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
	data := []primitive.ObjectID{}

	err := c.Bind(&data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	err = authority.RemoveMulti(db, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "authority.change")

	c.JSON(200, nil)
}

func authorityGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	authorityId, ok := utils.ParseObjectId(c.Param("authority_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	authr, err := authority.Get(db, authorityId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, authr)
}

func authoritiesGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{}

	authrId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = authrId
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

	principal := strings.TrimSpace(c.Query("principal"))
	if principal != "" {
		query["principals"] = principal
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
