package ahandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/config"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/middlewear"
	"github.com/pritunl/pritunl-cloud/requires"
	"github.com/pritunl/pritunl-cloud/static"
)

var (
	store      *static.Store
	fileServer http.Handler
)

func Register(engine *gin.Engine) {
	engine.Use(middlewear.Limiter)
	engine.Use(middlewear.Counter)
	engine.Use(middlewear.Recovery)
	engine.Use(middlewear.Headers)

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

	csrfGroup.GET("/alert", alertsGet)
	csrfGroup.PUT("/alert/:alert_id", alertPut)
	csrfGroup.POST("/alert", alertPost)
	csrfGroup.DELETE("/alert", alertsDelete)
	csrfGroup.DELETE("/alert/:alert_id", alertDelete)

	engine.GET("/auth/state", authStateGet)
	dbGroup.POST("/auth/session", authSessionPost)
	dbGroup.POST("/auth/secondary", authSecondaryPost)
	dbGroup.GET("/auth/request", authRequestGet)
	dbGroup.GET("/auth/callback", authCallbackGet)
	dbGroup.GET("/auth/webauthn/request", authWanRequestGet)
	dbGroup.POST("/auth/webauthn/respond", authWanRespondPost)
	dbGroup.GET("/auth/webauthn/register", authWanRegisterGet)
	dbGroup.POST("/auth/webauthn/register", authWanRegisterPost)
	sessGroup.GET("/logout", logoutGet)

	csrfGroup.GET("/authority", authoritiesGet)
	csrfGroup.GET("/authority/:authority_id", authorityGet)
	csrfGroup.PUT("/authority/:authority_id", authorityPut)
	csrfGroup.POST("/authority", authorityPost)
	csrfGroup.DELETE("/authority", authoritiesDelete)
	csrfGroup.DELETE("/authority/:authority_id", authorityDelete)

	csrfGroup.GET("/balancer", balancersGet)
	csrfGroup.GET("/balancer/:balancer_id", balancerGet)
	csrfGroup.PUT("/balancer/:balancer_id", balancerPut)
	csrfGroup.POST("/balancer", balancerPost)
	csrfGroup.DELETE("/balancer", balancersDelete)
	csrfGroup.DELETE("/balancer/:balancer_id", balancerDelete)

	csrfGroup.GET("/block", blocksGet)
	csrfGroup.GET("/block/:block_id", blockGet)
	csrfGroup.PUT("/block/:block_id", blockPut)
	csrfGroup.POST("/block", blockPost)
	csrfGroup.DELETE("/block/:block_id", blockDelete)

	csrfGroup.GET("/certificate", certificatesGet)
	csrfGroup.GET("/certificate/:cert_id", certificateGet)
	csrfGroup.PUT("/certificate/:cert_id", certificatePut)
	csrfGroup.POST("/certificate", certificatePost)
	csrfGroup.DELETE("/certificate/:cert_id", certificateDelete)

	engine.GET("/check", checkGet)

	authGroup.GET("/csrf", csrfGet)

	csrfGroup.GET("/completion", completionGet)

	csrfGroup.GET("/datacenter", datacentersGet)
	csrfGroup.GET("/datacenter/:dc_id", datacenterGet)
	csrfGroup.PUT("/datacenter/:dc_id", datacenterPut)
	csrfGroup.POST("/datacenter", datacenterPost)
	csrfGroup.DELETE("/datacenter/:dc_id", datacenterDelete)

	csrfGroup.GET("/device/:user_id", devicesGet)
	csrfGroup.PUT("/device/:device_id", devicePut)
	csrfGroup.POST("/device", devicePost)
	csrfGroup.DELETE("/device/:device_id", deviceDelete)
	csrfGroup.POST("/device/:resource_id/:method", deviceMethodPost)
	csrfGroup.GET("/device/:user_id/webauthn/register", deviceWanRegisterGet)
	csrfGroup.POST("/device/:resource_id/webauthn/register",
		deviceWanRegisterPost)

	csrfGroup.GET("/disk", disksGet)
	csrfGroup.GET("/disk/:disk_id", diskGet)
	csrfGroup.PUT("/disk", disksPut)
	csrfGroup.PUT("/disk/:disk_id", diskPut)
	csrfGroup.POST("/disk", diskPost)
	csrfGroup.DELETE("/disk", disksDelete)
	csrfGroup.DELETE("/disk/:disk_id", diskDelete)

	csrfGroup.GET("/domain", domainsGet)
	csrfGroup.GET("/domain/:domain_id", domainGet)
	csrfGroup.PUT("/domain/:domain_id", domainPut)
	csrfGroup.POST("/domain", domainPost)
	csrfGroup.DELETE("/domain", domainsDelete)
	csrfGroup.DELETE("/domain/:domain_id", domainDelete)

	csrfGroup.GET("/event", eventGet)

	csrfGroup.GET("/firewall", firewallsGet)
	csrfGroup.GET("/firewall/:firewall_id", firewallGet)
	csrfGroup.PUT("/firewall/:firewall_id", firewallPut)
	csrfGroup.POST("/firewall", firewallPost)
	csrfGroup.DELETE("/firewall", firewallsDelete)
	csrfGroup.DELETE("/firewall/:firewall_id", firewallDelete)

	csrfGroup.GET("/image", imagesGet)
	csrfGroup.GET("/image/:image_id", imageGet)
	csrfGroup.PUT("/image/:image_id", imagePut)
	csrfGroup.DELETE("/image", imagesDelete)
	csrfGroup.DELETE("/image/:image_id", imageDelete)

	csrfGroup.GET("/instance", instancesGet)
	csrfGroup.PUT("/instance", instancesPut)
	csrfGroup.GET("/instance/:instance_id", instanceGet)
	csrfGroup.GET("/instance/:instance_id/vnc", instanceVncGet)
	csrfGroup.PUT("/instance/:instance_id", instancePut)
	csrfGroup.POST("/instance", instancePost)
	csrfGroup.DELETE("/instance", instancesDelete)
	csrfGroup.DELETE("/instance/:instance_id", instanceDelete)

	csrfGroup.PUT("/license", licensePut)

	csrfGroup.GET("/log", logsGet)
	csrfGroup.GET("/log/:log_id", logGet)

	csrfGroup.GET("/node", nodesGet)
	csrfGroup.GET("/node/:node_id", nodeGet)
	csrfGroup.PUT("/node/:node_id", nodePut)
	csrfGroup.PUT("/node/:node_id/:operation", nodeOperationPut)
	csrfGroup.POST("/node/:node_id/init", nodeInitPost)
	csrfGroup.DELETE("/node/:node_id", nodeDelete)

	csrfGroup.GET("/organization", organizationsGet)
	csrfGroup.GET("/organization/:org_id", organizationGet)
	csrfGroup.PUT("/organization/:org_id", organizationPut)
	csrfGroup.POST("/organization", organizationPost)
	csrfGroup.DELETE("/organization/:org_id", organizationDelete)

	csrfGroup.GET("/plan", plansGet)
	csrfGroup.GET("/plan/:plan_id", planGet)
	csrfGroup.PUT("/plan/:plan_id", planPut)
	csrfGroup.POST("/plan", planPost)
	csrfGroup.DELETE("/plan", plansDelete)
	csrfGroup.DELETE("/plan/:plan_id", planDelete)

	csrfGroup.GET("/policy", policiesGet)
	csrfGroup.GET("/policy/:policy_id", policyGet)
	csrfGroup.PUT("/policy/:policy_id", policyPut)
	csrfGroup.POST("/policy", policyPost)
	csrfGroup.DELETE("/policy/:policy_id", policyDelete)

	csrfGroup.GET("/pool", poolsGet)
	csrfGroup.GET("/pool/:pool_id", poolGet)
	csrfGroup.PUT("/pool/:pool_id", poolPut)
	csrfGroup.POST("/pool", poolPost)
	csrfGroup.DELETE("/pool", poolsDelete)
	csrfGroup.DELETE("/pool/:pool_id", poolDelete)

	csrfGroup.GET("/secret", secretsGet)
	csrfGroup.GET("/secret/:secr_id", secretGet)
	csrfGroup.PUT("/secret/:secr_id", secretPut)
	csrfGroup.POST("/secret", secretPost)
	csrfGroup.DELETE("/secret/:secr_id", secretDelete)

	csrfGroup.GET("/session/:user_id", sessionsGet)
	csrfGroup.DELETE("/session/:session_id", sessionDelete)

	csrfGroup.GET("/settings", settingsGet)
	csrfGroup.PUT("/settings", settingsPut)

	csrfGroup.GET("/pod", podsGet)
	csrfGroup.GET("/pod/:pod_id", podGet)
	csrfGroup.PUT("/pod/:pod_id", podPut)
	csrfGroup.POST("/pod", podPost)
	csrfGroup.DELETE("/pod", podsDelete)
	csrfGroup.DELETE("/pod/:pod_id", podDelete)
	csrfGroup.GET("/pod/:pod_id/unit/:unit_id", podUnitGet)
	csrfGroup.PUT("/pod/:pod_id/unit/:unit_id/deployment",
		podUnitDeploymentsPut)
	csrfGroup.POST("/pod/:pod_id/unit/:unit_id/deployment",
		podUnitDeploymentPost)
	csrfGroup.PUT("/pod/:pod_id/unit/:unit_id/deployment/:deployment_id",
		podUnitDeploymentPut)
	csrfGroup.GET(
		"/pod/:pod_id/unit/:unit_id/deployment/:deployment_id/log",
		podUnitDeploymentLogGet,
	)
	csrfGroup.GET("/pod/:pod_id/unit/:unit_id/spec", podUnitSpecsGet)
	csrfGroup.GET("/pod/:pod_id/unit/:unit_id/spec/:spec_id", podUnitSpecGet)

	csrfGroup.GET("/shape", shapesGet)
	csrfGroup.GET("/shape/:shape_id", shapeGet)
	csrfGroup.PUT("/shape/:shape_id", shapePut)
	csrfGroup.POST("/shape", shapePost)
	csrfGroup.DELETE("/shape", shapesDelete)
	csrfGroup.DELETE("/shape/:shape_id", shapeDelete)

	csrfGroup.GET("/storage", storagesGet)
	csrfGroup.GET("/storage/:store_id", storageGet)
	csrfGroup.PUT("/storage/:store_id", storagePut)
	csrfGroup.POST("/storage", storagePost)
	csrfGroup.DELETE("/storage/:store_id", storageDelete)

	csrfGroup.GET("/subscription", subscriptionGet)
	csrfGroup.GET("/subscription/update", subscriptionUpdateGet)
	csrfGroup.POST("/subscription", subscriptionPost)

	csrfGroup.PUT("/theme", themePut)

	csrfGroup.GET("/user", usersGet)
	csrfGroup.GET("/user/:user_id", userGet)
	csrfGroup.PUT("/user/:user_id", userPut)
	csrfGroup.POST("/user", userPost)
	csrfGroup.DELETE("/user", usersDelete)

	csrfGroup.GET("/vpc", vpcsGet)
	csrfGroup.GET("/vpc/:vpc_id", vpcGet)
	csrfGroup.PUT("/vpc/:vpc_id", vpcPut)
	csrfGroup.GET("/vpc/:vpc_id/routes", vpcRoutesGet)
	csrfGroup.PUT("/vpc/:vpc_id/routes", vpcRoutesPut)
	csrfGroup.POST("/vpc", vpcPost)
	csrfGroup.DELETE("/vpc", vpcsDelete)
	csrfGroup.DELETE("/vpc/:vpc_id", vpcDelete)

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

		sessGroup.GET("/", staticTestingGet)
		engine.GET("/login", staticTestingGet)
		engine.GET("/logo.png", staticTestingGet)
		authGroup.GET("/static/*path", staticTestingGet)
	}
}

func init() {
	module := requires.New("ahandlers")
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
