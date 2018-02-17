package ahandlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/config"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/middlewear"
	"github.com/pritunl/pritunl-cloud/requires"
	"github.com/pritunl/pritunl-cloud/static"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	store      *static.Store
	fileServer http.Handler
	pushFiles  []string
)

func Register(engine *gin.Engine) {
	engine.Use(middlewear.Limiter)
	engine.Use(middlewear.Counter)
	engine.Use(middlewear.Recovery)

	dbGroup := engine.Group("")
	dbGroup.Use(middlewear.Database)

	sessGroup := dbGroup.Group("")
	sessGroup.Use(middlewear.SessionAdmin)

	authGroup := sessGroup.Group("")
	authGroup.Use(middlewear.AuthAdmin)

	csrfGroup := authGroup.Group("")
	csrfGroup.Use(middlewear.CsrfToken)

	engine.NoRoute(middlewear.NotFound)

	csrfGroup.GET("/audit/:user_id", auditsGet)

	engine.GET("/auth/state", authStateGet)
	dbGroup.POST("/auth/session", authSessionPost)
	dbGroup.POST("/auth/secondary", authSecondaryPost)
	dbGroup.GET("/auth/request", authRequestGet)
	dbGroup.GET("/auth/callback", authCallbackGet)
	sessGroup.GET("/logout", logoutGet)

	csrfGroup.GET("/certificate", certificatesGet)
	csrfGroup.GET("/certificate/:cert_id", certificateGet)
	csrfGroup.PUT("/certificate/:cert_id", certificatePut)
	csrfGroup.POST("/certificate", certificatePost)
	csrfGroup.DELETE("/certificate/:cert_id", certificateDelete)

	engine.GET("/check", checkGet)

	authGroup.GET("/csrf", csrfGet)

	csrfGroup.GET("/datacenter", datacentersGet)
	csrfGroup.GET("/datacenter/:dc_id", datacenterGet)
	csrfGroup.PUT("/datacenter/:dc_id", datacenterPut)
	csrfGroup.POST("/datacenter", datacenterPost)
	csrfGroup.DELETE("/datacenter/:dc_id", datacenterDelete)

	csrfGroup.GET("/event", eventGet)

	csrfGroup.GET("/instance", instancesGet)
	csrfGroup.PUT("/instance", instancesPut)
	csrfGroup.GET("/instance/:instance_id", instanceGet)
	csrfGroup.PUT("/instance/:instance_id", instancePut)
	csrfGroup.POST("/instance", instancePost)
	csrfGroup.DELETE("/instance", instancesDelete)
	csrfGroup.DELETE("/instance/:instance_id", instanceDelete)

	csrfGroup.GET("/log", logsGet)
	csrfGroup.GET("/log/:log_id", logGet)

	csrfGroup.GET("/node", nodesGet)
	csrfGroup.GET("/node/:node_id", nodeGet)
	csrfGroup.PUT("/node/:node_id", nodePut)
	csrfGroup.DELETE("/node/:node_id", nodeDelete)

	csrfGroup.GET("/organization", organizationsGet)
	csrfGroup.GET("/organization/:org_id", organizationGet)
	csrfGroup.PUT("/organization/:org_id", organizationPut)
	csrfGroup.POST("/organization", organizationPost)
	csrfGroup.DELETE("/organization/:org_id", organizationDelete)

	csrfGroup.GET("/policy", policiesGet)
	csrfGroup.GET("/policy/:policy_id", policyGet)
	csrfGroup.PUT("/policy/:policy_id", policyPut)
	csrfGroup.POST("/policy", policyPost)
	csrfGroup.DELETE("/policy/:policy_id", policyDelete)

	csrfGroup.GET("/session/:user_id", sessionsGet)
	csrfGroup.DELETE("/session/:session_id", sessionDelete)

	csrfGroup.GET("/settings", settingsGet)
	csrfGroup.PUT("/settings", settingsPut)

	csrfGroup.GET("/subscription", subscriptionGet)
	csrfGroup.GET("/subscription/update", subscriptionUpdateGet)
	csrfGroup.POST("/subscription", subscriptionPost)

	csrfGroup.PUT("/theme", themePut)

	csrfGroup.GET("/user", usersGet)
	csrfGroup.GET("/user/:user_id", userGet)
	csrfGroup.PUT("/user/:user_id", userPut)
	csrfGroup.POST("/user", userPost)
	csrfGroup.DELETE("/user", usersDelete)

	csrfGroup.GET("/zone", zonesGet)
	csrfGroup.GET("/zone/:zone_id", zoneGet)
	csrfGroup.PUT("/zone/:zone_id", zonePut)
	csrfGroup.POST("/zone", zonePost)
	csrfGroup.DELETE("/zone/:zone_id", zoneDelete)

	engine.GET("/robots.txt", middlewear.RobotsGet)

	if constants.Production {
		sessGroup.GET("/", staticIndexGet)
		engine.GET("/login", staticLoginGet)
		engine.GET("/logo.png", staticLogoGet)
		authGroup.GET("/static/*path", staticGet)
	} else {
		fs := gin.Dir(config.StaticTestingRoot, false)
		fileServer = http.FileServer(fs)

		pushFiles = []string{}
		walk := path.Join(config.StaticTestingRoot, "aapp")
		err := filepath.Walk(walk, func(
			pth string, _ os.FileInfo, e error) (err error) {

			if e != nil {
				err = e
				return
			}

			if strings.HasSuffix(pth, ".js") ||
				strings.HasSuffix(pth, ".js.map") {

				pth = strings.Replace(pth, walk, "/aapp", 1)
				pushFiles = append(pushFiles, pth)
			}

			return
		})
		if err != nil {
			panic(err)
		}

		sessGroup.GET("/", staticTestingGet)
		engine.GET("/login", staticTestingGet)
		engine.GET("/logo.png", staticTestingGet)
		authGroup.GET("/config.js", staticTestingGet)
		authGroup.GET("/build.js", staticTestingGet)
		authGroup.GET("/aapp/*path", staticTestingGet)
		authGroup.GET("/dist/*path", staticTestingGet)
		authGroup.GET("/styles/*path", staticTestingGet)
		authGroup.GET("/node_modules/*path", staticTestingGet)
		authGroup.GET("/jspm_packages/*path", staticTestingGet)
	}
}

func init() {
	module := requires.New("mhandlers")
	module.After("settings")

	module.Handler = func() (err error) {
		if constants.Production {
			store, err = static.NewStore(config.StaticRoot)
			if err != nil {
				return
			}
		}

		return
	}
}
