package ahandlers

import (
	"fmt"
	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
)

type domainData struct {
	Id           bson.ObjectId `json:"id"`
	Name         string        `json:"name"`
	Organization bson.ObjectId `json:"organization"`
	Type         string        `json:"type"`
	AwsId        string        `json:"aws_id"`
	AwsSecret    string        `json:"aws_secret"`
}

type domainsData struct {
	Domains []*domain.Domain `json:"domains"`
	Count   int              `json:"count"`
}

func domainPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &domainData{}

	domainId, ok := utils.ParseObjectId(c.Param("domain_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	domn, err := domain.Get(db, domainId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	domn.Name = data.Name
	domn.Organization = data.Organization
	domn.Type = data.Type
	domn.AwsId = data.AwsId
	domn.AwsSecret = data.AwsSecret

	fields := set.NewSet(
		"name",
		"organization",
		"type",
		"aws_id",
		"aws_secret",
	)

	errData, err := domn.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = domn.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "domain.change")

	c.JSON(200, domn)
}

func domainPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &domainData{
		Name: "new.domain",
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	domn := &domain.Domain{
		Name:         data.Name,
		Organization: data.Organization,
		Type:         data.Type,
		AwsId:        data.AwsId,
		AwsSecret:    data.AwsSecret,
	}

	errData, err := domn.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = domn.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "domain.change")

	c.JSON(200, domn)
}

func domainDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	domainId, ok := utils.ParseObjectId(c.Param("domain_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := domain.Remove(db, domainId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "domain.change")

	c.JSON(200, nil)
}

func domainsDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := []bson.ObjectId{}

	err := c.Bind(&data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	err = domain.RemoveMulti(db, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "domain.change")

	c.JSON(200, nil)
}

func domainGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	domainId, ok := utils.ParseObjectId(c.Param("domain_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	domn, err := domain.Get(db, domainId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, domn)
}

func domainsGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	page, _ := strconv.Atoi(c.Query("page"))
	pageCount, _ := strconv.Atoi(c.Query("page_count"))

	query := bson.M{}

	domainId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = domainId
	}

	name := strings.TrimSpace(c.Query("name"))
	if name != "" {
		query["name"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", name),
			"$options": "i",
		}
	}

	organization, ok := utils.ParseObjectId(c.Query("organization"))
	if ok {
		query["organization"] = organization
	}

	domains, count, err := domain.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	data := &domainsData{
		Domains: domains,
		Count:   count,
	}

	c.JSON(200, data)
}
