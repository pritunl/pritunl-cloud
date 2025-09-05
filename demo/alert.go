package demo

import (
	"github.com/pritunl/pritunl-cloud/alert"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Alerts = []*alert.Alert{
	{
		Id:           utils.ObjectIdHex("9cc278e67d0b4a3d173280c0"),
		Name:         "cloud",
		Comment:      "",
		Organization: utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
		Roles:        []string{"instance"},
		Resource:     "instance_offline",
		Level:        5,
		Frequency:    300,
	},
}
