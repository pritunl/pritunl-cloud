package handlers

import (
	"github.com/gin-gonic/gin"
)

type instanceData struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

func instanceGet(c *gin.Context) {
	data := &instanceData{
		Name:  "test",
		State: "start",
	}
	c.JSON(200, data)
}
