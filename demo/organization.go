package demo

import (
	"github.com/pritunl/pritunl-cloud/organization"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Organizations = []*organization.Organization{
	{
		Id: utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
		Roles: []string{
			"pritunl",
		},
		Name:    "pritunl",
		Comment: "",
	},
}
