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
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/journal"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/scheduler"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/utils"
)

type podData struct {
	Id               primitive.ObjectID `json:"id"`
	Name             string             `json:"name"`
	Comment          string             `json:"comment"`
	Organization     primitive.ObjectID `json:"organization"`
	DeleteProtection bool               `json:"delete_protection"`
	Units            []*pod.UnitInput   `json:"units"`
	Count            int                `json:"count"`
}

type podsData struct {
	Pods  []*pod.Pod `json:"pods"`
	Count int64      `json:"count"`
}

type podsDeployData struct {
	Count int                `json:"count"`
	Spec  primitive.ObjectID `json:"spec"`
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
	pd.DeleteProtection = data.DeleteProtection

	fields := set.NewSet(
		"id",
		"name",
		"comment",
		"delete_protection",
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

	errData, err = pd.CommitFieldsUnits(db, data.Units, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
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
		DeleteProtection: data.DeleteProtection,
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

	errData, err = pd.InitUnits(db, data.Units)
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

type PodUnit struct {
	Id          primitive.ObjectID      `json:"id"`
	Pod         primitive.ObjectID      `json:"pod"`
	Commits     []*spec.Commit          `json:"commits"`
	Deployments []*aggregate.Deployment `json:"deployments"`
}

func podUnitGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)

	podId, ok := utils.ParseObjectId(c.Param("pod_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	unitId, ok := utils.ParseObjectId(c.Param("unit_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	pd, err := pod.GetOrg(db, userOrg, podId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	unit := pd.GetUnit(unitId)
	if unit == nil {
		utils.AbortWithStatus(c, 404)
		return
	}

	deploys, err := aggregate.GetDeployments(db, unit.Id)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	commits, err := spec.GetAllProjectSorted(db, &bson.M{
		"unit": unitId,
	})
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	pdUnit := &PodUnit{
		Id:          unit.Id,
		Pod:         pd.Id,
		Commits:     commits,
		Deployments: deploys,
	}

	c.JSON(200, pdUnit)
}

func podUnitDeploymentsPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)
	data := []primitive.ObjectID{}

	podId, ok := utils.ParseObjectId(c.Param("pod_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	unitId, ok := utils.ParseObjectId(c.Param("unit_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(&data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	pd, err := pod.GetOrg(db, userOrg, podId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	unit := pd.GetUnit(unitId)
	if unit == nil {
		utils.AbortWithStatus(c, 404)
		return
	}

	state := c.Query("state")
	switch state {
	case deployment.Archive:
		err = deployment.ArchiveMulti(db, pd.Id, unit.Id, data)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
		break
	case deployment.Restore:
		err = deployment.RestoreMulti(db, pd.Id, unit.Id, data)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
		break
	case deployment.Destroy:
		err = deployment.RemoveMulti(db, pd.Id, unit.Id, data)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
		break
	}

	event.PublishDispatch(db, "instance.change")
	event.PublishDispatch(db, "pod.change")

	c.JSON(200, nil)
}

func podUnitDeploymentPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)
	data := &podsDeployData{}

	podId, ok := utils.ParseObjectId(c.Param("pod_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	unitId, ok := utils.ParseObjectId(c.Param("unit_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(&data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	pd, err := pod.GetOrg(db, userOrg, podId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	unit := pd.GetUnit(unitId)
	if unit == nil {
		utils.AbortWithStatus(c, 404)
		return
	}

	errData, err := scheduler.ManualSchedule(db, unit, data.Spec, data.Count)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	event.PublishDispatch(db, "instance.change")
	event.PublishDispatch(db, "pod.change")

	c.JSON(200, nil)
}

func podUnitDeploymentLogGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)

	podId, ok := utils.ParseObjectId(c.Param("pod_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	unitId, ok := utils.ParseObjectId(c.Param("unit_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	deplyId, ok := utils.ParseObjectId(c.Param("deployment_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	kind := 0
	resource := c.Query("resource")
	switch resource {
	case "agent":
		kind = journal.DeploymentAgent
		break
	default:
		utils.AbortWithStatus(c, 404)
		return
	}

	pd, err := pod.GetOrg(db, userOrg, podId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	unit := pd.GetUnit(unitId)
	if unit == nil {
		utils.AbortWithStatus(c, 404)
		return
	}

	if !unit.HasDeployment(deplyId) {
		utils.AbortWithStatus(c, 404)
		return
	}

	deply, err := deployment.Get(db, deplyId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	data, err := journal.GetOutput(c, db, deply.Id, kind)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, data)
}
