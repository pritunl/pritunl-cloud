package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/imds/server/state"
)

func primaryPut(c *gin.Context) {
	state.Global.SetPrimary()

	c.JSON(200, map[string]string{})
}
