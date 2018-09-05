package uhandlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
)

func domainsGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(bson.ObjectId)

	query := &bson.M{
		"organization": userOrg,
	}

	domns, err := domain.GetAllName(db, query)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, domns)
}
