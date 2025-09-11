package uhandlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/utils"
)

type domainData struct {
	Id         bson.ObjectID    `json:"id"`
	Name       string           `json:"name"`
	Comment    string           `json:"comment"`
	Type       string           `json:"type"`
	Secret     bson.ObjectID    `json:"secret"`
	RootDomain string           `json:"root_domain"`
	Records    []*domain.Record `json:"records"`
}

type domainsData struct {
	Domains []*aggregate.Domain `json:"domains"`
	Count   int64               `json:"count"`
}

func domainPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)
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

	domn, err := domain.GetOrg(db, userOrg, domainId)
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
	domn.Type = data.Type
	domn.Secret = data.Secret
	domn.RootDomain = data.RootDomain
	domn.Records = data.Records

	fields := set.NewSet(
		"name",
		"comment",
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
	userOrg := c.MustGet("organization").(bson.ObjectID)
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
		Organization: userOrg,
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
	userOrg := c.MustGet("organization").(bson.ObjectID)

	domainId, ok := utils.ParseObjectId(c.Param("domain_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := domain.RemoveOrg(db, userOrg, domainId)
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
	userOrg := c.MustGet("organization").(bson.ObjectID)
	data := []bson.ObjectID{}

	err := c.Bind(&data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	err = domain.RemoveMultiOrg(db, userOrg, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "domain.change")

	c.JSON(200, nil)
}

func domainGet(c *gin.Context) {
	if demo.IsDemo() {
		domn := demo.Domains[0]
		c.JSON(200, domn)
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)

	domainId, ok := utils.ParseObjectId(c.Param("domain_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	domn, err := domain.GetOrg(db, userOrg, domainId)
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
	if demo.IsDemo() {
		data := &domainsData{
			Domains: demo.Domains,
			Count:   int64(len(demo.Domains)),
		}

		c.JSON(200, data)
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)

	if c.Query("names") == "true" {
		domns, err := domain.GetAllName(db, &bson.M{
			"organization": userOrg,
		})
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		c.JSON(200, domns)
	} else {
		page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
		pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

		query := bson.M{
			"organization": userOrg,
		}

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

		domains, count, err := aggregate.GetDomainPaged(
			db, &query, page, pageCount)
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
