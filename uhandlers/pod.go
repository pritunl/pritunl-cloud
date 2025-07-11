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
	"github.com/pritunl/pritunl-cloud/authorizer"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/journal"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/relations"
	"github.com/pritunl/pritunl-cloud/scheduler"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/unit"
	"github.com/pritunl/pritunl-cloud/utils"
)

type podData struct {
	Id               primitive.ObjectID `json:"id"`
	Name             string             `json:"name"`
	Comment          string             `json:"comment"`
	Organization     primitive.ObjectID `json:"organization"`
	DeleteProtection bool               `json:"delete_protection"`
	Units            []*unit.UnitInput  `json:"units"`
	Drafts           []*pod.UnitDraft   `json:"drafts"`
	Count            int                `json:"count"`
}

type podsData struct {
	Pods  []*aggregate.PodAggregate `json:"pods"`
	Count int64                     `json:"count"`
}

type podsDeployData struct {
	Count int                `json:"count"`
	Spec  primitive.ObjectID `json:"spec"`
}

type deploymentData struct {
	Id   primitive.ObjectID `json:"id"`
	Tags []string           `json:"tags"`
}

type specsData struct {
	Specs []*spec.Spec `json:"specs"`
	Count int64        `json:"count"`
}

func podPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)
	authr := c.MustGet("authorizer").(*authorizer.Authorizer)
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

	usr, err := authr.GetUser(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	err = pod.UpdateDrafts(db, podId, usr.Id, []*pod.UnitDraft{})
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "pod.change")
	event.PublishDispatch(db, "unit.change")

	c.JSON(200, pd)
}

func podDraftsPut(c *gin.Context) {
	if demo.BlockedSilent(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)
	authr := c.MustGet("authorizer").(*authorizer.Authorizer)
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

	usr, err := authr.GetUser(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	err = pod.UpdateDraftsOrg(db, userOrg, podId, usr.Id, data.Drafts)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, nil)
}

func podDeployPut(c *gin.Context) {
	if demo.BlockedSilent(c) {
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

	units, err := unit.GetAll(db, &bson.M{
		"pod":          podId,
		"organization": userOrg,
	})
	if err != nil {
		return
	}

	unitsDataMap := map[primitive.ObjectID]*unit.UnitInput{}
	for _, unitData := range data.Units {
		unitsDataMap[unitData.Id] = unitData
	}

	for _, unt := range units {
		unitData := unitsDataMap[unt.Id]
		if unitData == nil || unitData.DeploySpec.IsZero() {
			continue
		}

		deploySpec, e := spec.Get(db, unitData.DeploySpec)
		if e != nil || deploySpec.Unit != unt.Id {
			errData := &errortypes.ErrorData{
				Error:   "unit_deploy_spec_invalid",
				Message: "Invalid unit deployment commit",
			}
			c.JSON(400, errData)
			return
		}

		unt.DeploySpec = unitData.DeploySpec
		err = unt.CommitFields(db, set.NewSet("deploy_spec"))
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
	}

	c.JSON(200, nil)
}

func podPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)
	data := &podData{
		Name: "new-pod",
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

	err = pd.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
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

	event.PublishDispatch(db, "pod.change")
	event.PublishDispatch(db, "unit.change")

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

	errData, err := relations.CanDeleteOrg(db, "pod", userOrg, podId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = pod.RemoveOrg(db, userOrg, podId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "pod.change")
	event.PublishDispatch(db, "unit.change")

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

	for _, podId := range data {
		errData, err := relations.CanDeleteOrg(db, "pod", userOrg, podId)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		if errData != nil {
			c.JSON(400, errData)
			return
		}
	}

	err = pod.RemoveMultiOrg(db, userOrg, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "pod.change")
	event.PublishDispatch(db, "unit.change")

	c.JSON(200, nil)
}

func podGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	authr := c.MustGet("authorizer").(*authorizer.Authorizer)
	userOrg := c.MustGet("organization").(primitive.ObjectID)

	podId, ok := utils.ParseObjectId(c.Param("pod_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	usr, err := authr.GetUser(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	pd, err := aggregate.GetPod(db, usr.Id, &bson.M{
		"_id":          podId,
		"organization": userOrg,
	})
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			c.AbortWithStatus(404)
		} else {
			utils.AbortWithError(c, 500, err)
		}
		return
	}

	c.JSON(200, pd)
}

func podsGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	authr := c.MustGet("authorizer").(*authorizer.Authorizer)
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

	usr, err := authr.GetUser(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
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

	pods, count, err := aggregate.GetPodsPaged(db, usr.Id,
		&query, page, pageCount)
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
	Kind        string                  `json:"kind"`
	Deployments []*aggregate.Deployment `json:"deployments"`
}

func podUnitGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)

	unitId, ok := utils.ParseObjectId(c.Param("unit_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	unt, err := unit.GetOrg(db, userOrg, unitId)
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			c.AbortWithStatus(404)
		} else {
			utils.AbortWithError(c, 500, err)
		}
		return
	}

	deploys, err := aggregate.GetDeployments(db, unt)
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			c.AbortWithStatus(404)
		} else {
			utils.AbortWithError(c, 500, err)
		}
		return
	}

	pdUnit := &PodUnit{
		Id:          unt.Id,
		Pod:         unt.Pod,
		Kind:        unt.Kind,
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

	unitId, ok := utils.ParseObjectId(c.Param("unit_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(&data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	unt, err := unit.GetOrg(db, userOrg, unitId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	action := c.Query("action")
	switch action {
	case deployment.Archive:
		err = deployment.ArchiveMulti(db, unt.Id, data)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
		break
	case deployment.Restore:
		err = deployment.RestoreMulti(db, unt.Id, data)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
		break
	case deployment.Destroy:
		err = deployment.RemoveMulti(db, unt.Id, data)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
		break
	case deployment.Migrate:
		commitId, ok := utils.ParseObjectId(c.Query("commit"))
		if !ok {
			utils.AbortWithStatus(c, 400)
			return
		}

		errData, err := unt.MigrateDeployements(db, commitId, data)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		if errData != nil {
			c.JSON(400, errData)
			return
		}

		break
	}

	event.PublishDispatch(db, "instance.change")
	event.PublishDispatch(db, "pod.change")
	event.PublishDispatch(db, "unit.change")

	c.JSON(200, nil)
}

func podUnitDeploymentPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)
	data := &podsDeployData{}

	unitId, ok := utils.ParseObjectId(c.Param("unit_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(&data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	unt, err := unit.GetOrg(db, userOrg, unitId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	errData, err := scheduler.ManualSchedule(db, unt, data.Spec, data.Count)
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
	event.PublishDispatch(db, "unit.change")

	c.JSON(200, nil)
}

func podUnitDeploymentPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)
	data := &deploymentData{}

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

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	deply, err := deployment.GetUnitOrg(db, userOrg, unitId, deplyId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	deply.Tags = data.Tags

	fields := set.NewSet(
		"tags",
	)

	errData, err := deply.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = deply.CommitFields(db, fields)
	if err != nil {
		return
	}

	event.PublishDispatch(db, "instance.change")
	event.PublishDispatch(db, "pod.change")

	c.JSON(200, nil)
}

func podUnitDeploymentLogGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)

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

	deply, err := deployment.GetUnitOrg(db, userOrg, unitId, deplyId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	data, err := journal.GetOutput(c, db, deply.Id, kind)
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			c.AbortWithStatus(404)
		} else {
			utils.AbortWithError(c, 500, err)
		}
		return
	}

	c.JSON(200, data)
}

func podUnitSpecsGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	unitId, ok := utils.ParseObjectId(c.Param("unit_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	specs, count, err := spec.GetAllPaged(db, &bson.M{
		"unit":         unitId,
		"organization": userOrg,
	}, page, pageCount)
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			c.AbortWithStatus(404)
		} else {
			utils.AbortWithError(c, 500, err)
		}
		return
	}

	data := &specsData{
		Specs: specs,
		Count: count,
	}

	c.JSON(200, data)
}

func podUnitSpecGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)

	unitId, ok := utils.ParseObjectId(c.Param("unit_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	specId, ok := utils.ParseObjectId(c.Param("spec_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	spec, err := spec.GetOne(db, &bson.M{
		"_id":          specId,
		"unit":         unitId,
		"organization": userOrg,
	})
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			c.AbortWithStatus(404)
		} else {
			utils.AbortWithError(c, 500, err)
		}
		return
	}

	c.JSON(200, spec)
}
