package proxy

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pritunl/pritunl-cloud/balancer"
	"github.com/pritunl/pritunl-cloud/logger"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Handler struct {
	Index           int
	State           int
	LastState       time.Time
	LastOnlineState time.Time
	*httputil.ReverseProxy
}

func NewHandler(index, state int, proxyProto string, proxyPort int,
	domain *Domain, backend *balancer.Backend, respHandler RespHandler,
	errHandler ErrorHandler) (hand *Handler) {

	proxyPortStr := strconv.Itoa(proxyPort)
	reqHost := domain.Domain.Host
	backendProto := backend.Protocol
	backendHost := utils.FormatHostPort(backend.Hostname, backend.Port)

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
			"server": fmt.Sprintf(
				"%s://%s:%d",
				backend.Protocol,
				backend.Hostname,
				backend.Port,
			),
		},
		Filters: []string{
			"context canceled",
		},
	}

	hand = &Handler{
		Index: index,
		State: state,
		ReverseProxy: &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.Header.Set("X-Forwarded-For",
					node.Self.GetRemoteAddr(req))
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
					Proxy: http.ProxyFromEnvironment,
					DialContext: (&net.Dialer{
						Timeout:   dialTimeout,
						KeepAlive: dialKeepAlive,
						DualStack: true,
					}).DialContext,
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

	return
}
