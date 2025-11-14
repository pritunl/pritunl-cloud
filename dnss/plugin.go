package dnss

import (
	"context"
	"fmt"

	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
)

type Plugin struct {
	Next plugin.Handler
}

func (p *Plugin) ServeDNS(ctx context.Context,
	w dns.ResponseWriter, r *dns.Msg) (int, error) {

	if len(r.Question) == 0 {
		return plugin.NextOrFailure(p.Name(), p.Next, ctx, w, r)
	}

	q := r.Question[0]
	name := q.Name
	qtype := q.Qtype

	var answers []dns.RR
	var found bool

	switch qtype {
	case dns.TypeA:
		db := database.Load()
		ips, ok := db.A[name]
		if ok {
			for _, ip := range ips {
				answers = append(answers, &dns.A{
					Hdr: dns.RR_Header{
						Name:   name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    10,
					},
					A: ip,
				})
			}
			found = true
		}
	case dns.TypeAAAA:
		db := database.Load()
		ips, ok := db.AAAA[name]
		if ok {
			for _, ip := range ips {
				answers = append(answers, &dns.AAAA{
					Hdr: dns.RR_Header{
						Name:   name,
						Rrtype: dns.TypeAAAA,
						Class:  dns.ClassINET,
						Ttl:    10,
					},
					AAAA: ip,
				})
			}
			found = true
		}
	case dns.TypeCNAME:
		db := database.Load()
		target, ok := db.CNAME[name]
		if ok {
			answers = append(answers, &dns.CNAME{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeCNAME,
					Class:  dns.ClassINET,
					Ttl:    10,
				},
				Target: target,
			})
			found = true
		}
	}

	fmt.Println(answers)

	if found {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = true
		m.Answer = answers
		w.WriteMsg(m)
		return dns.RcodeSuccess, nil
	}

	return plugin.NextOrFailure(p.Name(), p.Next, ctx, w, r)
}

func (p *Plugin) Name() string {
	return "pritunl-cloud"
}
