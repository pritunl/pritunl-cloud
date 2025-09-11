package uhandlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/alert"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/utils"
)

type alertData struct {
	Id        bson.ObjectID `json:"id"`
	Name      string        `json:"name"`
	Comment   string        `json:"comment"`
	Roles     []string      `json:"roles"`
	Resource  string        `json:"resource"`
	Level     int           `json:"level"`
	Frequency int           `bson:"frequency" json:"frequency"`
	Ignores   []string      `bson:"ignores" json:"ignores"`
	ValueInt  int           `json:"value_int"`
	ValueStr  string        `json:"value_str"`
}

type alertsData struct {
	Alerts []*alert.Alert `json:"alerts"`
	Count  int64          `json:"count"`
}

func alertPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)
	data := &alertData{}

	alertId, ok := utils.ParseObjectId(c.Param("alert_id"))
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

	alrt, err := alert.GetOrg(db, userOrg, alertId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	alrt.Name = data.Name
	alrt.Comment = data.Comment
	alrt.Roles = data.Roles
	alrt.Resource = data.Resource
	alrt.Level = data.Level
	alrt.Frequency = data.Frequency
	alrt.Ignores = data.Ignores
	alrt.ValueInt = data.ValueInt
	alrt.ValueStr = data.ValueStr

	fields := set.NewSet(
		"name",
		"comment",
		"roles",
		"resource",
		"level",
		"frequency",
		"ignores",
		"value_int",
		"value_str",
	)

	errData, err := alrt.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = alrt.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	_ = event.PublishDispatch(db, "alert.change")

	c.JSON(200, alrt)
}

func alertPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)
	data := &alertData{
		Name:     "New Alert",
		Resource: alert.InstanceOffline,
		Level:    alert.Medium,
	}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	alrt := &alert.Alert{
		Name:         data.Name,
		Comment:      data.Comment,
		Organization: userOrg,
		Roles:        data.Roles,
		Resource:     data.Resource,
		Level:        data.Level,
		Frequency:    data.Frequency,
		Ignores:      data.Ignores,
		ValueInt:     data.ValueInt,
		ValueStr:     data.ValueStr,
	}

	errData, err := alrt.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = alrt.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	_ = event.PublishDispatch(db, "alert.change")

	c.JSON(200, alrt)
}

func alertDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)

	alertId, ok := utils.ParseObjectId(c.Param("alert_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := alert.RemoveOrg(db, userOrg, alertId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	_ = event.PublishDispatch(db, "alert.change")

	c.JSON(200, nil)
}

func alertsDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)
	dta := []bson.ObjectID{}

	err := c.Bind(&dta)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	err = alert.RemoveMultiOrg(db, userOrg, dta)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	_ = event.PublishDispatch(db, "alert.change")

	c.JSON(200, nil)
}

func alertGet(c *gin.Context) {
	if demo.IsDemo() {
		alrt := demo.Alerts[0]
		c.JSON(200, alrt)
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)

	alertId, ok := utils.ParseObjectId(c.Query("id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	alrt, err := alert.GetOrg(db, userOrg, alertId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, alrt)
}

func alertsGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{
		"organization": userOrg,
	}

	alertId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = alertId
	}

	name := strings.TrimSpace(c.Query("name"))
	if name != "" {
		query["$or"] = []*bson.M{
			&bson.M{
				"name": &bson.M{
					"$regex":   fmt.Sprintf(".*%s.*", regexp.QuoteMeta(name)),
					"$options": "i",
				},
			},
		}
	}

	role := strings.TrimSpace(c.Query("role"))
	if role != "" {
		query["roles"] = role
	}

	comment := strings.TrimSpace(c.Query("comment"))
	if comment != "" {
		query["comment"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", comment),
			"$options": "i",
		}
	}

	alerts, count, err := alert.GetAllPaged(
		db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	dta := &alertsData{
		Alerts: alerts,
		Count:  count,
	}

	c.JSON(200, dta)
}
