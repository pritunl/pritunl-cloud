package uhandlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/authorizer"
	"github.com/pritunl/pritunl-cloud/completion"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/organization"
	"github.com/pritunl/pritunl-cloud/utils"
)

func completionGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	authr := c.MustGet("authorizer").(*authorizer.Authorizer)

	usr, err := authr.GetUser(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	var userOrg primitive.ObjectID
	orgIdStr := c.GetHeader("Organization")

	if orgIdStr != "" {
		orgId, ok := utils.ParseObjectId(orgIdStr)
		if !ok {
			utils.AbortWithStatus(c, 400)
			return
		}

		org, err := organization.Get(db, orgId)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		match := usr.RolesMatch(org.Roles)
		if !match {
			utils.AbortWithStatus(c, 401)
			return
		}

		userOrg = org.Id
	} else {
		orgs, err := organization.GetAll(db, &bson.M{
			"roles": &bson.M{
				"$in": usr.Roles,
			},
		})
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		if len(orgs) > 0 {
			org := orgs[0]

			match := usr.RolesMatch(org.Roles)
			if !match {
				utils.AbortWithStatus(c, 401)
				return
			}

			userOrg = org.Id
		}
	}

	if userOrg.IsZero() {
		utils.AbortWithStatus(c, 400)
		return
	}

	cmpl, err := completion.GetCompletion(db, userOrg, usr.Roles)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, cmpl)
}
