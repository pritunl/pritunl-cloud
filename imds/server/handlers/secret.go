package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/imds/server/config"
)

func secretsGet(c *gin.Context) {
	c.JSON(200, config.Config.Secrets)
}
