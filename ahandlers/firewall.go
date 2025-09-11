package ahandlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/relations"
	"github.com/pritunl/pritunl-cloud/utils"
)

type firewallData struct {
	Id           bson.ObjectID    `json:"id"`
	Name         string           `json:"name"`
	Comment      string           `json:"comment"`
	Organization bson.ObjectID    `json:"organization"`
	Roles        []string         `json:"roles"`
	Ingress      []*firewall.Rule `json:"ingress"`
}

type firewallsData struct {
	Firewalls []*firewall.Firewall `json:"firewalls"`
	Count     int64                `json:"count"`
}

func firewallPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &firewallData{}

	firewallId, ok := utils.ParseObjectId(c.Param("firewall_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	fire, err := firewall.Get(db, firewallId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	fire.Name = data.Name
	fire.Comment = data.Comment
	fire.Organization = data.Organization
	fire.Roles = data.Roles
	fire.Ingress = data.Ingress

	fields := set.NewSet(
		"name",
		"comment",
		"organization",
		"roles",
		"ingress",
	)

	errData, err := fire.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = fire.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "firewall.change")

	c.JSON(200, fire)
}

func firewallPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &firewallData{
		Name: "New Firewall",
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	fire := &firewall.Firewall{
		Name:         data.Name,
		Comment:      data.Comment,
		Organization: data.Organization,
		Roles:        data.Roles,
		Ingress:      data.Ingress,
	}

	errData, err := fire.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = fire.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "firewall.change")

	c.JSON(200, fire)
}

func firewallDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	firewallId, ok := utils.ParseObjectId(c.Param("firewall_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	errData, err := relations.CanDelete(db, "firewall", firewallId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = firewall.Remove(db, firewallId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "firewall.change")

	c.JSON(200, nil)
}

func firewallsDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := []bson.ObjectID{}

	err := c.Bind(&data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	errData, err := relations.CanDeleteAll(db, "firewall", data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = firewall.RemoveMulti(db, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "firewall.change")

	c.JSON(200, nil)
}

func firewallGet(c *gin.Context) {
	if demo.IsDemo() {
		fire := demo.Firewalls[0]
		c.JSON(200, fire)
		return
	}

	db := c.MustGet("db").(*database.Database)

	firewallId, ok := utils.ParseObjectId(c.Param("firewall_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	fire, err := firewall.Get(db, firewallId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, fire)
}

func firewallsGet(c *gin.Context) {
	if demo.IsDemo() {
		data := &firewallsData{
			Firewalls: demo.Firewalls,
			Count:     int64(len(demo.Firewalls)),
		}

		c.JSON(200, data)
		return
	}

	db := c.MustGet("db").(*database.Database)

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{}

	firewallId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = firewallId
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
		query["roles"] = role
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

	firewalls, count, err := firewall.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	data := &firewallsData{
		Firewalls: firewalls,
		Count:     count,
	}

	c.JSON(200, data)
}
