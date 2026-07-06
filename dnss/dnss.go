package dnss

import (
	"context"

	"github.com/dropbox/godropbox/errors"
	"github.com/miekg/dns"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type Server struct {
	mux    *dns.ServeMux
	udp    *dns.Server
	tcp    *dns.Server
	plugin *Plugin
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

func (s *Server) UpdateUpstream(dnsServers, dnsServers6 []string) {
	s.plugin.UpdateUpstream(dnsServers, dnsServers6)
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

	s.plugin.Shutdown()

	return
}

func NewServer(host string) (server *Server) {
	mux := dns.NewServeMux()

	custom := &Plugin{}
	custom.Init()

	mux.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		custom.ServeDNS(context.Background(), w, r)
	})

	return &Server{
		mux:    mux,
		plugin: custom,
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
