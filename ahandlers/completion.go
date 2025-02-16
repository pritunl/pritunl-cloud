package ahandlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/completion"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func completionGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	cmpl, err := completion.GetCompletion(db, primitive.NilObjectID)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, cmpl)
}
