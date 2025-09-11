package secondary

import "github.com/pritunl/mongo-go-driver/v2/bson"

const (
	Duo      = "duo"
	OneLogin = "one_login"
	Okta     = "okta"
	Push     = "push"
	Phone    = "phone"
	Passcode = "passcode"
	Sms      = "sms"

	Admin                    = "admin"
	AdminDevice              = "admin_device"
	AdminDeviceRegister      = "admin_device_register"
	User                     = "user"
	UserDevice               = "user_device"
	UserDeviceRegister       = "user_device_register"
	UserManage               = "user_manage"
	UserManageDevice         = "user_manage_device"
	UserManageDeviceRegister = "user_manage_device_register"
)

var (
	DeviceProvider, _ = bson.ObjectIDFromHex("100000000000000000000000")
)
