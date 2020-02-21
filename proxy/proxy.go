package proxy

import (
	"bytes"
	"net/http"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/balancer"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Proxy struct {
	Domains map[string]*Domain
	lock    sync.Mutex
}

type balancerState struct {
	Balancer *balancer.Balancer
	State    *balancer.State
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

func (p *Proxy) Update(db *database.Database, balncs []*balancer.Balancer) (
	err error) {

	domains := map[string]*Domain{}
	states := []*balancerState{}

	proxyProto := node.Self.Protocol
	proxyPort := node.Self.Port

	p.lock.Lock()
	for _, balnc := range balncs {
		if !balnc.State {
			continue
		}

		onlineWeb := set.NewSet()
		unknownHighWeb := set.NewSet()
		unknownMidWeb := set.NewSet()
		unknownLowWeb := set.NewSet()
		offlineWeb := set.NewSet()

		state := &balancer.State{
			Timestamp:   time.Now(),
			Online:      []string{},
			UnknownHigh: []string{},
			UnknownMid:  []string{},
			UnknownLow:  []string{},
			Offline:     []string{},
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
				state.Requests += curDomain.RequestsTotal
				state.Retries += curDomain.RetriesTotal

				curDomain.Lock.Lock()
				for _, hand := range curDomain.OnlineWebFirst {
					onlineWeb.Add(hand.Key)
				}
				for _, hand := range curDomain.UnknownHighWebFirst {
					unknownHighWeb.Add(hand.Key)
				}
				for _, hand := range curDomain.UnknownMidWebFirst {
					unknownMidWeb.Add(hand.Key)
				}
				for _, hand := range curDomain.UnknownLowWebFirst {
					unknownLowWeb.Add(hand.Key)
				}
				for _, hand := range curDomain.OfflineWebFirst {
					offlineWeb.Add(hand.Key)
				}
				curDomain.Lock.Unlock()

				if bytes.Equal(curDomain.Hash, proxyDomain.Hash) {
					domains[domain.Domain] = curDomain
					continue
				} else {
					proxyDomain.Requests = curDomain.Requests
					proxyDomain.RequestsPrev = curDomain.RequestsPrev
					proxyDomain.RequestsTotal = curDomain.RequestsTotal
					proxyDomain.Retries = curDomain.Retries
					proxyDomain.RetriesPrev = curDomain.RetriesPrev
					proxyDomain.RetriesTotal = curDomain.RetriesTotal
				}
			}

			proxyDomain.Init()

			domains[domain.Domain] = proxyDomain
		}

		recorded := set.NewSet()
		for keyInf := range offlineWeb.Iter() {
			if recorded.Contains(keyInf) {
				continue
			}
			recorded.Add(keyInf)

			state.Offline = append(state.Offline, keyInf.(string))
		}
		for keyInf := range unknownLowWeb.Iter() {
			if recorded.Contains(keyInf) {
				continue
			}
			recorded.Add(keyInf)

			state.UnknownLow = append(state.UnknownLow, keyInf.(string))
		}
		for keyInf := range unknownMidWeb.Iter() {
			if recorded.Contains(keyInf) {
				continue
			}
			recorded.Add(keyInf)

			state.UnknownMid = append(state.UnknownMid, keyInf.(string))
		}
		for keyInf := range unknownHighWeb.Iter() {
			if recorded.Contains(keyInf) {
				continue
			}
			recorded.Add(keyInf)

			state.UnknownHigh = append(state.UnknownHigh, keyInf.(string))
		}
		for keyInf := range onlineWeb.Iter() {
			if recorded.Contains(keyInf) {
				continue
			}
			recorded.Add(keyInf)

			state.Online = append(state.Online, keyInf.(string))
		}

		states = append(states, &balancerState{
			Balancer: balnc,
			State:    state,
		})
	}

	p.Domains = domains
	p.lock.Unlock()

	for _, balncState := range states {
		err = balncState.Balancer.CommitState(db, balncState.State)
		if err != nil {
			return
		}
	}

	return
}

func (p *Proxy) syncCount() {
	p.lock.Lock()
	defer p.lock.Unlock()

	domains := p.Domains
	for _, dom := range domains {
		dom.Print()

		reqPrev := dom.RequestsPrev
		dom.RequestsTotal = reqPrev[0] + reqPrev[1] + reqPrev[2] +
			reqPrev[3] + reqPrev[4] + dom.Requests
		reqPrev[0] = reqPrev[1]
		reqPrev[1] = reqPrev[2]
		reqPrev[2] = reqPrev[3]
		reqPrev[3] = reqPrev[4]
		reqPrev[4] = dom.Requests
		dom.RequestsPrev = reqPrev
		dom.Requests = 0

		retPrev := dom.RetriesPrev
		dom.RetriesTotal = retPrev[0] + retPrev[1] + retPrev[2] +
			retPrev[3] + retPrev[4] + dom.Retries
		retPrev[0] = retPrev[1]
		retPrev[1] = retPrev[2]
		retPrev[2] = retPrev[3]
		retPrev[3] = retPrev[4]
		retPrev[4] = dom.Retries
		dom.RetriesPrev = retPrev
		dom.Retries = 0
	}
}

func (p *Proxy) runCounter() {
	for {
		time.Sleep(10 * time.Second)
		p.syncCount()
	}
}

func (p *Proxy) Init() {
	p.Domains = map[string]*Domain{}
	go p.runCounter()
}
