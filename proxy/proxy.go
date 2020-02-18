package proxy

import (
	"bytes"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/pritunl/pritunl-cloud/balancer"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Proxy struct {
	Domains map[string]*Domain
}

func (p *Proxy) ServeHTTP(hst string, rw http.ResponseWriter,
	r *http.Request) {

	domain := p.Domains[hst]
	if domain == nil {
		utils.WriteStatus(rw, 404)
		return
	}

	domain.ServeHTTPFirst(rw, r)
}

func (p *Proxy) Update(balncs []*balancer.Balancer) (err error) {
	domains := map[string]*Domain{}

	proxyProto := node.Self.Protocol
	proxyPort := node.Self.Port

	for _, balnc := range balncs {
		if !balnc.State {
			continue
		}

		for _, domain := range balnc.Domains {
			if domains[domain.Domain] != nil {
				conflictDomain := domains[domain.Domain]
				logrus.WithFields(logrus.Fields{
					"first_balancer_id":    conflictDomain.Balancer.Id.Hex(),
					"first_balancer_name":  conflictDomain.Balancer.Name,
					"second_balancer_id":   balnc.Id.Hex(),
					"second_balancer_name": balnc.Name,
					"conflict_domain":      domain.Domain,
				}).Error("proxy: Balancer domain conflict")
				continue
			}

			proxyDomain := &Domain{
				SkipVerify: settings.Router.SkipVerify,
				ProxyProto: proxyProto,
				ProxyPort:  proxyPort,
				Balancer:   balnc,
				Domain:     domain,
			}
			proxyDomain.CalculateHash()

			curDomain := p.Domains[domain.Domain]
			if curDomain != nil && curDomain.Balancer.Id == balnc.Id {
				curDomain.Print()

				if bytes.Equal(curDomain.Hash, proxyDomain.Hash) {
					domains[domain.Domain] = curDomain
					continue
				} else {
					proxyDomain.Requests = curDomain.Requests
					proxyDomain.RequestsPrev = curDomain.RequestsPrev
					proxyDomain.Retries = curDomain.Retries
					proxyDomain.RetriesPrev = curDomain.RetriesPrev
				}
			}

			proxyDomain.Init()

			domains[domain.Domain] = proxyDomain
		}
	}

	p.Domains = domains

	return
}
