package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/imds/resource"
)

func queryGet(c *gin.Context) {
	resrc := c.Param("resource")
	key1 := c.Param("key1")
	key2 := c.Param("key2")
	key3 := c.Param("key3")
	key4 := c.Param("key4")

	keys := []string{}
	if key1 != "" {
		keys = append(keys, key1)
		if key2 != "" {
			keys = append(keys, key2)
			if key3 != "" {
				keys = append(keys, key3)
				if key4 != "" {
					keys = append(keys, key4)
				}
			}
		}
	}

	val, err := resource.Query(resrc, keys...)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.String(200, val)
}
