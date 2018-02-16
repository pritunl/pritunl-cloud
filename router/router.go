package router

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/acme"
	"github.com/pritunl/pritunl-cloud/ahandlers"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/uhandlers"
	"github.com/pritunl/pritunl-cloud/utils"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Router struct {
	nodeHash       []byte
	typ            string
	port           int
	protocol       string
	certificates   []*certificate.Certificate
	adminDomain    string
	userDomain     string
	aRouter        *gin.Engine
	uRouter        *gin.Engine
	waiter         sync.WaitGroup
	lock           sync.Mutex
	redirectServer *http.Server
	webServer      *http.Server
	stop           bool
}

func (r *Router) ServeHTTP(w http.ResponseWriter, re *http.Request) {
	hst := utils.StripPort(re.Host)
	if r.typ == node.Admin {
		r.aRouter.ServeHTTP(w, re)
		return
	} else if r.typ == node.User {
		r.uRouter.ServeHTTP(w, re)
		return
	} else if strings.Contains(r.typ, node.Admin) && hst == r.adminDomain {
		r.aRouter.ServeHTTP(w, re)
		return
	} else if strings.Contains(r.typ, node.User) && hst == r.userDomain {
		r.uRouter.ServeHTTP(w, re)
		return
	}

	if re.URL.Path == "/check" {
		utils.WriteText(w, 200, "ok")
		return
	}

	utils.WriteStatus(w, 404)
}

func (r *Router) initRedirect() (err error) {
	r.redirectServer = &http.Server{
		Addr:           ":80",
		ReadTimeout:    1 * time.Minute,
		WriteTimeout:   1 * time.Minute,
		IdleTimeout:    1 * time.Minute,
		MaxHeaderBytes: 8192,
		Handler: http.HandlerFunc(func(
			w http.ResponseWriter, req *http.Request) {

			if strings.HasPrefix(req.URL.Path, acme.AcmePath) {
				token := acme.ParsePath(req.URL.Path)
				if token != "" {
					chal, err := acme.GetChallenge(token)
					if err != nil {
						utils.WriteStatus(w, 400)
					} else {
						logrus.WithFields(logrus.Fields{
							"token": token,
						}).Info("router: Acme challenge requested")
						io.WriteString(w, chal.Resource)
					}
					return
				}
			} else if req.URL.Path == "/check" {
				utils.WriteText(w, 200, "ok")
				return
			}

			req.URL.Host = req.Host
			req.URL.Scheme = "https"

			http.Redirect(w, req, req.URL.String(),
				http.StatusMovedPermanently)
		}),
	}

	return
}

func (r *Router) startRedirect() {
	defer r.waiter.Done()

	if r.port == 80 {
		return
	}

	logrus.WithFields(logrus.Fields{
		"production": constants.Production,
		"protocol":   "http",
		"port":       80,
	}).Info("router: Starting redirect server")

	err := r.redirectServer.ListenAndServe()
	if err != nil {
		if err == http.ErrServerClosed {
			err = nil
		} else {
			err = &errortypes.UnknownError{
				errors.Wrap(err, "router: Server listen failed"),
			}
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("router: Redirect server error")
		}
	}
}

func (r *Router) initWeb() (err error) {
	r.typ = node.Self.Type
	r.adminDomain = node.Self.AdminDomain
	r.userDomain = node.Self.UserDomain
	r.certificates = node.Self.CertificateObjs

	r.port = node.Self.Port
	if r.port == 0 {
		r.port = 443
	}

	r.protocol = node.Self.Protocol
	if r.protocol == "" {
		r.protocol = "https"
	}

	if strings.Contains(r.typ, node.Admin) {
		r.aRouter = gin.New()

		if !constants.Production {
			r.aRouter.Use(gin.Logger())
		}

		ahandlers.Register(r.aRouter)
	}

	if strings.Contains(r.typ, node.User) {
		r.uRouter = gin.New()

		if !constants.Production {
			r.uRouter.Use(gin.Logger())
		}

		uhandlers.Register(r.uRouter)
	}

	readTimeout := time.Duration(settings.Router.ReadTimeout) * time.Second
	readHeaderTimeout := time.Duration(
		settings.Router.ReadHeaderTimeout) * time.Second
	writeTimeout := time.Duration(settings.Router.WriteTimeout) * time.Second
	idleTimeout := time.Duration(settings.Router.IdleTimeout) * time.Second

	r.webServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", r.port),
		Handler:           r,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    4096,
	}

	if r.protocol != "http" &&
		(r.certificates == nil || len(r.certificates) == 0) {

		_, _, err = certificate.SelfCert()
		if err != nil {
			return
		}
	}

	return
}

func (r *Router) startWeb() {
	defer r.waiter.Done()

	logrus.WithFields(logrus.Fields{
		"production":          constants.Production,
		"protocol":            r.protocol,
		"port":                r.port,
		"read_timeout":        settings.Router.ReadTimeout,
		"write_timeout":       settings.Router.WriteTimeout,
		"idle_timeout":        settings.Router.IdleTimeout,
		"read_header_timeout": settings.Router.ReadHeaderTimeout,
	}).Info("router: Starting web server")

	if r.protocol == "http" {
		err := r.webServer.ListenAndServe()
		if err != nil {
			if err == http.ErrServerClosed {
				err = nil
			} else {
				err = &errortypes.UnknownError{
					errors.Wrap(err, "router: Server listen failed"),
				}
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("router: Web server error")
				return
			}
		}
	} else {
		tlsConfig := &tls.Config{}
		tlsConfig.Certificates = []tls.Certificate{}

		if r.certificates != nil {
			for _, cert := range r.certificates {
				keypair, err := tls.X509KeyPair(
					[]byte(cert.Certificate),
					[]byte(cert.Key),
				)
				if err != nil {
					err = &errortypes.ReadError{
						errors.Wrap(
							err,
							"router: Failed to load certificate",
						),
					}
					logrus.WithFields(logrus.Fields{
						"error": err,
					}).Error("router: Web server certificate error")
					return
				}

				tlsConfig.Certificates = append(
					tlsConfig.Certificates,
					keypair,
				)
			}
		}

		if len(tlsConfig.Certificates) == 0 {
			certPem, keyPem, err := certificate.SelfCert()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("router: Web server self certificate error")
				return
			}

			keypair, err := tls.X509KeyPair(certPem, keyPem)
			if err != nil {
				err = &errortypes.ReadError{
					errors.Wrap(
						err,
						"router: Failed to load self certificate",
					),
				}
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("router: Web server self certificate error")
				return
			}

			tlsConfig.Certificates = append(
				tlsConfig.Certificates,
				keypair,
			)
		}

		tlsConfig.BuildNameToCertificate()

		r.webServer.TLSConfig = tlsConfig

		listener, err := tls.Listen("tcp", r.webServer.Addr, tlsConfig)
		if err != nil {
			err = &errortypes.UnknownError{
				errors.Wrap(err, "router: TLS listen failed"),
			}
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("router: Web server TLS error")
			return
		}

		err = r.webServer.Serve(listener)
		if err != nil {
			if err == http.ErrServerClosed {
				err = nil
			} else {
				err = &errortypes.UnknownError{
					errors.Wrap(err, "router: Server listen failed"),
				}
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("router: Web server error")
				return
			}
		}
	}

	return
}

func (r *Router) initServers() (err error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	err = r.initRedirect()
	if err != nil {
		return
	}

	err = r.initWeb()
	if err != nil {
		return
	}

	return
}

func (r *Router) startServers() {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.redirectServer == nil || r.webServer == nil {
		return
	}

	r.waiter.Add(2)
	go r.startRedirect()
	go r.startWeb()

	time.Sleep(250 * time.Millisecond)

	return
}

func (r *Router) Restart() {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.redirectServer != nil {
		redirectCtx, redirectCancel := context.WithTimeout(
			context.Background(),
			1*time.Second,
		)
		defer redirectCancel()
		r.redirectServer.Shutdown(redirectCtx)
	}
	if r.webServer != nil {
		webCtx, webCancel := context.WithTimeout(
			context.Background(),
			1*time.Second,
		)
		defer webCancel()
		r.webServer.Shutdown(webCtx)
	}

	func() {
		defer func() {
			recover()
		}()
		if r.redirectServer != nil {
			r.redirectServer.Close()
		}
		if r.webServer != nil {
			r.webServer.Close()
		}
	}()

	event.WebSocketsStop()

	r.redirectServer = nil
	r.webServer = nil

	time.Sleep(250 * time.Millisecond)
}

func (r *Router) Shutdown() {
	r.stop = true
	r.Restart()
	time.Sleep(1 * time.Second)
	r.Restart()
	time.Sleep(1 * time.Second)
	r.Restart()
}

func (r *Router) hashNode() []byte {
	hash := md5.New()
	io.WriteString(hash, node.Self.Type)
	io.WriteString(hash, node.Self.AdminDomain)
	io.WriteString(hash, node.Self.UserDomain)
	io.WriteString(hash, strconv.Itoa(node.Self.Port))
	io.WriteString(hash, node.Self.Protocol)

	io.WriteString(hash, strconv.Itoa(settings.Router.ReadTimeout))
	io.WriteString(hash, strconv.Itoa(settings.Router.ReadHeaderTimeout))
	io.WriteString(hash, strconv.Itoa(settings.Router.WriteTimeout))
	io.WriteString(hash, strconv.Itoa(settings.Router.IdleTimeout))

	certs := node.Self.CertificateObjs
	if certs != nil {
		for _, cert := range certs {
			io.WriteString(hash, cert.Hash())
		}
	}

	return hash.Sum(nil)
}

func (r *Router) watchNode() {
	for {
		time.Sleep(1 * time.Second)

		hash := r.hashNode()
		if bytes.Compare(r.nodeHash, hash) != 0 {
			r.nodeHash = hash
			time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
			r.Restart()
			time.Sleep(2 * time.Second)
		}
	}

	return
}

func (r *Router) Run() (err error) {
	r.nodeHash = r.hashNode()
	go r.watchNode()

	for {
		if !strings.Contains(node.Self.Type, node.Admin) &&
			!strings.Contains(node.Self.Type, node.User) {

			time.Sleep(500 * time.Millisecond)
			continue
		}

		err = r.initServers()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("router: Failed to init web servers")
			time.Sleep(1 * time.Second)
			continue
		}

		r.waiter = sync.WaitGroup{}
		r.startServers()
		r.waiter.Wait()

		if r.stop {
			break
		}
	}

	return
}

func (r *Router) Init() {
	if constants.Production {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
}
