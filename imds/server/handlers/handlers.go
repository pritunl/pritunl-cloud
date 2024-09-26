package handlers

import (
	"crypto/subtle"
	"net/http"

	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/imds/server/constants"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/server/logger"
)

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

func Auth(c *gin.Context) {
	token := ""
	if constants.Authenticated {
		token = c.Request.Header.Get("Auth-Token")
		if token == "" {
			token = c.Query("token")
		}
	}

	if c.Request.Header.Get("Origin") != "" ||
		c.Request.Header.Get("Referer") != "" ||
		c.Request.Header.Get("User-Agent") != "pritunl-imds" ||
		(constants.Authenticated && subtle.ConstantTimeCompare([]byte(token),
			[]byte(constants.AuthKey)) != 1) {

		c.AbortWithStatus(401)
		return
	}
	c.Next()
}

func Register(engine *gin.Engine) {
	engine.Use(Auth)
	engine.Use(Recovery)
	engine.Use(Errors)

	engine.GET("/instance", instanceGet)
}
