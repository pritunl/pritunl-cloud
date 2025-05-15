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
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vpc"
)

type vpcData struct {
	Id            primitive.ObjectID `json:"id"`
	Name          string             `json:"name"`
	Comment       string             `json:"comment"`
	Network       string             `json:"network"`
	IcmpRedirects bool               `json:"icmp_redirects"`
	Subnets       []*vpc.Subnet      `json:"subnets"`
	Organization  primitive.ObjectID `json:"organization"`
	Datacenter    primitive.ObjectID `json:"datacenter"`
	Routes        []*vpc.Route       `json:"routes"`
	Maps          []*vpc.Map         `json:"maps"`
	Arps          []*vpc.Arp         `json:"arps"`
}

type vpcsData struct {
	Vpcs  []*vpc.Vpc `json:"vpcs"`
	Count int64      `json:"count"`
}

func vpcPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &vpcData{}

	vpcId, ok := utils.ParseObjectId(c.Param("vpc_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	vc, err := vpc.Get(db, vpcId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	vc.PreCommit()

	vc.Name = data.Name
	vc.Comment = data.Comment
	vc.IcmpRedirects = data.IcmpRedirects
	vc.Routes = data.Routes
	vc.Maps = data.Maps
	vc.Arps = data.Arps
	vc.Subnets = data.Subnets

	fields := set.NewSet(
		"name",
		"comment",
		"icmp_redirects",
		"routes",
		"maps",
		"arps",
		"subnets",
	)

	errData, err := vc.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	errData, err = vc.PostCommit(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = vc.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "vpc.change")

	vc.Json()

	c.JSON(200, vc)
}

func vpcPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &vpcData{
		Name: "New VPC",
	}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "ahandler: Failed to bind"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	vc := &vpc.Vpc{
		Name:          data.Name,
		Comment:       data.Comment,
		Network:       data.Network,
		Subnets:       data.Subnets,
		Organization:  data.Organization,
		Datacenter:    data.Datacenter,
		IcmpRedirects: data.IcmpRedirects,
		Routes:        data.Routes,
		Maps:          data.Maps,
		Arps:          data.Arps,
	}

	vc.InitVpc()

	errData, err := vc.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = vc.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "vpc.change")

	vc.Json()

	c.JSON(200, vc)
}

func vpcDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	vpcId, ok := utils.ParseObjectId(c.Param("vpc_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := vpc.Remove(db, vpcId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "vpc.change")

	c.JSON(200, nil)
}

func vpcsDelete(c *gin.Context) {
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

	err = vpc.RemoveMulti(db, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "vpc.change")

	c.JSON(200, nil)
}

func vpcGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	vpcId, ok := utils.ParseObjectId(c.Param("vpc_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	vc, err := vpc.Get(db, vpcId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	vc.Json()

	c.JSON(200, vc)
}

func vpcRoutesGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	vpcId, ok := utils.ParseObjectId(c.Param("vpc_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	vc, err := vpc.Get(db, vpcId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, vc.Routes)
}

func vpcRoutesPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := []*vpc.Route{}

	vpcId, ok := utils.ParseObjectId(c.Param("vpc_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(&data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	vc, err := vpc.Get(db, vpcId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	vc.Routes = data

	fields := set.NewSet(
		"routes",
	)

	errData, err := vc.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = vc.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "vpc.change")

	vc.Json()

	c.JSON(200, vc)
}

func vpcsGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	if c.Query("names") == "true" {
		query := &bson.M{}

		vpcs, err := vpc.GetAllNames(db, query)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		c.JSON(200, vpcs)
	} else {
		page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
		pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

		query := bson.M{}

		vpcId, ok := utils.ParseObjectId(c.Query("id"))
		if ok {
			query["_id"] = vpcId
		}

		name := strings.TrimSpace(c.Query("name"))
		if name != "" {
			query["name"] = &bson.M{
				"$regex":   fmt.Sprintf(".*%s.*", regexp.QuoteMeta(name)),
				"$options": "i",
			}
		}

		network := strings.TrimSpace(c.Query("network"))
		if network != "" {
			query["network"] = network
		}

		organization, ok := utils.ParseObjectId(c.Query("organization"))
		if ok {
			query["organization"] = organization
		}

		dc, ok := utils.ParseObjectId(c.Query("datacenter"))
		if ok {
			query["datacenter"] = dc
		}

		vpcs, count, err := vpc.GetAllPaged(db, &query, page, pageCount)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		for _, vc := range vpcs {
			vc.Json()
		}

		data := &vpcsData{
			Vpcs:  vpcs,
			Count: count,
		}

		c.JSON(200, data)
	}
}
