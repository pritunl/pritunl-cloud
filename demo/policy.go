package demo

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/policy"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Policies = []*policy.Policy{
	{
		Id:       utils.ObjectIdHex("67b8a03e4866ba90e6c45a8c"),
		Name:     "policy",
		Comment:  "",
		Disabled: false,
		Roles: []string{
			"pritunl",
		},
		Rules: map[string]*policy.Rule{
			"location": {
				Type:    "location",
				Disable: false,
				Values: []string{
					"US",
				},
			},
			"whitelist_networks": {
				Type:    "whitelist_networks",
				Disable: false,
				Values: []string{
					"10.0.0.0/8",
				},
			},
		},
		AdminSecondary:       primitive.ObjectID{},
		UserSecondary:        primitive.ObjectID{},
		AdminDeviceSecondary: true,
		UserDeviceSecondary:  true,
	},
}
