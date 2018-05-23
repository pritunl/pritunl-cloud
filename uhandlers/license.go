package uhandlers

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/authorizer"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/utils"
)

type licenseData struct {
	Oracle bool `json:"oracle"`
}

func licensePut(c *gin.Context) {
	if demo.IsDemo() {
		c.JSON(200, nil)
		return
	}

	db := c.MustGet("db").(*database.Database)
	authr := c.MustGet("authorizer").(*authorizer.Authorizer)
	data := &licenseData{}

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

	usr.OracleLicense = data.Oracle

	err = usr.CommitFields(db, set.NewSet("oracle_licese"))
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, data)
	return
}
