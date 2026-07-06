package dnss

import (
	"context"
	"net"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/forward"
	"github.com/coredns/coredns/plugin/pkg/proxy"
	"github.com/coredns/coredns/plugin/pkg/transport"
	"github.com/miekg/dns"
)

type Plugin struct {
	next      atomic.Pointer[forward.Forward]
	upstreams []string
	lock      sync.Mutex
}

func (p *Plugin) setNext(next *forward.Forward) {
	p.next.Store(next)
}

func (p *Plugin) getNext() *forward.Forward {
	return p.next.Load()
}

func (p *Plugin) UpdateUpstream(dnsServers []string) {
	upstreams := []string{}
	for _, addr := range dnsServers {
		if addr == "" {
			continue
		}
		upstreams = append(upstreams, net.JoinHostPort(addr, "53"))
	}

	if len(upstreams) == 0 {
		upstreams = []string{DefaultDnsServer}
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	if slices.Equal(p.upstreams, upstreams) {
		return
	}

	fwd := forward.New()
	for _, upstream := range upstreams {
		prxy := proxy.NewProxy(upstream, upstream, transport.DNS)
		prxy.SetReadTimeout(2 * time.Second)
		prxy.Start(60 * time.Second)
		fwd.SetProxy(prxy)
	}

	oldFwd := p.getNext()
	p.setNext(fwd)

	if oldFwd != nil {
		oldFwd.OnShutdown()
	}
	p.upstreams = upstreams
}

func (p *Plugin) Shutdown() {
	p.lock.Lock()
	defer p.lock.Unlock()

	fwd := p.getNext()
	if fwd != nil {
		fwd.OnShutdown()
	}

	p.upstreams = []string{}
	p.setNext(nil)
}

func (p *Plugin) ServeDNS(ctx context.Context,
	w dns.ResponseWriter, r *dns.Msg) (int, error) {

	next := p.getNext()
	if next == nil {
		return dns.RcodeServerFailure, nil
	}

	if len(r.Question) == 0 {
		return plugin.NextOrFailure(p.Name(), next, ctx, w, r)
	}

	q := r.Question[0]
	name := q.Name
	qtype := q.Qtype
	db := database.Load()
	found := false
	var answers []dns.RR

	targetCname, okCname := db.CNAME[name]
	ipsA, okA := db.A[name]
	ipsAAAA, okAAAA := db.AAAA[name]
	internalDomain := false
	if okCname || okA || okAAAA {
		internalDomain = true
	}

	if okCname {
		answers = append(answers, &dns.CNAME{
			Hdr: dns.RR_Header{
				Name:   name,
				Rrtype: dns.TypeCNAME,
				Class:  dns.ClassINET,
				Ttl:    Ttl,
			},
			Target: targetCname,
		})
		found = true

		if qtype == dns.TypeA || qtype == dns.TypeAAAA {
			switch qtype {
			case dns.TypeA:
				if ips, ok := db.A[targetCname]; ok {
					for _, ip := range ips {
						answers = append(answers, &dns.A{
							Hdr: dns.RR_Header{
								Name:   targetCname,
								Rrtype: dns.TypeA,
								Class:  dns.ClassINET,
								Ttl:    Ttl,
							},
							A: ip,
						})
					}
				} else {
					targetQuery := new(dns.Msg)
					targetQuery.SetQuestion(targetCname, dns.TypeA)

					rw := &Response{
						ResponseWriter: w,
					}
					code, err := next.ServeDNS(ctx, rw, targetQuery)

					if err == nil && rw.msg != nil {
						answers = append(answers, rw.msg.Answer...)
					}

					m := new(dns.Msg)
					m.SetReply(r)
					m.Authoritative = false
					m.RecursionAvailable = true
					m.Answer = answers
					w.WriteMsg(m)
					return code, err
				}
			case dns.TypeAAAA:
				if ips, ok := db.AAAA[targetCname]; ok {
					for _, ip := range ips {
						answers = append(answers, &dns.AAAA{
							Hdr: dns.RR_Header{
								Name:   targetCname,
								Rrtype: dns.TypeAAAA,
								Class:  dns.ClassINET,
								Ttl:    Ttl,
							},
							AAAA: ip,
						})
					}
				} else {
					targetQuery := new(dns.Msg)
					targetQuery.SetQuestion(targetCname, dns.TypeAAAA)

					rw := &Response{
						ResponseWriter: w,
					}
					code, err := next.ServeDNS(ctx, rw, targetQuery)

					if err == nil && rw.msg != nil {
						answers = append(answers, rw.msg.Answer...)
					}

					msg := new(dns.Msg)
					msg.SetReply(r)
					msg.Authoritative = false
					msg.RecursionAvailable = true
					msg.Answer = answers
					w.WriteMsg(msg)
					return code, err
				}
			}
		}

		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true
		msg.RecursionAvailable = true
		msg.Answer = answers
		w.WriteMsg(msg)
		return dns.RcodeSuccess, nil
	}

	switch qtype {
	case dns.TypeA:
		if okA {
			for _, ip := range ipsA {
				answers = append(answers, &dns.A{
					Hdr: dns.RR_Header{
						Name:   name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    Ttl,
					},
					A: ip,
				})
			}
			found = true
		}
	case dns.TypeAAAA:
		if okAAAA {
			for _, ip := range ipsAAAA {
				answers = append(answers, &dns.AAAA{
					Hdr: dns.RR_Header{
						Name:   name,
						Rrtype: dns.TypeAAAA,
						Class:  dns.ClassINET,
						Ttl:    Ttl,
					},
					AAAA: ip,
				})
			}
			found = true
		}
	}

	if found {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true
		msg.RecursionAvailable = true
		msg.Answer = answers
		w.WriteMsg(msg)
		return dns.RcodeSuccess, nil
	} else if internalDomain {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true
		msg.RecursionAvailable = true
		w.WriteMsg(msg)
		return dns.RcodeSuccess, nil
	}

	return plugin.NextOrFailure(p.Name(), next, ctx, w, r)
}

func (p *Plugin) Name() string {
	return "pritunl-cloud"
}

func (p *Plugin) Init() {
	p.UpdateUpstream(nil)
}
