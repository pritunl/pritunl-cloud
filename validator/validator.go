package validator

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/policy"
	"github.com/pritunl/pritunl-cloud/user"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

func ValidateAdmin(db *database.Database, usr *user.User,
	isApi bool, r *http.Request) (deviceAuth bool, secProvider bson.ObjectId,
	errData *errortypes.ErrorData, err error) {

	if usr.Disabled || usr.Administrator != "super" {
		errData = &errortypes.ErrorData{
			Error:   "unauthorized",
			Message: "Not authorized",
		}
		return
	}

	if !isApi {
		policies, e := policy.GetRoles(db, usr.Roles)
		if e != nil {
			err = e
			return
		}

		for _, polcy := range policies {
			if polcy.AdminDevice {
				deviceAuth = true
			}

			if polcy.AdminSecondary != "" {
				secProvider = polcy.AdminSecondary
				break
			}
		}
	}

	return
}

func ValidateUser(db *database.Database, usr *user.User,
	isApi bool, r *http.Request) (deviceAuth bool, secProvider bson.ObjectId,
	errData *errortypes.ErrorData, err error) {

	if usr.Disabled {
		errData = &errortypes.ErrorData{
			Error:   "unauthorized",
			Message: "Not authorized",
		}
		return
	}

	if !isApi {
		policies, e := policy.GetRoles(db, usr.Roles)
		if e != nil {
			err = e
			return
		}

		for _, polcy := range policies {
			errData, err = polcy.ValidateUser(db, usr, r)
			if err != nil || errData != nil {
				return
			}
		}

		for _, polcy := range policies {
			if polcy.UserDevice {
				deviceAuth = true
			}

			if polcy.UserSecondary != "" {
				secProvider = polcy.UserSecondary
				break
			}
		}
	}

	return
}
