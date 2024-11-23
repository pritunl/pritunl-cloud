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
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/scheduler"
	"github.com/pritunl/pritunl-cloud/service"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/utils"
)

type serviceData struct {
	Id               primitive.ObjectID   `json:"id"`
	Name             string               `json:"name"`
	Comment          string               `json:"comment"`
	Organization     primitive.ObjectID   `json:"organization"`
	DeleteProtection bool                 `json:"delete_protection"`
	Units            []*service.UnitInput `json:"units"`
	Count            int                  `json:"count"`
}

type servicesData struct {
	Services []*service.Service `json:"services"`
	Count    int64              `json:"count"`
}

func servicePut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &serviceData{}

	serviceId, ok := utils.ParseObjectId(c.Param("service_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	pd, err := service.Get(db, serviceId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	pd.Name = data.Name
	pd.Comment = data.Comment
	pd.Organization = data.Organization
	pd.DeleteProtection = data.DeleteProtection

	fields := set.NewSet(
		"id",
		"name",
		"comment",
		"organization",
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

	event.PublishDispatch(db, "service.change")

	c.JSON(200, pd)
}

func servicePost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &serviceData{
		Name: "New Service",
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	pd := &service.Service{
		Name:             data.Name,
		Comment:          data.Comment,
		Organization:     data.Organization,
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

	event.PublishDispatch(db, "service.change")

	c.JSON(200, pd)
}

func serviceDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	serviceId, ok := utils.ParseObjectId(c.Param("service_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := service.Remove(db, serviceId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "service.change")

	c.JSON(200, nil)
}

func servicesDelete(c *gin.Context) {
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

	err = service.RemoveMulti(db, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "service.change")

	c.JSON(200, nil)
}

func serviceGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	serviceId, ok := utils.ParseObjectId(c.Param("service_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	pd, err := service.Get(db, serviceId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, pd)
}

func servicesGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{}

	serviceId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = serviceId
	}

	name := strings.TrimSpace(c.Query("name"))
	if name != "" {
		query["name"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", regexp.QuoteMeta(name)),
			"$options": "i",
		}
	}

	role := strings.TrimSpace(c.Query("role"))
	if role != "" {
		query["role"] = role
	}

	organization, ok := utils.ParseObjectId(c.Query("organization"))
	if ok {
		query["organization"] = organization
	}

	comment := strings.TrimSpace(c.Query("comment"))
	if comment != "" {
		query["comment"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", comment),
			"$options": "i",
		}
	}

	services, count, err := service.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	data := &servicesData{
		Services: services,
		Count:    count,
	}

	c.JSON(200, data)
}

type ServiceUnit struct {
	Id          primitive.ObjectID      `json:"id"`
	Service     primitive.ObjectID      `json:"service"`
	Kind        string                  `bson:"kind" json:"kind"`
	Commits     []*spec.Commit          `json:"commits"`
	Deployments []*aggregate.Deployment `json:"deployments"`
}

func serviceUnitGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	serviceId, ok := utils.ParseObjectId(c.Param("service_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	unitId, ok := utils.ParseObjectId(c.Param("unit_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	srvc, err := service.Get(db, serviceId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	unit := srvc.GetUnit(unitId)
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

	srvcUnit := &ServiceUnit{
		Id:          unit.Id,
		Service:     srvc.Id,
		Kind:        unit.Kind,
		Commits:     commits,
		Deployments: deploys,
	}

	c.JSON(200, srvcUnit)
}

func serviceUnitDeploymentPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := []primitive.ObjectID{}

	serviceId, ok := utils.ParseObjectId(c.Param("service_id"))
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

	srvc, err := service.Get(db, serviceId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	unit := srvc.GetUnit(unitId)
	if unit == nil {
		utils.AbortWithStatus(c, 404)
		return
	}

	state := c.Query("state")
	switch state {
	case deployment.Archive:
		err = deployment.ArchiveMulti(db, srvc.Id, unit.Id, data)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
		break
	case deployment.Restore:
		err = deployment.RestoreMulti(db, srvc.Id, unit.Id, data)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
		break
	case deployment.Destroy:
		err = deployment.RemoveMulti(db, srvc.Id, unit.Id, data)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
		break
	}

	event.PublishDispatch(db, "instance.change")
	event.PublishDispatch(db, "service.change")

	c.JSON(200, nil)
}

func serviceUnitDeploymentPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &serviceData{}

	serviceId, ok := utils.ParseObjectId(c.Param("service_id"))
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

	srvc, err := service.Get(db, serviceId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	unit := srvc.GetUnit(unitId)
	if unit == nil {
		utils.AbortWithStatus(c, 404)
		return
	}

	errData, err := scheduler.ManualSchedule(db, unit, data.Count)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	event.PublishDispatch(db, "instance.change")
	event.PublishDispatch(db, "service.change")

	c.JSON(200, nil)
}
