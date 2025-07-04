package proxy

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/gorilla/websocket"
	"github.com/pritunl/pritunl-cloud/balancer"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/logger"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Key                string
	Index              int
	State              int
	Domain             *Domain
	CheckUrl           string
	LastState          time.Time
	LastOnlineState    time.Time
	BackendHost        string
	BackendProto       string
	BackendProtoWs     string
	RequestHost        string
	ForwardedProto     string
	ForwardedPort      string
	TlsConfig          *tls.Config
	Dialer             *StaticDialer
	WebSockets         bool
	WebSocketsUpgrader *websocket.Upgrader
	ErrorHandler       ErrorHandler
	*httputil.ReverseProxy
}

func (h *Handler) ServeWS(rw http.ResponseWriter, r *http.Request) {
	header := utils.CloneHeader(r.Header)
	u := &url.URL{}
	*u = *r.URL

	u.Scheme = h.BackendProtoWs
	u.Host = h.BackendHost

	if h.RequestHost != "" {
		r.Host = h.RequestHost
	}

	header.Set("X-Forwarded-For",
		node.Self.GetRemoteAddr(r))
	header.Set("X-Forwarded-Host", r.Host)
	header.Set("X-Forwarded-Proto", h.ForwardedProto)
	header.Set("X-Forwarded-Port", h.ForwardedPort)

	header.Del("Upgrade")
	header.Del("Connection")
	header.Del("Sec-Websocket-Key")
	header.Del("Sec-Websocket-Version")
	header.Del("Sec-Websocket-Extensions")

	var backConn *websocket.Conn
	var backResp *http.Response
	var err error

	dialer := &websocket.Dialer{
		NetDialContext: h.Dialer.DialContext,
		Proxy: func(req *http.Request) (url *url.URL, err error) {
			if h.RequestHost != "" {
				req.Host = h.RequestHost
			} else {
				req.Host = r.Host
			}
			return
		},
		HandshakeTimeout: 45 * time.Second,
		TLSClientConfig:  h.TlsConfig,
	}

	backConn, backResp, err = dialer.Dial(u.String(), header)
	if err != nil {
		if backResp != nil {
			err = &errortypes.RequestError{
				errors.Wrapf(err, "proxy: WebSocket dial error %d",
					backResp.StatusCode),
			}
		} else {
			err = &errortypes.RequestError{
				errors.Wrap(err, "proxy: WebSocket dial error"),
			}
		}

		h.ErrorHandler(h, rw, r, err)
		return
	}
	defer backConn.Close()

	upgradeHeaders := http.Header{}
	val := backResp.Header.Get("Sec-Websocket-Protocol")
	if val != "" {
		upgradeHeaders.Set("Sec-Websocket-Protocol", val)
	}
	val = backResp.Header.Get("Set-Cookie")
	if val != "" {
		upgradeHeaders.Set("Set-Cookie", val)
	}

	frontConn, err := h.WebSocketsUpgrader.Upgrade(rw, r, upgradeHeaders)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "proxy: WebSocket upgrade error"),
		}

		h.ErrorHandler(h, rw, r, err)
		return
	}
	defer frontConn.Close()

	conn := &webSocketConn{
		front: frontConn,
		back:  backConn,
		r:     r,
	}

	conn.Run(h.Domain)
}

func (h *Handler) Serve(rw http.ResponseWriter, r *http.Request) {
	if h.WebSockets && strings.ToLower(
		r.Header.Get("Upgrade")) == "websocket" {

		h.ServeWS(rw, r)
	} else {
		h.ServeHTTP(rw, r)
	}
}

func NewHandler(index, state int, proxyProto string, proxyPort int,
	domain *Domain, backend *balancer.Backend, respHandler RespHandler,
	errHandler ErrorHandler) (hand *Handler) {

	proxyPortStr := strconv.Itoa(proxyPort)
	reqHost := domain.Domain.Host
	backendProto := backend.Protocol
	backendHost := utils.FormatHostPort(backend.Hostname, backend.Port)

	backendProtoWs := ""
	if backendProto == "https" {
		backendProtoWs = "wss"
	} else {
		backendProtoWs = "ws"
	}

	handUrl := fmt.Sprintf(
		"%s://%s:%d",
		backend.Protocol,
		backend.Hostname,
		backend.Port,
	)

	checkUrl, err := url.Parse(handUrl)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"balancer":   domain.Balancer.Name,
			"domain":     domain.Domain.Domain,
			"protocol":   backend.Protocol,
			"hostname":   backend.Hostname,
			"port":       backend.Port,
			"check_path": domain.Balancer.CheckPath,
		}).Error("proxy: Error parsing balancer backend URL")

		checkUrl, _ = url.Parse("http://0.0.0.0")
	}
	checkUrl.Path = domain.Balancer.CheckPath

	dialTimeout := time.Duration(
		settings.Router.DialTimeout) * time.Second
	dialKeepAlive := time.Duration(
		settings.Router.DialKeepAlive) * time.Second
	maxIdleConns := settings.Router.MaxIdleConns
	maxIdleConnsPerHost := settings.Router.MaxIdleConnsPerHost
	idleConnTimeout := time.Duration(
		settings.Router.IdleConnTimeout) * time.Second
	handshakeTimeout := time.Duration(
		settings.Router.HandshakeTimeout) * time.Second
	continueTimeout := time.Duration(
		settings.Router.ContinueTimeout) * time.Second

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
	}
	if domain.SkipVerify || net.ParseIP(backend.Hostname) != nil {
		tlsConfig.InsecureSkipVerify = true
	}

	if domain.ClientCertificate != nil {
		tlsConfig.Certificates = []tls.Certificate{
			*domain.ClientCertificate,
		}
	}

	writer := &logger.ErrorWriter{
		Message: "proxy: Balancer server error",
		Fields: logrus.Fields{
			"balancer": domain.Balancer.Name,
			"domain":   domain.Domain.Domain,
			"server":   handUrl,
		},
		Filters: []string{
			"context canceled",
		},
	}

	dialer := NewStaticDialer(&net.Dialer{
		Timeout:   dialTimeout,
		KeepAlive: dialKeepAlive,
		DualStack: true,
	})

	checkClient := &http.Client{
		Transport: &http.Transport{
			DialContext:         dialer.DialContext,
			DisableKeepAlives:   true,
			TLSHandshakeTimeout: 5 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS12,
				MaxVersion:         tls.VersionTLS13,
			},
		},
		Timeout: 5 * time.Second,
	}

	hand = &Handler{
		Key:            fmt.Sprintf("%s:%d", backend.Hostname, backend.Port),
		Index:          index,
		State:          state,
		Domain:         domain,
		CheckUrl:       checkUrl.String(),
		BackendHost:    backendHost,
		BackendProto:   backendProto,
		BackendProtoWs: backendProtoWs,
		RequestHost:    reqHost,
		ForwardedProto: proxyProto,
		ForwardedPort:  proxyPortStr,
		WebSockets:     domain.Balancer.WebSockets,
		TlsConfig:      tlsConfig,
		Dialer:         dialer,
		ErrorHandler:   errHandler,
		ReverseProxy: &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.Header.Set("X-Forwarded-Host", req.Host)
				req.Header.Set("X-Forwarded-Proto", proxyProto)
				req.Header.Set("X-Forwarded-Port", proxyPortStr)

				if reqHost != "" {
					req.Host = reqHost
				}

				req.URL.Scheme = backendProto
				req.URL.Host = backendHost
			},
			Transport: &TransportFix{
				transport: &http.Transport{
					Proxy:                 http.ProxyFromEnvironment,
					DialContext:           dialer.DialContext,
					MaxIdleConns:          maxIdleConns,
					MaxIdleConnsPerHost:   maxIdleConnsPerHost,
					IdleConnTimeout:       idleConnTimeout,
					TLSHandshakeTimeout:   handshakeTimeout,
					ExpectContinueTimeout: continueTimeout,
					TLSClientConfig:       tlsConfig,
				},
			},
			ErrorLog: log.New(writer, "", 0),
			ModifyResponse: func(resp *http.Response) error {
				return respHandler(hand, resp)
			},
			ErrorHandler: func(rw http.ResponseWriter,
				r *http.Request, err error) {

				errHandler(hand, rw, r, err)
			},
		},
	}

	if hand.WebSockets {
		hand.WebSocketsUpgrader = &websocket.Upgrader{
			HandshakeTimeout: time.Duration(
				settings.Router.HandshakeTimeout) * time.Second,
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
	}

	return
}
