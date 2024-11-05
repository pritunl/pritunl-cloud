package router

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/imds/server/config"
	"github.com/pritunl/pritunl-cloud/imds/server/constants"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/server/handlers"
	"github.com/pritunl/tools/logger"
)

type Router struct {
	server *http.Server
}

func (r *Router) Run() (err error) {
	logger.WithFields(logger.Fields{
		"config":  config.Path,
		"host":    constants.Host,
		"port":    constants.Port,
		"version": constants.Version,
	}).Info("main: Starting imds server")

	err = r.server.ListenAndServe()
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "main: Server listen error"),
		}
		return
	}

	return
}

func (r *Router) Shutdown() {
	defer func() {
		recover()
	}()

	webCtx, webCancel := context.WithTimeout(
		context.Background(),
		1*time.Second,
	)
	defer webCancel()

	_ = r.server.Shutdown(webCtx)
	_ = r.server.Close()
}

func (r *Router) Init() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	handlers.Register(router)

	r.server = &http.Server{
		Addr: fmt.Sprintf(
			"%s:%d",
			constants.Host,
			constants.Port,
		),
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 4096,
	}

	return
}
