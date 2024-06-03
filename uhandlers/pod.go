package uhandlers

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
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/utils"
)

type podData struct {
	Id               primitive.ObjectID `json:"id"`
	Name             string             `json:"name"`
	Comment          string             `json:"comment"`
	Organization     primitive.ObjectID `json:"organization"`
	Type             string             `json:"type"`
	DeleteProtection bool               `json:"delete_protection"`
	Zone             primitive.ObjectID `json:"zone"`
	Roles            []string           `json:"roles"`
	Spec             string             `json:"spec"`
}

type podsData struct {
	Pods  []*pod.Pod `json:"pods"`
	Count int64      `json:"count"`
}

func podPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)
	data := &podData{}

	podId, ok := utils.ParseObjectId(c.Param("pod_id"))
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

	pd, err := pod.GetOrg(db, userOrg, podId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	pd.Name = data.Name
	pd.Comment = data.Comment
	pd.Type = data.Type
	pd.DeleteProtection = data.DeleteProtection
	pd.Zone = data.Zone
	pd.Roles = data.Roles
	pd.Spec = data.Spec

	fields := set.NewSet(
		"id",
		"name",
		"comment",
		"type",
		"delete_protection",
		"zone",
		"roles",
		"spec",
	)

	errData, err := pd.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = pd.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "pod.change")

	c.JSON(200, pd)
}

func podPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)
	data := &podData{
		Name: "New Pod",
	}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	pd := &pod.Pod{
		Name:             data.Name,
		Comment:          data.Comment,
		Organization:     userOrg,
		Type:             data.Type,
		DeleteProtection: data.DeleteProtection,
		Zone:             data.Zone,
		Roles:            data.Roles,
		Spec:             data.Spec,
	}

	errData, err := pd.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = pd.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "pod.change")

	c.JSON(200, pd)
}

func podDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)

	podId, ok := utils.ParseObjectId(c.Param("pod_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := pod.RemoveOrg(db, userOrg, podId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "pod.change")

	c.JSON(200, nil)
}

func podsDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)
	data := []primitive.ObjectID{}

	err := c.Bind(&data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	err = pod.RemoveMultiOrg(db, userOrg, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "pod.change")

	c.JSON(200, nil)
}

func podGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)

	podId, ok := utils.ParseObjectId(c.Param("pod_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	pd, err := pod.GetOrg(db, userOrg, podId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, pd)
}

func podsGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{
		"organization": userOrg,
	}

	podId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = podId
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

	role := strings.TrimSpace(c.Query("role"))
	if role != "" {
		query["role"] = role
	}

	pods, count, err := pod.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	data := &podsData{
		Pods:  pods,
		Count: count,
	}

	c.JSON(200, data)
}
