package ahandlers

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/policy"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
)

type policyData struct {
	Id                 bson.ObjectId           `json:"id"`
	Name               string                  `json:"name"`
	Services           []bson.ObjectId         `json:"services"`
	Authorities        []bson.ObjectId         `json:"authorities"`
	Roles              []string                `json:"roles"`
	Rules              map[string]*policy.Rule `json:"rules"`
	KeybaseMode        string                  `json:"keybase_mode"`
	AdminSecondary     bson.ObjectId           `json:"admin_secondary"`
	UserSecondary      bson.ObjectId           `json:"user_secondary"`
	ProxySecondary     bson.ObjectId           `json:"proxy_secondary"`
	AuthoritySecondary bson.ObjectId           `json:"authority_secondary"`
}

func policyPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &policyData{}

	polcyId, ok := utils.ParseObjectId(c.Param("policy_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	polcy, err := policy.Get(db, polcyId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	polcy.Name = data.Name
	polcy.Roles = data.Roles
	polcy.Rules = data.Rules
	polcy.AdminSecondary = data.AdminSecondary
	polcy.UserSecondary = data.UserSecondary

	fields := set.NewSet(
		"name",
		"roles",
		"rules",
		"admin_secondary",
		"user_secondary",
	)

	errData, err := polcy.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = polcy.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "policy.change")

	c.JSON(200, polcy)
}

func policyPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &policyData{
		Name: "New Policy",
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	polcy := &policy.Policy{
		Name:               data.Name,
		Roles:              data.Roles,
		Rules:              data.Rules,
		AdminSecondary:     data.AdminSecondary,
		UserSecondary:      data.UserSecondary,
	}

	errData, err := polcy.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = polcy.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "policy.change")

	c.JSON(200, polcy)
}

func policyDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	polcyId, ok := utils.ParseObjectId(c.Param("policy_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := policy.Remove(db, polcyId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "policy.change")

	c.JSON(200, nil)
}

func policyGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	polcyId, ok := utils.ParseObjectId(c.Param("policy_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	polcy, err := policy.Get(db, polcyId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, polcy)
}

func policiesGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	policies, err := policy.GetAll(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, policies)
}
