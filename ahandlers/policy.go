package ahandlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/policy"
	"github.com/pritunl/pritunl-cloud/relations"
	"github.com/pritunl/pritunl-cloud/utils"
)

type policyData struct {
	Id                   bson.ObjectID           `json:"id"`
	Name                 string                  `json:"name"`
	Comment              string                  `json:"comment"`
	Disabled             bool                    `json:"disabled"`
	Authorities          []bson.ObjectID         `json:"authorities"`
	Roles                []string                `json:"roles"`
	Rules                map[string]*policy.Rule `json:"rules"`
	AdminSecondary       bson.ObjectID           `json:"admin_secondary"`
	UserSecondary        bson.ObjectID           `json:"user_secondary"`
	ProxySecondary       bson.ObjectID           `json:"proxy_secondary"`
	AuthoritySecondary   bson.ObjectID           `json:"authority_secondary"`
	AdminDeviceSecondary bool                    `json:"admin_device_secondary"`
	UserDeviceSecondary  bool                    `json:"user_device_secondary"`
}

type policiesData struct {
	Policies []*policy.Policy `json:"policies"`
	Count    int64            `json:"count"`
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
	polcy.Comment = data.Comment
	polcy.Disabled = data.Disabled
	polcy.Roles = data.Roles
	polcy.Rules = data.Rules
	polcy.AdminSecondary = data.AdminSecondary
	polcy.UserSecondary = data.UserSecondary
	polcy.AdminDeviceSecondary = data.AdminDeviceSecondary
	polcy.UserDeviceSecondary = data.UserDeviceSecondary

	fields := set.NewSet(
		"name",
		"comment",
		"disabled",
		"roles",
		"rules",
		"admin_secondary",
		"user_secondary",
		"admin_device_secondary",
		"user_device_secondary",
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
		Name:                 data.Name,
		Comment:              data.Comment,
		Disabled:             data.Disabled,
		Roles:                data.Roles,
		Rules:                data.Rules,
		AdminSecondary:       data.AdminSecondary,
		UserSecondary:        data.UserSecondary,
		AdminDeviceSecondary: data.AdminDeviceSecondary,
		UserDeviceSecondary:  data.UserDeviceSecondary,
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

	errData, err := relations.CanDelete(db, "policy", polcyId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = policy.Remove(db, polcyId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "policy.change")

	c.JSON(200, nil)
}

func policiesDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := []bson.ObjectID{}

	err := c.Bind(&data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	errData, err := relations.CanDeleteAll(db, "policy", data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = policy.RemoveMulti(db, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "policy.change")

	c.JSON(200, nil)
}

func policyGet(c *gin.Context) {
	if demo.IsDemo() {
		polcy := demo.Policies[0]
		c.JSON(200, polcy)
		return
	}

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
	if demo.IsDemo() {
		data := &policiesData{
			Policies: demo.Policies,
			Count:    int64(len(demo.Policies)),
		}

		c.JSON(200, data)
		return
	}

	db := c.MustGet("db").(*database.Database)

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{}

	policyId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = policyId
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

	comment := strings.TrimSpace(c.Query("comment"))
	if comment != "" {
		query["comment"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", comment),
			"$options": "i",
		}
	}

	polcies, count, err := policy.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	data := &policiesData{
		Policies: polcies,
		Count:    count,
	}

	c.JSON(200, data)
}
