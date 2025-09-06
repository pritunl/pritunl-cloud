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
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/organization"
	"github.com/pritunl/pritunl-cloud/relations"
	"github.com/pritunl/pritunl-cloud/subscription"
	"github.com/pritunl/pritunl-cloud/utils"
)

type organizationData struct {
	Id      primitive.ObjectID `json:"id"`
	Name    string             `json:"name"`
	Comment string             `json:"comment"`
	Roles   []string           `json:"roles"`
}

type organizationsData struct {
	Organizations []*organization.Organization `json:"organizations"`
	Count         int64                        `json:"count"`
}

func organizationPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &organizationData{}

	orgId, ok := utils.ParseObjectId(c.Param("org_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	org, err := organization.Get(db, orgId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	org.Name = data.Name
	org.Comment = data.Comment
	org.Roles = data.Roles

	fields := set.NewSet(
		"name",
		"comment",
		"roles",
	)

	errData, err := org.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = org.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "organization.change")

	c.JSON(200, org)
}

func organizationPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &organizationData{
		Name: "New Organization",
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if !subscription.Sub.Active {
		count, e := organization.Count(db)
		if e != nil {
			utils.AbortWithError(c, 500, e)
			return
		}

		if count > 0 {
			errData := &errortypes.ErrorData{
				Error:   "subscription_required",
				Message: "Subscription required for multiple organizations",
			}
			c.JSON(400, errData)
			return
		}
	}

	org := &organization.Organization{
		Name:    data.Name,
		Comment: data.Comment,
		Roles:   data.Roles,
	}

	errData, err := org.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = org.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "organization.change")

	c.JSON(200, org)
}

func organizationDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	orgId, ok := utils.ParseObjectId(c.Param("org_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	errData, err := relations.CanDelete(db, "organization", orgId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = organization.Remove(db, orgId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "organization.change")

	c.JSON(200, nil)
}

func organizationGet(c *gin.Context) {
	if demo.IsDemo() {
		org := demo.Organizations[0]
		c.JSON(200, org)
		return
	}

	db := c.MustGet("db").(*database.Database)

	orgId, ok := utils.ParseObjectId(c.Param("org_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	org, err := organization.Get(db, orgId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, org)
}

func organizationsGet(c *gin.Context) {
	if demo.IsDemo() {
		data := &organizationsData{
			Organizations: demo.Organizations,
			Count:         int64(len(demo.Organizations)),
		}

		c.JSON(200, data)
		return
	}

	db := c.MustGet("db").(*database.Database)

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{}

	organizationId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = organizationId
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

	secrs, count, err := organization.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	data := &organizationsData{
		Organizations: secrs,
		Count:         count,
	}

	c.JSON(200, data)
}
