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
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/utils"
)

type secretData struct {
	Id           primitive.ObjectID `json:"id"`
	Name         string             `json:"name"`
	Comment      string             `json:"comment"`
	Organization primitive.ObjectID `json:"organization"`
	Type         string             `json:"type"`
	Key          string             `json:"key"`
	Value        string             `json:"value"`
	Region       string             `json:"region"`
}

type secretsData struct {
	Secrets []*secret.Secret `json:"secrets"`
	Count   int64            `json:"count"`
}

func secretPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &secretData{}

	secrId, ok := utils.ParseObjectId(c.Param("secr_id"))
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

	secr, err := secret.Get(db, secrId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	secr.Name = data.Name
	secr.Comment = data.Comment
	secr.Organization = data.Organization
	secr.Type = data.Type
	secr.Key = data.Key
	secr.Value = data.Value
	secr.Region = data.Region

	fields := set.NewSet(
		"name",
		"comment",
		"organization",
		"type",
		"key",
		"value",
		"region",
		"public_key",
		"private_key",
	)

	errData, err := secr.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = secr.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "secret.change")

	c.JSON(200, secr)
}

func secretPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &secretData{
		Name: "New Secret",
	}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	secr := &secret.Secret{
		Name:         data.Name,
		Comment:      data.Comment,
		Organization: data.Organization,
		Type:         data.Type,
		Key:          data.Key,
		Value:        data.Value,
		Region:       data.Region,
	}

	errData, err := secr.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = secr.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "secret.change")

	c.JSON(200, secr)
}

func secretDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	secrId, ok := utils.ParseObjectId(c.Param("secr_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := secret.Remove(db, secrId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "secret.change")

	c.JSON(200, nil)
}

func secretsDelete(c *gin.Context) {
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

	err = secret.RemoveMulti(db, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "secret.change")

	c.JSON(200, nil)
}

func secretGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	secrId, ok := utils.ParseObjectId(c.Param("secr_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	secr, err := secret.Get(db, secrId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if demo.IsDemo() {
		secr.Key = "demo"
		secr.Value = "demo"
	}

	c.JSON(200, secr)
}

func secretsGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{}

	secretId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = secretId
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

	comment := strings.TrimSpace(c.Query("comment"))
	if comment != "" {
		query["comment"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", comment),
			"$options": "i",
		}
	}

	secrs, count, err := secret.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	data := &secretsData{
		Secrets: secrs,
		Count:   count,
	}

	c.JSON(200, data)
}
