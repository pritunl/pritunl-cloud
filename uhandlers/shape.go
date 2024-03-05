package uhandlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/utils"
)

func shapesGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	shapes, err := shape.GetAllNames(db, &bson.M{})
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, shapes)
}
