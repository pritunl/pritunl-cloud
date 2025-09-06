package ahandlers

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
	"github.com/pritunl/pritunl-cloud/balancer"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/utils"
)

type balancerData struct {
	Id           primitive.ObjectID   `json:"id"`
	Name         string               `json:"name"`
	Comment      string               `json:"comment"`
	State        bool                 `json:"state"`
	Type         string               `json:"type"`
	Organization primitive.ObjectID   `json:"organization"`
	Datacenter   primitive.ObjectID   `json:"datacenter"`
	Certificates []primitive.ObjectID `json:"certificates"`
	WebSockets   bool                 `json:"websockets"`
	Domains      []*balancer.Domain   `json:"domains"`
	Backends     []*balancer.Backend  `json:"backends"`
	CheckPath    string               `json:"check_path"`
}

type balancersData struct {
	Balancers []*balancer.Balancer `json:"balancers"`
	Count     int64                `json:"count"`
}

func balancerPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &balancerData{}

	balancerId, ok := utils.ParseObjectId(c.Param("balancer_id"))
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

	balnc, err := balancer.Get(db, balancerId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	balnc.Name = data.Name
	balnc.Comment = data.Comment
	balnc.State = data.State
	balnc.Type = data.Type
	balnc.Organization = data.Organization
	balnc.Datacenter = data.Datacenter
	balnc.Certificates = data.Certificates
	balnc.WebSockets = data.WebSockets
	balnc.Domains = data.Domains
	balnc.Backends = data.Backends
	balnc.CheckPath = data.CheckPath

	fields := set.NewSet(
		"name",
		"comment",
		"state",
		"type",
		"organization",
		"datacenter",
		"certificates",
		"websockets",
		"domains",
		"backends",
		"check_path",
	)

	errData, err := balnc.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = balnc.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "balancer.change")

	balnc.Json()

	c.JSON(200, balnc)
}

func balancerPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &balancerData{
		Name: "New Balancer",
	}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	balnc := &balancer.Balancer{
		Name:         data.Name,
		Comment:      data.Comment,
		State:        data.State,
		Type:         data.Type,
		Organization: data.Organization,
		Datacenter:   data.Datacenter,
		Certificates: data.Certificates,
		WebSockets:   data.WebSockets,
		Domains:      data.Domains,
		Backends:     data.Backends,
		CheckPath:    data.CheckPath,
	}

	errData, err := balnc.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = balnc.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "balancer.change")

	balnc.Json()

	c.JSON(200, balnc)
}

func balancerDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	balancerId, ok := utils.ParseObjectId(c.Param("balancer_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := balancer.Remove(db, balancerId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "balancer.change")

	c.JSON(200, nil)
}

func balancersDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := []primitive.ObjectID{}

	err := c.Bind(&data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	err = balancer.RemoveMulti(db, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "balancer.change")

	c.JSON(200, nil)
}

func balancerGet(c *gin.Context) {
	if demo.IsDemo() {
		balnc := demo.Balancers[0]
		c.JSON(200, balnc)
		return
	}

	db := c.MustGet("db").(*database.Database)

	balancerId, ok := utils.ParseObjectId(c.Param("balancer_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	balnc, err := balancer.Get(db, balancerId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	balnc.Json()

	c.JSON(200, balnc)
}

func balancersGet(c *gin.Context) {
	if demo.IsDemo() {
		data := &balancersData{
			Balancers: demo.Balancers,
			Count:     int64(len(demo.Balancers)),
		}

		c.JSON(200, data)
		return
	}

	db := c.MustGet("db").(*database.Database)

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{}

	balancerId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = balancerId
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

	datacenter, ok := utils.ParseObjectId(c.Query("datacenter"))
	if ok {
		query["datacenter"] = datacenter
	}

	comment := strings.TrimSpace(c.Query("comment"))
	if comment != "" {
		query["comment"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", comment),
			"$options": "i",
		}
	}

	balncs, count, err := balancer.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	for _, balnc := range balncs {
		balnc.Json()
	}

	data := &balancersData{
		Balancers: balncs,
		Count:     count,
	}

	c.JSON(200, data)
}
