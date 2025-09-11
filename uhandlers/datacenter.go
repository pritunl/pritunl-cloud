package uhandlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/utils"
)

func datacentersGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectID)

	dcs, err := datacenter.GetAllNamesOrg(db, userOrg)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, dcs)
}
