package ahandlers

import (
	"fmt"
	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
)

type nodeData struct {
	Id                 bson.ObjectId   `json:"id"`
	Zone               bson.ObjectId   `json:"zone"`
	Name               string          `json:"name"`
	Type               string          `json:"type"`
	Port               int             `json:"port"`
	Protocol           string          `json:"protocol"`
	Certificates       []bson.ObjectId `json:"certificates"`
	AdminDomain        string          `json:"admin_domain"`
	UserDomain         string          `json:"user_domain"`
	Services           []bson.ObjectId `json:"services"`
	ForwardedForHeader string          `json:"forwarded_for_header"`
}

type nodesData struct {
	Nodes []*node.Node `json:"nodes"`
	Count int          `json:"count"`
}

func nodePut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &nodeData{}

	nodeId, ok := utils.ParseObjectId(c.Param("node_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	nde, err := node.Get(db, nodeId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	nde.Name = data.Name
	nde.Type = data.Type
	nde.Port = data.Port
	nde.Protocol = data.Protocol
	nde.Certificates = data.Certificates
	nde.AdminDomain = data.AdminDomain
	nde.UserDomain = data.UserDomain
	nde.ForwardedForHeader = data.ForwardedForHeader

	fields := set.NewSet(
		"name",
		"type",
		"port",
		"protocol",
		"certificates",
		"admin_domain",
		"user_domain",
		"forwarded_for_header",
	)

	if data.Zone != "" && data.Zone != nde.Zone {
		if nde.Zone != "" {
			errData := &errortypes.ErrorData{
				Error:   "zone_modified",
				Message: "Cannot modify zone once set",
			}
			c.JSON(400, errData)
			return
		}
		nde.Zone = data.Zone
		fields.Add("zone")
	}

	errData, err := nde.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = nde.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "node.change")

	c.JSON(200, nde)
}

func nodeDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	nodeId, ok := utils.ParseObjectId(c.Param("node_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := node.Remove(db, nodeId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "node.change")

	c.JSON(200, nil)
}

func nodeGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	nodeId, ok := utils.ParseObjectId(c.Param("node_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	nde, err := node.Get(db, nodeId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if demo.IsDemo() {
		nde.RequestsMin = 32
		nde.Memory = 25.0
		nde.Load1 = 10.0
		nde.Load5 = 15.0
		nde.Load15 = 20.0
	}

	c.JSON(200, nde)
}

func nodesGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	page, _ := strconv.Atoi(c.Query("page"))
	pageCount, _ := strconv.Atoi(c.Query("page_count"))

	query := bson.M{}

	name := strings.TrimSpace(c.Query("name"))
	if name != "" {
		query["name"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", name),
			"$options": "i",
		}
	}

	nodes, count, err := node.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if demo.IsDemo() {
		for _, nde := range nodes {
			nde.RequestsMin = 32
			nde.Memory = 25.0
			nde.Load1 = 10.0
			nde.Load5 = 15.0
			nde.Load15 = 20.0
		}
	}

	data := &nodesData{
		Nodes: nodes,
		Count: count,
	}

	c.JSON(200, data)
}
