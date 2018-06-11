package secondary

import (
	"gopkg.in/mgo.v2/bson"
)

const (
	Duo      = "duo"
	OneLogin = "one_login"
	Okta     = "okta"
	Push     = "push"
	Phone    = "phone"
	Passcode = "passcode"
	Sms      = "sms"
	Device   = "device"

	Admin = "admin"
	User  = "user"
)

var (
	DeviceProvider = bson.ObjectIdHex("100000000000000000000000")
)
