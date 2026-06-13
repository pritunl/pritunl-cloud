package uhandlers

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
	userOrg := c.MustGet("organization").(bson.ObjectID)

	advisoryId, ok := utils.ParseObjectId(c.Param("advisory_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	adv, err := advisory.GetOrg(db, userOrg, advisoryId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, adv)
}

func advisoriesGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{
		"organization": userOrg,
	}

	advisoryId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = advisoryId
	}

	reference := strings.TrimSpace(c.Query("reference"))
	if reference != "" {
		query["reference"] = &bson.M{
			"$regex":   fmt.Sprintf("^([^:]+:)?%s", regexp.QuoteMeta(reference)),
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

type advisoryData struct {
	Restore    bool            `json:"restore"`
	Dismiss    bool            `json:"dismiss"`
	Restores   []bson.ObjectID `json:"restores"`
	Dismissals []bson.ObjectID `json:"dismissals"`
}

type advisoryMultiData struct {
	Ids     []bson.ObjectID `json:"ids"`
	Restore bool            `json:"restore"`
	Dismiss bool            `json:"dismiss"`
}

func advisoryPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)
	dta := &advisoryData{}

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

	err = advisory.UpdateDismissOrg(db, userOrg, advisoryId,
		dta.Dismiss, dta.Restore, dta.Dismissals, dta.Restores)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "advisory.change")

	c.JSON(200, nil)
}

func advisoriesPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)
	dta := &advisoryMultiData{}

	err := c.Bind(dta)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	err = advisory.UpdateMultiOrg(db, userOrg, dta.Ids, dta.Dismiss, dta.Restore)
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
	userOrg := c.MustGet("organization").(bson.ObjectID)

	advisoryId, ok := utils.ParseObjectId(c.Param("advisory_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := advisory.RemoveOrg(db, userOrg, advisoryId)
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
	userOrg := c.MustGet("organization").(bson.ObjectID)
	dta := []bson.ObjectID{}

	err := c.Bind(&dta)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	err = advisory.RemoveMultiOrg(db, userOrg, dta)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "advisory.change")

	c.JSON(200, nil)
}
