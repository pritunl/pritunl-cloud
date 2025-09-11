package uhandlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/zone"
)

func nodesGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)

	zoneStr := c.Query("zone")
	if zoneStr == "" {
		c.JSON(200, []interface{}{})
		return
	}

	zneId, _ := utils.ParseObjectId(zoneStr)

	zne, err := zone.Get(db, zneId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	exists, err := datacenter.ExistsOrg(db, userOrg, zne.Datacenter)
	if err != nil {
		return
	}
	if !exists {
		utils.AbortWithStatus(c, 405)
		return
	}

	query := &bson.M{
		"zone": zneId,
	}

	nodes, err := node.GetAllHypervisors(db, query)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, nodes)
}
