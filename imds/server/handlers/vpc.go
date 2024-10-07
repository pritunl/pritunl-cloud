package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/imds/server/config"
)

func vpcGet(c *gin.Context) {
	c.JSON(200, config.Config.Vpc)
}

func subnetGet(c *gin.Context) {
	c.JSON(200, config.Config.Subnet)
}
