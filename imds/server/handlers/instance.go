package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/imds/server/config"
)

func instanceGet(c *gin.Context) {
	c.JSON(200, config.Config.Instance)
}
