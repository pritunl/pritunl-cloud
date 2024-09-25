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
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/plan"
	"github.com/pritunl/pritunl-cloud/utils"
)

type planData struct {
	Id           primitive.ObjectID `json:"id"`
	Name         string             `json:"name"`
	Comment      string             `json:"comment"`
	Organization primitive.ObjectID `json:"organization"`
	Type         string             `json:"type"`
}

type plansData struct {
	Plans []*plan.Plan `json:"plans"`
	Count int64        `json:"count"`
}

func planPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &planData{}

	planId, ok := utils.ParseObjectId(c.Param("plan_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	domn, err := plan.Get(db, planId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	domn.Name = data.Name
	domn.Comment = data.Comment
	domn.Organization = data.Organization
	domn.Type = data.Type

	fields := set.NewSet(
		"name",
		"comment",
		"organization",
		"type",
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

	event.PublishDispatch(db, "plan.change")

	c.JSON(200, domn)
}

func planPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &planData{
		Name: "new.plan",
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	domn := &plan.Plan{
		Name:         data.Name,
		Comment:      data.Comment,
		Organization: data.Organization,
		Type:         data.Type,
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

	event.PublishDispatch(db, "plan.change")

	c.JSON(200, domn)
}

func planDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	planId, ok := utils.ParseObjectId(c.Param("plan_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := plan.Remove(db, planId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "plan.change")

	c.JSON(200, nil)
}

func plansDelete(c *gin.Context) {
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

	err = plan.RemoveMulti(db, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "plan.change")

	c.JSON(200, nil)
}

func planGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	planId, ok := utils.ParseObjectId(c.Param("plan_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	domn, err := plan.Get(db, planId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, domn)
}

func plansGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	if c.Query("names") == "true" {
		query := &bson.M{}

		domns, err := plan.GetAllName(db, query)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		c.JSON(200, domns)
	} else {
		page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
		pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

		query := bson.M{}

		planId, ok := utils.ParseObjectId(c.Query("id"))
		if ok {
			query["_id"] = planId
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

		plans, count, err := plan.GetAllPaged(db, &query, page, pageCount)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		data := &plansData{
			Plans: plans,
			Count: count,
		}

		c.JSON(200, data)
	}
}
