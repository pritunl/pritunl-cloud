package router

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/acme"
	"github.com/pritunl/pritunl-cloud/ahandlers"
	"github.com/pritunl/pritunl-cloud/balancer"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/crypto"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/proxy"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/uhandlers"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/tools/commander"
	"github.com/sirupsen/logrus"
)

var (
	client = &http.Client{
		Timeout: 10 * time.Second,
	}
)

type Router struct {
	nodeHash         []byte
	singleType       bool
	adminType        bool
	userType         bool
	balancerType     bool
	port             int
	noRedirectServer bool
	redirectSystemd  bool
	protocol         string
	adminDomain      string
	userDomain       string
	stateLock        sync.Mutex
	balancers        []*balancer.Balancer
	certificates     *Certificates
	box              *crypto.AsymNaclHmac
	aRouter          *gin.Engine
	uRouter          *gin.Engine
	waiter           sync.WaitGroup
	lock             sync.Mutex
	redirectServer   *http.Server
	redirectContext  context.Context
	redirectCancel   context.CancelFunc
	webServer        *http.Server
	proxy            *proxy.Proxy
	stop             bool
}

func (r *Router) ServeHTTP(w http.ResponseWriter, re *http.Request) {
	if node.Self.ForwardedProtoHeader != "" &&
		strings.ToLower(re.Header.Get(
			node.Self.ForwardedProtoHeader)) == "http" {

		re.URL.Host = utils.StripPort(re.Host)
		re.URL.Scheme = "https"

		http.Redirect(w, re, re.URL.String(),
			http.StatusMovedPermanently)
		return
	}

	if r.singleType {
		if r.adminType {
			r.aRouter.ServeHTTP(w, re)
		} else if r.userType {
			r.uRouter.ServeHTTP(w, re)
		} else if r.balancerType {
			r.proxy.ServeHTTP(utils.StripPort(re.Host), w, re)
		} else {
			utils.WriteStatus(w, 520)
		}
		return
	} else {
		hst := utils.StripPort(re.Host)
		if r.adminType && hst == r.adminDomain {
			r.aRouter.ServeHTTP(w, re)
			return
		} else if r.userType && hst == r.userDomain {
			r.uRouter.ServeHTTP(w, re)
			return
		} else if r.balancerType {
			r.proxy.ServeHTTP(hst, w, re)
			return
		}
	}

	if re.URL.Path == "/check" {
		utils.WriteText(w, 200, "ok")
		return
	}

	utils.WriteStatus(w, 404)
}

func (r *Router) initRedirect() (err error) {
	if r.redirectSystemd {
		libPath := settings.Hypervisor.LibPath
		err = utils.ExistsMkdir(libPath, 0755)
		if err != nil {
			return
		}

		redirectPth := path.Join(libPath, "redirect.conf")

		r.box = &crypto.AsymNaclHmac{}
		err = r.box.Generate()
		if err != nil {
			return
		}

		privKeyStr, secStr := r.box.Export()

		redirectOutput := &bytes.Buffer{}
		redirectData := &redirectConfData{
			WebPort:    r.port,
			PrivateKey: privKeyStr,
			Secret:     secStr,
		}

		err = redirectConf.Execute(redirectOutput, redirectData)
		if err != nil {
			err = &errortypes.ParseError{
				errors.Wrap(err, "router: Failed to exec redirect template"),
			}
			return
		}

		err = utils.CreateWrite(
			redirectPth,
			redirectOutput.String(),
			0600,
		)
		if err != nil {
			return
		}
	} else {
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
					token = utils.FilterStr(token, 96)
					if token != "" {
						chal, err := acme.GetChallenge(token)
						if err != nil {
							utils.WriteStatus(w, 400)
						} else {
							logrus.WithFields(logrus.Fields{
								"token": token,
							}).Info("router: Acme challenge requested")
							utils.WriteText(w, 200, chal.Resource)
						}
						return
					}
				} else if req.URL.Path == "/check" {
					utils.WriteText(w, 200, "ok")
					return
				}

				newHost := utils.StripPort(req.Host)
				if r.port != 443 {
					newHost += fmt.Sprintf(":%d", r.port)
				}

				req.URL.Host = newHost
				req.URL.Scheme = "https"

				http.Redirect(w, req, req.URL.String(),
					http.StatusMovedPermanently)
			}),
		}
	}

	return
}

func (r *Router) redirectChallengeListen(ctx context.Context) {
	db := database.GetDatabase()
	defer db.Close()

	lst, e := event.SubscribeListener(db, []string{"acme"})
	if e != nil {
		select {
		case <-ctx.Done():
			return
		default:
		}

		logrus.WithFields(logrus.Fields{
			"error": e,
		}).Error("acme: Event watch error")
		return
	}

	sub := lst.Listen()
	defer lst.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-sub:
			if !ok {
				break
			}

			go func() {
				err := r.sendChallenge(msg.Data)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"error": err,
					}).Error("router: Failed to send challenge " +
						"to redirect server")
				}
			}()
		}
	}
}

func (r *Router) startRedirect() {
	defer r.waiter.Done()

	if r.redirectSystemd {
		resp, e := commander.Exec(&commander.Opt{
			Name: "systemctl",
			Args: []string{
				"restart",
				"pritunl-cloud-redirect.service",
			},
			Timeout: 30 * time.Second,
			PipeOut: true,
			PipeErr: true,
		})
		if e != nil {
			logrus.WithFields(resp.Map()).Error(
				"router: Failed to start redirect server")
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		r.redirectContext = ctx
		r.redirectCancel = cancel

		for {
			r.redirectChallengeListen(ctx)

			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	} else {
		if r.port == 80 || r.noRedirectServer {
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
}

func (r *Router) sendChallenge(chal any) (err error) {
	encData, err := r.box.SealJson(chal)
	if err != nil {
		return
	}

	req, err := http.NewRequest(
		"POST",
		"http://127.0.0.1:80/token",
		bytes.NewReader([]byte(encData)),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "acme: Redirect token request failed"),
		}
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "acme: Redirect token request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logrus.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
		}).Error("acme: Redirect request bad status")
		return
	}

	return
}

func (r *Router) initWeb() (err error) {
	r.adminType = node.Self.IsAdmin()
	r.userType = node.Self.IsUser()
	r.balancerType = node.Self.IsBalancer()
	r.adminDomain = node.Self.AdminDomain
	r.userDomain = node.Self.UserDomain
	r.noRedirectServer = node.Self.NoRedirectServer
	r.redirectSystemd = settings.Router.RedirectServerSystemd

	if r.adminType && !r.userType && !r.balancerType {
		r.singleType = true
	} else if r.userType && !r.balancerType && !r.adminType {
		r.singleType = true
	} else if r.balancerType && !r.adminType && !r.userType {
		r.singleType = true
	} else {
		r.singleType = false
	}

	r.port = node.Self.Port
	if r.port == 0 {
		r.port = 443
	}

	r.protocol = node.Self.Protocol
	if r.protocol == "" {
		r.protocol = "https"
	}

	if r.adminType {
		r.aRouter = gin.New()

		if constants.DebugWeb {
			r.aRouter.Use(gin.Logger())
		}

		ahandlers.Register(r.aRouter)
	}

	if r.userType {
		r.uRouter = gin.New()

		if constants.DebugWeb {
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
		MaxHeaderBytes:    settings.Router.MaxHeaderBytes,
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
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
			CipherSuites: []uint16{
				tls.TLS_AES_128_GCM_SHA256,                        // 0x1301
				tls.TLS_AES_256_GCM_SHA384,                        // 0x1302
				tls.TLS_CHACHA20_POLY1305_SHA256,                  // 0x1303
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,       // 0xc02b
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,         // 0xc02f
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,       // 0xc02c
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,         // 0xc030
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256, // 0xcca9
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,   // 0xcca8
			},
			GetCertificate: r.certificates.GetCertificate,
		}

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

	err = r.certificates.Init()
	if err != nil {
		return
	}

	err = r.updateState()
	if err != nil {
		return
	}

	err = r.initWeb()
	if err != nil {
		return
	}

	err = r.initRedirect()
	if err != nil {
		return
	}

	return
}

func (r *Router) startServers() {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.webServer == nil {
		return
	}

	if !r.redirectSystemd && r.redirectServer == nil {
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
	for _, typ := range node.Self.Types {
		io.WriteString(hash, typ)
	}
	io.WriteString(hash, node.Self.AdminDomain)
	io.WriteString(hash, node.Self.UserDomain)
	io.WriteString(hash, strconv.Itoa(node.Self.Port))
	io.WriteString(hash, fmt.Sprintf("%t", node.Self.NoRedirectServer))
	io.WriteString(hash, node.Self.Protocol)

	io.WriteString(hash, strconv.Itoa(settings.Router.ReadTimeout))
	io.WriteString(hash, strconv.Itoa(settings.Router.ReadHeaderTimeout))
	io.WriteString(hash, strconv.Itoa(settings.Router.WriteTimeout))
	io.WriteString(hash, strconv.Itoa(settings.Router.IdleTimeout))
	io.WriteString(hash, strconv.Itoa(settings.Router.MaxHeaderBytes))

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
}

func (r *Router) updateState() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	if node.Self.IsBalancer() {
		dcId, e := node.Self.GetDatacenter(db)
		if e != nil {
			err = e
			return
		}

		balncs, e := balancer.GetAll(db, &bson.M{
			"datacenter": dcId,
		})
		if e != nil {
			r.balancers = []*balancer.Balancer{}
			return
		}

		r.balancers = balncs
	} else {
		r.balancers = []*balancer.Balancer{}
	}

	r.stateLock.Lock()
	defer r.stateLock.Unlock()

	err = r.certificates.Update(db, r.balancers)
	if err != nil {
		return
	}

	err = r.proxy.Update(db, r.balancers)
	if err != nil {
		return
	}

	return
}

func (r *Router) watchState() {
	for {
		time.Sleep(4 * time.Second)

		err := r.updateState()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("proxy: Failed to load proxy state")
		}
	}
}

func (r *Router) Run() (err error) {
	r.nodeHash = r.hashNode()
	go r.watchNode()
	go r.watchState()

	for {
		if !node.Self.IsAdmin() && !node.Self.IsUser() &&
			!node.Self.IsBalancer() {

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
	if constants.DebugWeb {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r.certificates = &Certificates{}
	r.proxy = &proxy.Proxy{}
	r.proxy.Init()
}
