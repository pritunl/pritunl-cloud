package uhandlers

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/authorizer"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/utils"
)

type themeData struct {
	Theme       string `json:"theme"`
	EditorTheme string `json:"editor_theme"`
}

func themePut(c *gin.Context) {
	if demo.IsDemo() {
		c.JSON(200, nil)
		return
	}

	db := c.MustGet("db").(*database.Database)
	authr := c.MustGet("authorizer").(*authorizer.Authorizer)
	data := &themeData{}

	err := c.Bind(&data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	usr, err := authr.GetUser(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	usr.Theme = data.Theme
	usr.EditorTheme = data.EditorTheme

	err = usr.CommitFields(db, set.NewSet("theme", "editor_theme"))
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, data)
	return
}
