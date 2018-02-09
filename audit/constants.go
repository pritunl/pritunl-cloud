package audit

const (
	AdminLogin       = "admin_login"
	AdminLoginFailed = "admin_login_failed"
	UserLogin        = "user_login"
	UserLoginFailed  = "user_login_failed"
	DuoApprove       = "duo_approve"
	DuoDeny          = "duo_deny"
	OneLoginApprove  = "one_login_approve"
	OneLoginDeny     = "one_login_deny"
	OktaApprove      = "okta_approve"
	OktaDeny         = "okta_deny"
)
