package router

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/imds/server/constants"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/server/handlers"
	"github.com/pritunl/tools/logger"
)

type Router struct {
	virtServer *http.Server
	hostServer *http.Server
}

func (r *Router) Run() (err error) {
	logger.WithFields(logger.Fields{
		"host":    constants.Host,
		"port":    constants.Port,
		"sock":    constants.Sock,
		"version": constants.Version,
	}).Info("main: Starting imds server")

	waiters := &sync.WaitGroup{}
	waiters.Add(2)

	go func() {
		defer waiters.Done()

		e := r.virtServer.ListenAndServe()
		if e != nil {
			e = &errortypes.WriteError{
				errors.Wrap(e, "main: Server listen error"),
			}
			if err == nil {
				err = e
			}
			r.Shutdown()
			return
		}
	}()

	go func() {
		defer waiters.Done()

		listener, e := net.Listen("unix", constants.Sock)
		if e != nil {
			e = &errortypes.WriteError{
				errors.Wrap(e, "main: Failed to create unix socket"),
			}
			if err == nil {
				err = e
			}
			r.Shutdown()
			return
		}

		e = r.hostServer.Serve(listener)
		if e != nil {
			e = &errortypes.WriteError{
				errors.Wrap(e, "main: Server listen error"),
			}
			if err == nil {
				err = e
			}
			r.Shutdown()
			return
		}
	}()

	waiters.Wait()
	if err != nil {
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

	_ = r.virtServer.Shutdown(webCtx)
	_ = r.virtServer.Close()

	_ = r.hostServer.Shutdown(webCtx)
	_ = r.hostServer.Close()
}

func (r *Router) Init() {
	gin.SetMode(gin.ReleaseMode)

	virtRouter := gin.New()
	handlers.RegisterVirt(virtRouter)

	r.virtServer = &http.Server{
		Addr: fmt.Sprintf(
			"%s:%d",
			constants.Host,
			constants.Port,
		),
		Handler:        virtRouter,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 4096,
	}

	hostRouter := gin.New()
	handlers.RegisterHost(hostRouter)

	r.hostServer = &http.Server{
		Addr:           "127.0.0.1:99999",
		Handler:        hostRouter,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 4096,
	}

	return
}
