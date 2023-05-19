package ahandlers

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/secondary"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
)

type settingsData struct {
	AuthProviders             []*settings.Provider          `json:"auth_providers"`
	AuthSecondaryProviders    []*settings.SecondaryProvider `json:"auth_secondary_providers"`
	AuthAdminExpire           int                           `json:"auth_admin_expire"`
	AuthAdminMaxDuration      int                           `json:"auth_admin_max_duration"`
	AuthProxyExpire           int                           `json:"auth_proxy_expire"`
	AuthProxyMaxDuration      int                           `json:"auth_proxy_max_duration"`
	AuthUserExpire            int                           `json:"auth_user_expire"`
	AuthUserMaxDuration       int                           `json:"auth_user_max_duration"`
	AuthFastLogin             bool                          `json:"auth_fast_login"`
	AuthForceFastUserLogin    bool                          `json:"auth_force_fast_user_login"`
	AuthForceFastServiceLogin bool                          `json:"auth_force_fast_service_login"`
}

func getSettingsData() *settingsData {
	data := &settingsData{
		AuthProviders:          settings.Auth.Providers,
		AuthSecondaryProviders: settings.Auth.SecondaryProviders,
		AuthAdminExpire:        settings.Auth.AdminExpire,
		AuthAdminMaxDuration:   settings.Auth.AdminMaxDuration,
		AuthUserExpire:         settings.Auth.UserExpire,
		AuthUserMaxDuration:    settings.Auth.UserMaxDuration,
		AuthFastLogin:          settings.Auth.FastLogin,
		AuthForceFastUserLogin: settings.Auth.ForceFastUserLogin,
	}

	return data
}

func settingsGet(c *gin.Context) {
	data := getSettingsData()
	c.JSON(200, data)
}

func settingsPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &settingsData{}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	fields := set.NewSet(
		"providers",
		"secondary_providers",
	)

	if settings.Auth.AdminExpire != data.AuthAdminExpire {
		settings.Auth.AdminExpire = data.AuthAdminExpire
		fields.Add("admin_expire")
	}
	if settings.Auth.AdminMaxDuration != data.AuthAdminMaxDuration {
		settings.Auth.AdminMaxDuration = data.AuthAdminMaxDuration
		fields.Add("admin_max_duration")
	}
	if settings.Auth.UserExpire != data.AuthUserExpire {
		settings.Auth.UserExpire = data.AuthUserExpire
		fields.Add("user_expire")
	}
	if settings.Auth.UserMaxDuration != data.AuthUserMaxDuration {
		settings.Auth.UserMaxDuration = data.AuthUserMaxDuration
		fields.Add("user_max_duration")
	}
	if settings.Auth.FastLogin != data.AuthFastLogin {
		settings.Auth.FastLogin = data.AuthFastLogin
		fields.Add("auth_fast_login")
	}
	if settings.Auth.ForceFastUserLogin != data.AuthForceFastUserLogin {
		settings.Auth.ForceFastUserLogin = data.AuthForceFastUserLogin
		fields.Add("auth_force_fast_user_login")
	}

	for _, provider := range data.AuthProviders {
		provider.Label = utils.FilterStr(provider.Label, 32)

		if provider.Id.IsZero() {
			provider.Id = primitive.NewObjectID()
		}
	}
	settings.Auth.Providers = data.AuthProviders

	for _, provider := range data.AuthSecondaryProviders {
		provider.Name = utils.FilterStr(provider.Name, 32)
		provider.Label = utils.FilterStr(provider.Label, 32)

		if provider.Id.IsZero() {
			provider.Id = primitive.NewObjectID()
		}

		if provider.Type == secondary.OneLogin &&
			provider.OneLoginRegion == "" {

			provider.OneLoginRegion = "us"
		}
	}
	settings.Auth.SecondaryProviders = data.AuthSecondaryProviders

	err = settings.Commit(db, settings.Auth, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "settings.change")

	data = getSettingsData()
	c.JSON(200, data)
}
