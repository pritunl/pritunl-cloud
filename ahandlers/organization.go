package ahandlers

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/organization"
	"github.com/pritunl/pritunl-cloud/utils"
)

type organizationData struct {
	Id    primitive.ObjectID `json:"id"`
	Name  string             `json:"name"`
	Roles []string           `json:"roles"`
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
	org.Roles = data.Roles

	fields := set.NewSet(
		"name",
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

	org := &organization.Organization{
		Name:  data.Name,
		Roles: data.Roles,
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

	err := organization.Remove(db, orgId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "organization.change")

	c.JSON(200, nil)
}

func organizationGet(c *gin.Context) {
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
	db := c.MustGet("db").(*database.Database)

	orgs, err := organization.GetAll(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, orgs)
}
