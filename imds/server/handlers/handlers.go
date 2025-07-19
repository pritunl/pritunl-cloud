package handlers

import (
	"crypto/subtle"
	"net/http"

	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/imds/server/constants"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
	"github.com/pritunl/tools/logger"
)

type AuthenticationError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func Recovery(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			err := &errortypes.UnknownError{
				errors.Newf("handlers: Handler panic %s", r),
			}
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("handlers: Handler panic")
			c.Writer.WriteHeader(http.StatusInternalServerError)
		}
	}()

	c.Next()
}

func Errors(c *gin.Context) {
	c.Next()
	for _, err := range c.Errors {
		logger.WithFields(logger.Fields{
			"error": err,
		}).Error("handlers: Handler error")
	}
}

func AuthVirt(c *gin.Context) {
	token := c.Request.Header.Get("Auth-Token")
	if token == "" {
		token = c.Query("token")
	}

	// TODO config.Config.ClientIps not loaded
	// addr := utils.StripPort(c.Request.RemoteAddr)
	// if len(config.Config.ClientIps) != 0 && config.Config.ClientIps[0] == "" &&
	// 	!utils.StringsContains(config.Config.ClientIps, addr) {

	// 	c.AbortWithStatusJSON(401, &AuthenticationError{
	// 		Error:   "authentication",
	// 		Message: "Source IP address invalid",
	// 	})
	// 	return
	// }

	if c.Request.Header.Get("Origin") != "" ||
		c.Request.Header.Get("Referer") != "" ||
		c.Request.Header.Get("User-Agent") != "pritunl-imds" ||
		constants.ClientSecret == "" ||
		(subtle.ConstantTimeCompare([]byte(token),
			[]byte(constants.ClientSecret)) != 1) {

		c.AbortWithStatus(401)
		return
	}
	c.Next()
}

func AuthDhcp(c *gin.Context) {
	token := c.Request.Header.Get("Auth-Token")
	if token == "" {
		token = c.Query("token")
	}

	if c.Request.Header.Get("Origin") != "" ||
		c.Request.Header.Get("Referer") != "" ||
		c.Request.Header.Get("User-Agent") != "pritunl-dhcp" ||
		constants.DhcpSecret == "" ||
		(subtle.ConstantTimeCompare([]byte(token),
			[]byte(constants.DhcpSecret)) != 1) {

		c.AbortWithStatus(401)
		return
	}
	c.Next()
}

func AuthHost(c *gin.Context) {
	token := c.Request.Header.Get("Auth-Token")
	if token == "" {
		token = c.Query("token")
	}

	if c.Request.Header.Get("Origin") != "" ||
		c.Request.Header.Get("Referer") != "" ||
		c.Request.Header.Get("User-Agent") != "pritunl-imds" ||
		constants.HostSecret == "" ||
		(subtle.ConstantTimeCompare([]byte(token),
			[]byte(constants.HostSecret)) != 1) {

		c.AbortWithStatus(401)
		return
	}
	c.Next()
}

func RegisterVirt(engine *gin.Engine) {
	engine.Use(AuthVirt)
	engine.Use(Recovery)
	engine.Use(Errors)

	engine.GET("/query/:resource", queryGet)
	engine.GET("/query/:resource/:key1", queryGet)
	engine.GET("/query/:resource/:key1/:key2", queryGet)
	engine.GET("/query/:resource/:key1/:key2/:key3", queryGet)
	engine.GET("/query/:resource/:key1/:key2/:key3/:key4", queryGet)
	engine.GET("/instance", instanceGet)
	engine.GET("/vpc", vpcGet)
	engine.GET("/subnet", subnetGet)
	engine.GET("/certificate", certificatesGet)
	engine.GET("/secret", secretsGet)
	engine.PUT("/sync", syncPut)
}

func RegisterHost(engine *gin.Engine) {
	engine.Use(AuthHost)
	engine.Use(Recovery)
	engine.Use(Errors)

	engine.PUT("/sync", hostSyncPut)
	engine.GET("/sync", hostSyncGet)
}
