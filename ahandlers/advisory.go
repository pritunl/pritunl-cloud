package ahandlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/advisory"
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/utils"
)

type advisoriesData struct {
	Advisories []*aggregate.AdvisoryAggregate `json:"advisories"`
	Count      int64                          `json:"count"`
}

func advisoryGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	advisoryId, ok := utils.ParseObjectId(c.Param("advisory_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	adv, err := advisory.Get(db, advisoryId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, adv)
}

func advisoriesGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{}

	advisoryId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = advisoryId
	}

	reference := strings.TrimSpace(c.Query("reference"))
	if reference != "" {
		query["reference"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", regexp.QuoteMeta(reference)),
			"$options": "i",
		}
	}

	advisoryType := strings.TrimSpace(c.Query("type"))
	if advisoryType != "" {
		query["type"] = advisoryType
	}

	severity := strings.TrimSpace(c.Query("severity"))
	if severity != "" {
		query["severity"] = severity
	}

	organization, ok := utils.ParseObjectId(c.Query("organization"))
	if ok {
		query["organization"] = organization
	}

	if c.Query("dismissed") != "true" {
		query["dismissed"] = false
	}

	advisories, count, err := aggregate.GetAdvisoryPaged(
		db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	dta := &advisoriesData{
		Advisories: advisories,
		Count:      count,
	}

	c.JSON(200, dta)
}

type advisoryDismissData struct {
	Dismissed bool `json:"dismissed"`
}

type advisoryDismissalsData struct {
	Dismissals []bson.ObjectID `json:"dismissals"`
}

func advisoryDismissPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	dta := &advisoryDismissData{}

	advisoryId, ok := utils.ParseObjectId(c.Param("advisory_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(dta)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	err = advisory.SetDismissed(db, advisoryId, dta.Dismissed)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "advisory.change")

	c.JSON(200, nil)
}

func advisoryDismissalsPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	dta := &advisoryDismissalsData{}

	advisoryId, ok := utils.ParseObjectId(c.Param("advisory_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(dta)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	err = advisory.AddDismissals(db, advisoryId, dta.Dismissals)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "advisory.change")

	c.JSON(200, nil)
}

func advisoryDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	advisoryId, ok := utils.ParseObjectId(c.Param("advisory_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := advisory.Remove(db, advisoryId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "advisory.change")

	c.JSON(200, nil)
}

func advisoriesDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	dta := []bson.ObjectID{}

	err := c.Bind(&dta)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	err = advisory.RemoveMulti(db, dta)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "advisory.change")

	c.JSON(200, nil)
}
