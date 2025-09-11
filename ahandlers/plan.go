package ahandlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/plan"
	"github.com/pritunl/pritunl-cloud/utils"
)

type planData struct {
	Id           bson.ObjectID     `json:"id"`
	Name         string            `json:"name"`
	Comment      string            `json:"comment"`
	Organization bson.ObjectID     `json:"organization"`
	Statements   []*plan.Statement `json:"statements"`
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

	pln, err := plan.Get(db, planId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	pln.Name = data.Name
	pln.Comment = data.Comment
	pln.Organization = data.Organization

	err = pln.UpdateStatements(data.Statements)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	fields := set.NewSet(
		"name",
		"comment",
		"organization",
		"statements",
	)

	errData, err := pln.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = pln.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "plan.change")

	c.JSON(200, pln)
}

func planPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &planData{
		Name: "New Plan",
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	pln := &plan.Plan{
		Name:         data.Name,
		Comment:      data.Comment,
		Organization: data.Organization,
	}

	err = pln.UpdateStatements(data.Statements)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	errData, err := pln.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = pln.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "plan.change")

	c.JSON(200, pln)
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
	data := []bson.ObjectID{}

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
	if demo.IsDemo() {
		pln := demo.Plans[0]
		c.JSON(200, pln)
		return
	}

	db := c.MustGet("db").(*database.Database)

	planId, ok := utils.ParseObjectId(c.Param("plan_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	pln, err := plan.Get(db, planId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, pln)
}

func plansGet(c *gin.Context) {
	if demo.IsDemo() {
		data := &plansData{
			Plans: demo.Plans,
			Count: int64(len(demo.Plans)),
		}

		c.JSON(200, data)
		return
	}

	db := c.MustGet("db").(*database.Database)

	if c.Query("names") == "true" {
		query := &bson.M{}

		plns, err := plan.GetAllName(db, query)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		c.JSON(200, plns)
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
