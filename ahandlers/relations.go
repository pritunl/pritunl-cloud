package ahandlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/relations"
	"github.com/pritunl/pritunl-cloud/utils"
)

type relationsData struct {
	Id   any    `json:"id"`
	Kind string `json:"kind"`
	Data string `json:"data"`
}

func relationsGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	kind := c.Param("kind")
	resourceId, ok := utils.ParseObjectId(c.Param("id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	resp, err := relations.Aggregate(db, kind, resourceId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if resp == nil {
		utils.AbortWithStatus(c, 404)
		return
	}

	data := &relationsData{
		Id:   resp.Id,
		Kind: kind,
		Data: resp.Yaml(),
	}
	c.JSON(200, data)
}
