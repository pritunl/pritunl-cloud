package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/imds/resource"
)

func queryGet(c *gin.Context) {
	resrc := c.Param("resource")
	name := c.Param("name")
	key := c.Param("key")

	val, err := resource.Query(resrc, name, key)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.String(200, val)
}
