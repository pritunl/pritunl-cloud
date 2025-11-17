package dnss

import (
	"context"
	"time"

	"github.com/coredns/coredns/plugin/forward"
	"github.com/coredns/coredns/plugin/pkg/proxy"
	"github.com/coredns/coredns/plugin/pkg/transport"
	"github.com/dropbox/godropbox/errors"
	"github.com/miekg/dns"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type Server struct {
	mux *dns.ServeMux
	udp *dns.Server
	tcp *dns.Server
}

func (s *Server) ListenUdp() (err error) {
	err = s.udp.ListenAndServe()
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "dnss: Server udp listen error"),
		}
		return
	}

	return
}

func (s *Server) ListenTcp() (err error) {
	err = s.tcp.ListenAndServe()
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "dnss: Server tcp listen error"),
		}
		return
	}

	return
}

func (s *Server) Shutdown() (err error) {
	e := s.tcp.Shutdown()
	if e != nil {
		err = e
	}

	e = s.udp.Shutdown()
	if e != nil {
		err = e
	}

	return
}

func NewServer(host string) (server *Server) {
	mux := dns.NewServeMux()

	prxy := proxy.NewProxy("google", "8.8.8.8:53", transport.DNS)
	prxy.SetReadTimeout(2 * time.Second)
	prxy.Start(60 * time.Second)

	fwd := forward.New()
	fwd.SetProxy(prxy)

	custom := &Plugin{
		Next: fwd,
	}

	mux.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		custom.ServeDNS(context.Background(), w, r)
	})

	return &Server{
		mux: mux,
		udp: &dns.Server{
			Addr:    host,
			Net:     "udp",
			Handler: mux,
		},
		tcp: &dns.Server{
			Addr:    host,
			Net:     "tcp",
			Handler: mux,
		},
	}
}
