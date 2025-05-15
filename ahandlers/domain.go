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
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/utils"
)

type domainData struct {
	Id           primitive.ObjectID `json:"id"`
	Name         string             `json:"name"`
	Comment      string             `json:"comment"`
	Organization primitive.ObjectID `json:"organization"`
	Type         string             `json:"type"`
	Secret       primitive.ObjectID `json:"secret"`
	RootDomain   string             `json:"root_domain"`
	Records      []*domain.Record   `json:"records"`
}

type domainsData struct {
	Domains []*domain.Domain `json:"domains"`
	Count   int64            `json:"count"`
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

	err = domn.LoadRecords(db, true)
	if err != nil {
		return
	}

	domn.PreCommit()

	domn.Name = data.Name
	domn.Comment = data.Comment
	domn.Organization = data.Organization
	domn.Type = data.Type
	domn.Secret = data.Secret
	domn.RootDomain = data.RootDomain
	domn.Records = data.Records

	fields := set.NewSet(
		"name",
		"comment",
		"organization",
		"type",
		"secret",
		"root_domain",
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

	err = domn.CommitRecords(db)
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
		Comment:      data.Comment,
		Organization: data.Organization,
		Type:         data.Type,
		Secret:       data.Secret,
		RootDomain:   data.RootDomain,
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
	data := []primitive.ObjectID{}

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

	err = domn.LoadRecords(db, true)
	if err != nil {
		return
	}

	domn.Json()

	c.JSON(200, domn)
}

func domainsGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	if c.Query("names") == "true" {
		query := &bson.M{}

		domns, err := domain.GetAllName(db, query)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		c.JSON(200, domns)
	} else {
		page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
		pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

		query := bson.M{}

		domainId, ok := utils.ParseObjectId(c.Query("id"))
		if ok {
			query["_id"] = domainId
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
}
