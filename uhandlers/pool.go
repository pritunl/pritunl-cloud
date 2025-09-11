package uhandlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/utils"
)

func poolsGet(c *gin.Context) {
	if demo.IsDemo() {
		c.JSON(200, demo.Pools)
		return
	}

	db := c.MustGet("db").(*database.Database)

	pools, err := pool.GetAllNames(db, &bson.M{})
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, pools)
}
