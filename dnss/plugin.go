package dnss

import (
	"context"

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
	db := database.Load()
	found := false
	var answers []dns.RR

	target, ok := db.CNAME[name]
	if ok {
		answers = append(answers, &dns.CNAME{
			Hdr: dns.RR_Header{
				Name:   name,
				Rrtype: dns.TypeCNAME,
				Class:  dns.ClassINET,
				Ttl:    Ttl,
			},
			Target: target,
		})
		found = true

		if qtype == dns.TypeA || qtype == dns.TypeAAAA {
			switch qtype {
			case dns.TypeA:
				if ips, ok := db.A[target]; ok {
					for _, ip := range ips {
						answers = append(answers, &dns.A{
							Hdr: dns.RR_Header{
								Name:   target,
								Rrtype: dns.TypeA,
								Class:  dns.ClassINET,
								Ttl:    Ttl,
							},
							A: ip,
						})
					}
				} else {
					targetQuery := new(dns.Msg)
					targetQuery.SetQuestion(target, dns.TypeA)

					rw := &Response{
						ResponseWriter: w,
					}
					code, err := p.Next.ServeDNS(ctx, rw, targetQuery)

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
				if ips, ok := db.AAAA[target]; ok {
					for _, ip := range ips {
						answers = append(answers, &dns.AAAA{
							Hdr: dns.RR_Header{
								Name:   target,
								Rrtype: dns.TypeAAAA,
								Class:  dns.ClassINET,
								Ttl:    Ttl,
							},
							AAAA: ip,
						})
					}
				} else {
					targetQuery := new(dns.Msg)
					targetQuery.SetQuestion(target, dns.TypeAAAA)

					rw := &Response{
						ResponseWriter: w,
					}
					code, err := p.Next.ServeDNS(ctx, rw, targetQuery)

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
		ips, ok := db.A[name]
		if ok {
			for _, ip := range ips {
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
		ips, ok := db.AAAA[name]
		if ok {
			for _, ip := range ips {
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
	}

	return plugin.NextOrFailure(p.Name(), p.Next, ctx, w, r)
}

func (p *Plugin) Name() string {
	return "pritunl-cloud"
}
