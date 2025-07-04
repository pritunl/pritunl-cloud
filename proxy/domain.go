package proxy

import (
	"crypto/md5"
	"crypto/tls"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/authority"
	"github.com/pritunl/pritunl-cloud/balancer"
)

type Domain struct {
	Hash              []byte
	Requests          *int32
	RequestsPrev      [5]int
	RequestsTotal     int
	Retries           *int32
	RetriesPrev       [5]int
	RetriesTotal      int
	Lock              sync.Mutex
	ProxyProto        string
	ProxyPort         int
	SkipVerify        bool
	Balancer          *balancer.Balancer
	Domain            *balancer.Domain
	ClientAuthority   *authority.Authority
	ClientCertificate *tls.Certificate

	OnlineWebFirst      []*Handler
	UnknownHighWebFirst []*Handler
	UnknownMidWebFirst  []*Handler
	UnknownLowWebFirst  []*Handler
	OfflineWebFirst     []*Handler

	OnlineWebSecond      []*Handler
	UnknownHighWebSecond []*Handler
	UnknownMidWebSecond  []*Handler
	UnknownLowWebSecond  []*Handler
	OfflineWebSecond     []*Handler

	OnlineWebThird      []*Handler
	UnknownHighWebThird []*Handler
	UnknownMidWebThird  []*Handler
	UnknownLowWebThird  []*Handler
	OfflineWebThird     []*Handler

	WebSocketConns     set.Set
	WebSocketConnsLock sync.Mutex
}

func (d *Domain) CalculateHash() {
	h := md5.New()

	h.Write([]byte(d.ProxyProto))
	h.Write([]byte(strconv.Itoa(d.ProxyPort)))
	h.Write([]byte(strconv.FormatBool(d.SkipVerify)))

	h.Write([]byte(d.Balancer.Id.Hex()))
	h.Write([]byte(d.Balancer.Name))
	h.Write([]byte(d.Balancer.CheckPath))
	h.Write([]byte(strconv.FormatBool(d.Balancer.WebSockets)))
	h.Write([]byte(d.Domain.Domain))
	h.Write([]byte(d.Domain.Host))

	if !d.Balancer.ClientAuthority.IsZero() {
		h.Write([]byte(d.Balancer.ClientAuthority.Hex()))
	}
	for _, backend := range d.Balancer.Backends {
		h.Write([]byte(backend.Protocol))
		h.Write([]byte(backend.Hostname))
		h.Write([]byte(strconv.Itoa(backend.Port)))
	}

	d.Hash = h.Sum(nil)
}

func (d *Domain) Init() {
	d.Lock.Lock()
	defer d.Lock.Unlock()

	if !d.Balancer.ClientAuthority.IsZero() {
		//clientAuthr, err := authority.Get(db, d.Balancer.ClientAuthority)
		//if err != nil {
		//	if _, ok := err.(*database.NotFoundError); ok {
		//		err = nil
		//
		//		logrus.WithFields(logrus.Fields{
		//			"balancer_id":         d.Balancer.Id.Hex(),
		//			"client_authority_id": d.Balancer.ClientAuthority.Hex(),
		//		}).Warn("proxy: Service client authority not found")
		//	} else {
		//		return
		//	}
		//}
		//
		// var cert *tls.Certificate
		//if clientAuthr != nil {
		//	cert, err = clientAuthr.CreateClientCertificate(db)
		//	if err != nil {
		//		return
		//	}
		//}
	}

	unknownHighWebFirst := []*Handler{}
	unknownHighWebSecond := []*Handler{}
	unknownHighWebThird := []*Handler{}

	for i, backend := range d.Balancer.Backends {
		hand := NewHandler(i, UnknownHigh, d.ProxyProto, d.ProxyPort, d,
			backend, d.ResponseHandler, d.ErrorHandlerFirst)
		unknownHighWebFirst = append(unknownHighWebFirst, hand)

		hand = NewHandler(i, UnknownHigh, d.ProxyProto, d.ProxyPort, d,
			backend, d.ResponseHandler, d.ErrorHandlerSecond)
		unknownHighWebSecond = append(unknownHighWebSecond, hand)

		hand = NewHandler(i, UnknownHigh, d.ProxyProto, d.ProxyPort, d,
			backend, d.ResponseHandler, d.ErrorHandlerThird)
		unknownHighWebThird = append(unknownHighWebThird, hand)
	}

	d.OnlineWebFirst = []*Handler{}
	d.UnknownHighWebFirst = unknownHighWebFirst
	d.UnknownMidWebFirst = []*Handler{}
	d.UnknownLowWebFirst = []*Handler{}
	d.OfflineWebFirst = []*Handler{}

	d.OnlineWebSecond = []*Handler{}
	d.UnknownHighWebSecond = unknownHighWebSecond
	d.UnknownMidWebSecond = []*Handler{}
	d.UnknownLowWebSecond = []*Handler{}
	d.OfflineWebSecond = []*Handler{}

	d.OnlineWebThird = []*Handler{}
	d.UnknownHighWebThird = unknownHighWebThird
	d.UnknownMidWebThird = []*Handler{}
	d.UnknownLowWebThird = []*Handler{}
	d.OfflineWebThird = []*Handler{}

	d.WebSocketConns = set.NewSet()
}

func (d *Domain) ServeHTTPFirst(rw http.ResponseWriter, r *http.Request) {
	atomic.AddInt32(d.Requests, 1)

	onlineWebFirst := d.OnlineWebFirst
	l := len(onlineWebFirst)
	if l != 0 {
		onlineWebFirst[rand.Intn(l)].Serve(rw, r)
		return
	}

	unknownHighWebFirst := d.UnknownHighWebFirst
	l = len(unknownHighWebFirst)
	if l != 0 {
		unknownHighWebFirst[rand.Intn(l)].Serve(rw, r)
		return
	}

	unknownMidWebFirst := d.UnknownMidWebFirst
	l = len(unknownMidWebFirst)
	if l != 0 {
		unknownMidWebFirst[rand.Intn(l)].Serve(rw, r)
		return
	}

	unknownLowWebFirst := d.UnknownLowWebFirst
	l = len(unknownLowWebFirst)
	if l != 0 {
		unknownLowWebFirst[rand.Intn(l)].Serve(rw, r)
		return
	}

	offlineWebFirst := d.OfflineWebFirst
	l = len(offlineWebFirst)
	if l != 0 {
		offlineWebFirst[rand.Intn(l)].Serve(rw, r)
		return
	}

	rw.WriteHeader(http.StatusBadGateway)
}

func (d *Domain) ServeHTTPSecond(rw http.ResponseWriter, r *http.Request) {
	atomic.AddInt32(d.Retries, 1)

	onlineWebSecond := d.OnlineWebSecond
	l := len(onlineWebSecond)
	if l != 0 {
		onlineWebSecond[rand.Intn(l)].Serve(rw, r)
		return
	}

	unknownHighWebSecond := d.UnknownHighWebSecond
	l = len(unknownHighWebSecond)
	if l != 0 {
		unknownHighWebSecond[rand.Intn(l)].Serve(rw, r)
		return
	}

	unknownMidWebSecond := d.UnknownMidWebSecond
	l = len(unknownMidWebSecond)
	if l != 0 {
		unknownMidWebSecond[rand.Intn(l)].Serve(rw, r)
		return
	}

	unknownLowWebSecond := d.UnknownLowWebSecond
	l = len(unknownLowWebSecond)
	if l != 0 {
		unknownLowWebSecond[rand.Intn(l)].Serve(rw, r)
		return
	}

	offlineWebSecond := d.OfflineWebSecond
	l = len(offlineWebSecond)
	if l != 0 {
		offlineWebSecond[rand.Intn(l)].Serve(rw, r)
		return
	}

	rw.WriteHeader(http.StatusBadGateway)
}

func (d *Domain) ServeHTTPThird(rw http.ResponseWriter, r *http.Request) {
	atomic.AddInt32(d.Retries, 1)

	onlineWebThird := d.OnlineWebThird
	l := len(onlineWebThird)
	if l != 0 {
		onlineWebThird[rand.Intn(l)].Serve(rw, r)
		return
	}

	unknownHighWebThird := d.UnknownHighWebThird
	l = len(unknownHighWebThird)
	if l != 0 {
		unknownHighWebThird[rand.Intn(l)].Serve(rw, r)
		return
	}

	unknownMidWebThird := d.UnknownMidWebThird
	l = len(unknownMidWebThird)
	if l != 0 {
		unknownMidWebThird[rand.Intn(l)].Serve(rw, r)
		return
	}

	unknownLowWebThird := d.UnknownLowWebThird
	l = len(unknownLowWebThird)
	if l != 0 {
		unknownLowWebThird[rand.Intn(l)].Serve(rw, r)
		return
	}

	offlineWebThird := d.OfflineWebThird
	l = len(offlineWebThird)
	if l != 0 {
		offlineWebThird[rand.Intn(l)].Serve(rw, r)
		return
	}

	rw.WriteHeader(http.StatusBadGateway)
}

func (d *Domain) checkHandler(hand *Handler) {
	resp, err := hand.CheckClient.Get(hand.CheckUrl)
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if hand.State != Offline {
			d.offlineHandler(hand)
		}
	} else {
		if hand.State != Online {
			d.upgradeHandler(hand)
		}
	}
}

func (d *Domain) Check() {
	d.Lock.Lock()
	defer d.Lock.Unlock()

	for _, hand := range d.OnlineWebFirst {
		go d.checkHandler(hand)
	}

	for _, hand := range d.UnknownHighWebFirst {
		go d.checkHandler(hand)
	}

	for _, hand := range d.UnknownMidWebFirst {
		go d.checkHandler(hand)
	}

	for _, hand := range d.UnknownLowWebFirst {
		go d.checkHandler(hand)
	}

	for _, hand := range d.OfflineWebFirst {
		go d.checkHandler(hand)
	}

	return
}

func (d *Domain) upgradeHandler(hand *Handler) {
	d.Lock.Lock()
	defer d.Lock.Unlock()

	index := hand.Index
	state := hand.State

	switch state {
	case Online:
		break
	case UnknownHigh:
		if time.Since(hand.LastOnlineState) > 5*time.Second {
			hand = d.UnknownHighWebFirst[index]
			d.UnknownHighWebFirst[index] =
				d.UnknownHighWebFirst[len(d.UnknownHighWebFirst)-1]
			d.UnknownHighWebFirst[len(d.UnknownHighWebFirst)-1] = nil
			d.UnknownHighWebFirst =
				d.UnknownHighWebFirst[:len(d.UnknownHighWebFirst)-1]
			for i, h := range d.UnknownHighWebFirst {
				h.Index = i
			}
			hand.Index = len(d.OnlineWebFirst)
			hand.State = Online
			hand.LastOnlineState = time.Now()
			d.OnlineWebFirst = append(d.OnlineWebFirst, hand)

			hand = d.UnknownHighWebSecond[index]
			d.UnknownHighWebSecond[index] =
				d.UnknownHighWebSecond[len(d.UnknownHighWebSecond)-1]
			d.UnknownHighWebSecond[len(d.UnknownHighWebSecond)-1] = nil
			d.UnknownHighWebSecond =
				d.UnknownHighWebSecond[:len(d.UnknownHighWebSecond)-1]
			for i, h := range d.UnknownHighWebSecond {
				h.Index = i
			}
			hand.Index = len(d.OnlineWebSecond)
			hand.State = Online
			hand.LastOnlineState = time.Now()
			d.OnlineWebSecond = append(d.OnlineWebSecond, hand)

			hand = d.UnknownHighWebThird[index]
			d.UnknownHighWebThird[index] =
				d.UnknownHighWebThird[len(d.UnknownHighWebThird)-1]
			d.UnknownHighWebThird[len(d.UnknownHighWebThird)-1] = nil
			d.UnknownHighWebThird =
				d.UnknownHighWebThird[:len(d.UnknownHighWebThird)-1]
			for i, h := range d.UnknownHighWebThird {
				h.Index = i
			}
			hand.Index = len(d.OnlineWebThird)
			hand.State = Online
			hand.LastOnlineState = time.Now()
			d.OnlineWebThird = append(d.OnlineWebThird, hand)
		}

		break
	case UnknownMid:
		if time.Since(hand.LastOnlineState) > 5*time.Second {
			hand = d.UnknownMidWebFirst[index]
			d.UnknownMidWebFirst[index] =
				d.UnknownMidWebFirst[len(d.UnknownMidWebFirst)-1]
			d.UnknownMidWebFirst[len(d.UnknownMidWebFirst)-1] = nil
			d.UnknownMidWebFirst =
				d.UnknownMidWebFirst[:len(d.UnknownMidWebFirst)-1]
			for i, h := range d.UnknownMidWebFirst {
				h.Index = i
			}
			hand.Index = len(d.OnlineWebFirst)
			hand.State = Online
			hand.LastOnlineState = time.Now()
			d.OnlineWebFirst = append(d.OnlineWebFirst, hand)

			hand = d.UnknownMidWebSecond[index]
			d.UnknownMidWebSecond[index] =
				d.UnknownMidWebSecond[len(d.UnknownMidWebSecond)-1]
			d.UnknownMidWebSecond[len(d.UnknownMidWebSecond)-1] = nil
			d.UnknownMidWebSecond =
				d.UnknownMidWebSecond[:len(d.UnknownMidWebSecond)-1]
			for i, h := range d.UnknownMidWebSecond {
				h.Index = i
			}
			hand.Index = len(d.OnlineWebSecond)
			hand.State = Online
			hand.LastOnlineState = time.Now()
			d.OnlineWebSecond = append(d.OnlineWebSecond, hand)

			hand = d.UnknownMidWebThird[index]
			d.UnknownMidWebThird[index] =
				d.UnknownMidWebThird[len(d.UnknownMidWebThird)-1]
			d.UnknownMidWebThird[len(d.UnknownMidWebThird)-1] = nil
			d.UnknownMidWebThird =
				d.UnknownMidWebThird[:len(d.UnknownMidWebThird)-1]
			for i, h := range d.UnknownMidWebThird {
				h.Index = i
			}
			hand.Index = len(d.OnlineWebThird)
			hand.State = Online
			hand.LastOnlineState = time.Now()
			d.OnlineWebThird = append(d.OnlineWebThird, hand)
		}

		break
	case UnknownLow:
		if time.Since(hand.LastOnlineState) > 5*time.Second {
			hand = d.UnknownLowWebFirst[index]
			d.UnknownLowWebFirst[index] =
				d.UnknownLowWebFirst[len(d.UnknownLowWebFirst)-1]
			d.UnknownLowWebFirst[len(d.UnknownLowWebFirst)-1] = nil
			d.UnknownLowWebFirst =
				d.UnknownLowWebFirst[:len(d.UnknownLowWebFirst)-1]
			for i, h := range d.UnknownLowWebFirst {
				h.Index = i
			}
			hand.Index = len(d.OnlineWebFirst)
			hand.State = Online
			hand.LastOnlineState = time.Now()
			d.OnlineWebFirst = append(d.OnlineWebFirst, hand)

			hand = d.UnknownLowWebSecond[index]
			d.UnknownLowWebSecond[index] =
				d.UnknownLowWebSecond[len(d.UnknownLowWebSecond)-1]
			d.UnknownLowWebSecond[len(d.UnknownLowWebSecond)-1] = nil
			d.UnknownLowWebSecond =
				d.UnknownLowWebSecond[:len(d.UnknownLowWebSecond)-1]
			for i, h := range d.UnknownLowWebSecond {
				h.Index = i
			}
			hand.Index = len(d.OnlineWebSecond)
			hand.State = Online
			hand.LastOnlineState = time.Now()
			d.OnlineWebSecond = append(d.OnlineWebSecond, hand)

			hand = d.UnknownLowWebThird[index]
			d.UnknownLowWebThird[index] =
				d.UnknownLowWebThird[len(d.UnknownLowWebThird)-1]
			d.UnknownLowWebThird[len(d.UnknownLowWebThird)-1] = nil
			d.UnknownLowWebThird =
				d.UnknownLowWebThird[:len(d.UnknownLowWebThird)-1]
			for i, h := range d.UnknownLowWebThird {
				h.Index = i
			}
			hand.Index = len(d.OnlineWebThird)
			hand.State = Online
			hand.LastOnlineState = time.Now()
			d.OnlineWebThird = append(d.OnlineWebThird, hand)
		}

		break
	case Offline:
		if time.Since(hand.LastOnlineState) > 5*time.Second {
			hand = d.OfflineWebFirst[index]
			d.OfflineWebFirst[index] =
				d.OfflineWebFirst[len(d.OfflineWebFirst)-1]
			d.OfflineWebFirst[len(d.OfflineWebFirst)-1] = nil
			d.OfflineWebFirst =
				d.OfflineWebFirst[:len(d.OfflineWebFirst)-1]
			for i, h := range d.OfflineWebFirst {
				h.Index = i
			}
			hand.Index = len(d.OnlineWebFirst)
			hand.State = Online
			hand.LastOnlineState = time.Now()
			d.OnlineWebFirst = append(d.OnlineWebFirst, hand)

			hand = d.OfflineWebSecond[index]
			d.OfflineWebSecond[index] =
				d.OfflineWebSecond[len(d.OfflineWebSecond)-1]
			d.OfflineWebSecond[len(d.OfflineWebSecond)-1] = nil
			d.OfflineWebSecond =
				d.OfflineWebSecond[:len(d.OfflineWebSecond)-1]
			for i, h := range d.OfflineWebSecond {
				h.Index = i
			}
			hand.Index = len(d.OnlineWebSecond)
			hand.State = Online
			hand.LastOnlineState = time.Now()
			d.OnlineWebSecond = append(d.OnlineWebSecond, hand)

			hand = d.OfflineWebThird[index]
			d.OfflineWebThird[index] =
				d.OfflineWebThird[len(d.OfflineWebThird)-1]
			d.OfflineWebThird[len(d.OfflineWebThird)-1] = nil
			d.OfflineWebThird =
				d.OfflineWebThird[:len(d.OfflineWebThird)-1]
			for i, h := range d.OfflineWebThird {
				h.Index = i
			}
			hand.Index = len(d.OnlineWebThird)
			hand.State = Online
			hand.LastOnlineState = time.Now()
			d.OnlineWebThird = append(d.OnlineWebThird, hand)
		}

		break
	}
}

func (d *Domain) downgradeHandler(hand *Handler) {
	d.Lock.Lock()
	defer d.Lock.Unlock()

	index := hand.Index
	state := hand.State

	switch state {
	case Online:
		hand = d.OnlineWebFirst[index]
		d.OnlineWebFirst[index] = d.OnlineWebFirst[len(d.OnlineWebFirst)-1]
		d.OnlineWebFirst[len(d.OnlineWebFirst)-1] = nil
		d.OnlineWebFirst = d.OnlineWebFirst[:len(d.OnlineWebFirst)-1]
		for i, h := range d.OnlineWebFirst {
			h.Index = i
		}
		hand.Index = len(d.UnknownMidWebFirst)
		hand.State = UnknownMid
		hand.LastState = time.Now()
		d.UnknownMidWebFirst = append(d.UnknownMidWebFirst, hand)

		hand = d.OnlineWebSecond[index]
		d.OnlineWebSecond[index] = d.OnlineWebSecond[len(d.OnlineWebSecond)-1]
		d.OnlineWebSecond[len(d.OnlineWebSecond)-1] = nil
		d.OnlineWebSecond = d.OnlineWebSecond[:len(d.OnlineWebSecond)-1]
		for i, h := range d.OnlineWebSecond {
			h.Index = i
		}
		hand.Index = len(d.UnknownMidWebSecond)
		hand.State = UnknownMid
		hand.LastState = time.Now()
		d.UnknownMidWebSecond = append(d.UnknownMidWebSecond, hand)

		hand = d.OnlineWebThird[index]
		d.OnlineWebThird[index] = d.OnlineWebThird[len(d.OnlineWebThird)-1]
		d.OnlineWebThird[len(d.OnlineWebThird)-1] = nil
		d.OnlineWebThird = d.OnlineWebThird[:len(d.OnlineWebThird)-1]
		for i, h := range d.OnlineWebThird {
			h.Index = i
		}
		hand.Index = len(d.UnknownMidWebThird)
		hand.State = UnknownMid
		hand.LastState = time.Now()
		d.UnknownMidWebThird = append(d.UnknownMidWebThird, hand)

		break
	case UnknownHigh:
		hand = d.UnknownHighWebFirst[index]
		d.UnknownHighWebFirst[index] =
			d.UnknownHighWebFirst[len(d.UnknownHighWebFirst)-1]
		d.UnknownHighWebFirst[len(d.UnknownHighWebFirst)-1] = nil
		d.UnknownHighWebFirst =
			d.UnknownHighWebFirst[:len(d.UnknownHighWebFirst)-1]
		for i, h := range d.UnknownHighWebFirst {
			h.Index = i
		}
		hand.Index = len(d.UnknownMidWebFirst)
		hand.State = UnknownMid
		hand.LastState = time.Now()
		d.UnknownMidWebFirst = append(d.UnknownMidWebFirst, hand)

		hand = d.UnknownHighWebSecond[index]
		d.UnknownHighWebSecond[index] =
			d.UnknownHighWebSecond[len(d.UnknownHighWebSecond)-1]
		d.UnknownHighWebSecond[len(d.UnknownHighWebSecond)-1] = nil
		d.UnknownHighWebSecond =
			d.UnknownHighWebSecond[:len(d.UnknownHighWebSecond)-1]
		for i, h := range d.UnknownHighWebSecond {
			h.Index = i
		}
		hand.Index = len(d.UnknownMidWebSecond)
		hand.State = UnknownMid
		hand.LastState = time.Now()
		d.UnknownMidWebSecond = append(d.UnknownMidWebSecond, hand)

		hand = d.UnknownHighWebThird[index]
		d.UnknownHighWebThird[index] =
			d.UnknownHighWebThird[len(d.UnknownHighWebThird)-1]
		d.UnknownHighWebThird[len(d.UnknownHighWebThird)-1] = nil
		d.UnknownHighWebThird =
			d.UnknownHighWebThird[:len(d.UnknownHighWebThird)-1]
		for i, h := range d.UnknownHighWebThird {
			h.Index = i
		}
		hand.Index = len(d.UnknownMidWebThird)
		hand.State = UnknownMid
		hand.LastState = time.Now()
		d.UnknownMidWebThird = append(d.UnknownMidWebThird, hand)

		break
	case UnknownMid:
		if time.Since(hand.LastState) > 1*time.Second {
			hand = d.UnknownMidWebFirst[index]
			d.UnknownMidWebFirst[index] =
				d.UnknownMidWebFirst[len(d.UnknownMidWebFirst)-1]
			d.UnknownMidWebFirst[len(d.UnknownMidWebFirst)-1] = nil
			d.UnknownMidWebFirst =
				d.UnknownMidWebFirst[:len(d.UnknownMidWebFirst)-1]
			for i, h := range d.UnknownMidWebFirst {
				h.Index = i
			}
			hand.Index = len(d.UnknownLowWebFirst)
			hand.State = UnknownLow
			hand.LastState = time.Now()
			d.UnknownLowWebFirst = append(d.UnknownLowWebFirst, hand)

			hand = d.UnknownMidWebSecond[index]
			d.UnknownMidWebSecond[index] =
				d.UnknownMidWebSecond[len(d.UnknownMidWebSecond)-1]
			d.UnknownMidWebSecond[len(d.UnknownMidWebSecond)-1] = nil
			d.UnknownMidWebSecond =
				d.UnknownMidWebSecond[:len(d.UnknownMidWebSecond)-1]
			for i, h := range d.UnknownMidWebSecond {
				h.Index = i
			}
			hand.Index = len(d.UnknownLowWebSecond)
			hand.State = UnknownLow
			hand.LastState = time.Now()
			d.UnknownLowWebSecond = append(d.UnknownLowWebSecond, hand)

			hand = d.UnknownMidWebThird[index]
			d.UnknownMidWebThird[index] =
				d.UnknownMidWebThird[len(d.UnknownMidWebThird)-1]
			d.UnknownMidWebThird[len(d.UnknownMidWebThird)-1] = nil
			d.UnknownMidWebThird =
				d.UnknownMidWebThird[:len(d.UnknownMidWebThird)-1]
			for i, h := range d.UnknownMidWebThird {
				h.Index = i
			}
			hand.Index = len(d.UnknownLowWebThird)
			hand.State = UnknownLow
			hand.LastState = time.Now()
			d.UnknownLowWebThird = append(d.UnknownLowWebThird, hand)
		}

		break
	case UnknownLow:
		if time.Since(hand.LastState) > 2*time.Second {
			hand = d.UnknownLowWebFirst[index]
			d.UnknownLowWebFirst[index] =
				d.UnknownLowWebFirst[len(d.UnknownLowWebFirst)-1]
			d.UnknownLowWebFirst[len(d.UnknownLowWebFirst)-1] = nil
			d.UnknownLowWebFirst =
				d.UnknownLowWebFirst[:len(d.UnknownLowWebFirst)-1]
			for i, h := range d.UnknownLowWebFirst {
				h.Index = i
			}
			hand.Index = len(d.OfflineWebFirst)
			hand.State = Offline
			hand.LastState = time.Now()
			d.OfflineWebFirst = append(d.OfflineWebFirst, hand)

			hand = d.UnknownLowWebSecond[index]
			d.UnknownLowWebSecond[index] =
				d.UnknownLowWebSecond[len(d.UnknownLowWebSecond)-1]
			d.UnknownLowWebSecond[len(d.UnknownLowWebSecond)-1] = nil
			d.UnknownLowWebSecond =
				d.UnknownLowWebSecond[:len(d.UnknownLowWebSecond)-1]
			for i, h := range d.UnknownLowWebSecond {
				h.Index = i
			}
			hand.Index = len(d.OfflineWebSecond)
			hand.State = Offline
			hand.LastState = time.Now()
			d.OfflineWebSecond = append(d.OfflineWebSecond, hand)

			hand = d.UnknownLowWebThird[index]
			d.UnknownLowWebThird[index] =
				d.UnknownLowWebThird[len(d.UnknownLowWebThird)-1]
			d.UnknownLowWebThird[len(d.UnknownLowWebThird)-1] = nil
			d.UnknownLowWebThird =
				d.UnknownLowWebThird[:len(d.UnknownLowWebThird)-1]
			for i, h := range d.UnknownLowWebThird {
				h.Index = i
			}
			hand.Index = len(d.OfflineWebThird)
			hand.State = Offline
			hand.LastState = time.Now()
			d.OfflineWebThird = append(d.OfflineWebThird, hand)
		}

		break
	case Offline:
		break
	}
}

func (d *Domain) offlineHandler(hand *Handler) {
	d.Lock.Lock()
	defer d.Lock.Unlock()

	index := hand.Index
	state := hand.State

	switch state {
	case Online:
		hand = d.OnlineWebFirst[index]
		d.OnlineWebFirst[index] = d.OnlineWebFirst[len(d.OnlineWebFirst)-1]
		d.OnlineWebFirst[len(d.OnlineWebFirst)-1] = nil
		d.OnlineWebFirst = d.OnlineWebFirst[:len(d.OnlineWebFirst)-1]
		for i, h := range d.OnlineWebFirst {
			h.Index = i
		}
		hand.Index = len(d.UnknownMidWebFirst)
		hand.State = Offline
		hand.LastState = time.Now()
		d.OfflineWebFirst = append(d.OfflineWebFirst, hand)

		hand = d.OnlineWebSecond[index]
		d.OnlineWebSecond[index] = d.OnlineWebSecond[len(d.OnlineWebSecond)-1]
		d.OnlineWebSecond[len(d.OnlineWebSecond)-1] = nil
		d.OnlineWebSecond = d.OnlineWebSecond[:len(d.OnlineWebSecond)-1]
		for i, h := range d.OnlineWebSecond {
			h.Index = i
		}
		hand.Index = len(d.OfflineWebSecond)
		hand.State = Offline
		hand.LastState = time.Now()
		d.OfflineWebSecond = append(d.OfflineWebSecond, hand)

		hand = d.OnlineWebThird[index]
		d.OnlineWebThird[index] = d.OnlineWebThird[len(d.OnlineWebThird)-1]
		d.OnlineWebThird[len(d.OnlineWebThird)-1] = nil
		d.OnlineWebThird = d.OnlineWebThird[:len(d.OnlineWebThird)-1]
		for i, h := range d.OnlineWebThird {
			h.Index = i
		}
		hand.Index = len(d.OfflineWebThird)
		hand.State = Offline
		hand.LastState = time.Now()
		d.OfflineWebThird = append(d.OfflineWebThird, hand)

		break
	case UnknownHigh:
		hand = d.UnknownHighWebFirst[index]
		d.UnknownHighWebFirst[index] =
			d.UnknownHighWebFirst[len(d.UnknownHighWebFirst)-1]
		d.UnknownHighWebFirst[len(d.UnknownHighWebFirst)-1] = nil
		d.UnknownHighWebFirst =
			d.UnknownHighWebFirst[:len(d.UnknownHighWebFirst)-1]
		for i, h := range d.UnknownHighWebFirst {
			h.Index = i
		}
		hand.Index = len(d.UnknownMidWebFirst)
		hand.State = Offline
		hand.LastState = time.Now()
		d.OfflineWebFirst = append(d.OfflineWebFirst, hand)

		hand = d.UnknownHighWebSecond[index]
		d.UnknownHighWebSecond[index] =
			d.UnknownHighWebSecond[len(d.UnknownHighWebSecond)-1]
		d.UnknownHighWebSecond[len(d.UnknownHighWebSecond)-1] = nil
		d.UnknownHighWebSecond =
			d.UnknownHighWebSecond[:len(d.UnknownHighWebSecond)-1]
		for i, h := range d.UnknownHighWebSecond {
			h.Index = i
		}
		hand.Index = len(d.OfflineWebSecond)
		hand.State = Offline
		hand.LastState = time.Now()
		d.OfflineWebSecond = append(d.OfflineWebSecond, hand)

		hand = d.UnknownHighWebThird[index]
		d.UnknownHighWebThird[index] =
			d.UnknownHighWebThird[len(d.UnknownHighWebThird)-1]
		d.UnknownHighWebThird[len(d.UnknownHighWebThird)-1] = nil
		d.UnknownHighWebThird =
			d.UnknownHighWebThird[:len(d.UnknownHighWebThird)-1]
		for i, h := range d.UnknownHighWebThird {
			h.Index = i
		}
		hand.Index = len(d.OfflineWebThird)
		hand.State = Offline
		hand.LastState = time.Now()
		d.OfflineWebThird = append(d.OfflineWebThird, hand)

		break
	case UnknownMid:
		if time.Since(hand.LastState) > 1*time.Second {
			hand = d.UnknownMidWebFirst[index]
			d.UnknownMidWebFirst[index] =
				d.UnknownMidWebFirst[len(d.UnknownMidWebFirst)-1]
			d.UnknownMidWebFirst[len(d.UnknownMidWebFirst)-1] = nil
			d.UnknownMidWebFirst =
				d.UnknownMidWebFirst[:len(d.UnknownMidWebFirst)-1]
			for i, h := range d.UnknownMidWebFirst {
				h.Index = i
			}
			hand.Index = len(d.UnknownLowWebFirst)
			hand.State = Offline
			hand.LastState = time.Now()
			d.OfflineWebFirst = append(d.OfflineWebFirst, hand)

			hand = d.UnknownMidWebSecond[index]
			d.UnknownMidWebSecond[index] =
				d.UnknownMidWebSecond[len(d.UnknownMidWebSecond)-1]
			d.UnknownMidWebSecond[len(d.UnknownMidWebSecond)-1] = nil
			d.UnknownMidWebSecond =
				d.UnknownMidWebSecond[:len(d.UnknownMidWebSecond)-1]
			for i, h := range d.UnknownMidWebSecond {
				h.Index = i
			}
			hand.Index = len(d.OfflineWebSecond)
			hand.State = Offline
			hand.LastState = time.Now()
			d.OfflineWebSecond = append(d.OfflineWebSecond, hand)

			hand = d.UnknownMidWebThird[index]
			d.UnknownMidWebThird[index] =
				d.UnknownMidWebThird[len(d.UnknownMidWebThird)-1]
			d.UnknownMidWebThird[len(d.UnknownMidWebThird)-1] = nil
			d.UnknownMidWebThird =
				d.UnknownMidWebThird[:len(d.UnknownMidWebThird)-1]
			for i, h := range d.UnknownMidWebThird {
				h.Index = i
			}
			hand.Index = len(d.OfflineWebThird)
			hand.State = Offline
			hand.LastState = time.Now()
			d.OfflineWebThird = append(d.OfflineWebThird, hand)
		}

		break
	case UnknownLow:
		if time.Since(hand.LastState) > 2*time.Second {
			hand = d.UnknownLowWebFirst[index]
			d.UnknownLowWebFirst[index] =
				d.UnknownLowWebFirst[len(d.UnknownLowWebFirst)-1]
			d.UnknownLowWebFirst[len(d.UnknownLowWebFirst)-1] = nil
			d.UnknownLowWebFirst =
				d.UnknownLowWebFirst[:len(d.UnknownLowWebFirst)-1]
			for i, h := range d.UnknownLowWebFirst {
				h.Index = i
			}
			hand.Index = len(d.OfflineWebFirst)
			hand.State = Offline
			hand.LastState = time.Now()
			d.OfflineWebFirst = append(d.OfflineWebFirst, hand)

			hand = d.UnknownLowWebSecond[index]
			d.UnknownLowWebSecond[index] =
				d.UnknownLowWebSecond[len(d.UnknownLowWebSecond)-1]
			d.UnknownLowWebSecond[len(d.UnknownLowWebSecond)-1] = nil
			d.UnknownLowWebSecond =
				d.UnknownLowWebSecond[:len(d.UnknownLowWebSecond)-1]
			for i, h := range d.UnknownLowWebSecond {
				h.Index = i
			}
			hand.Index = len(d.OfflineWebSecond)
			hand.State = Offline
			hand.LastState = time.Now()
			d.OfflineWebSecond = append(d.OfflineWebSecond, hand)

			hand = d.UnknownLowWebThird[index]
			d.UnknownLowWebThird[index] =
				d.UnknownLowWebThird[len(d.UnknownLowWebThird)-1]
			d.UnknownLowWebThird[len(d.UnknownLowWebThird)-1] = nil
			d.UnknownLowWebThird =
				d.UnknownLowWebThird[:len(d.UnknownLowWebThird)-1]
			for i, h := range d.UnknownLowWebThird {
				h.Index = i
			}
			hand.Index = len(d.OfflineWebThird)
			hand.State = Offline
			hand.LastState = time.Now()
			d.OfflineWebThird = append(d.OfflineWebThird, hand)
		}

		break
	case Offline:
		break
	}
}

func (d *Domain) ResponseHandler(hand *Handler, resp *http.Response) error {
	if hand.State != Online && resp.StatusCode < 500 {
		d.upgradeHandler(hand)
	}

	return nil
}

func (d *Domain) ErrorHandlerFirst(hand *Handler, rw http.ResponseWriter,
	r *http.Request, err error) {

	if _, ok := err.(*WebSocketBlock); ok {
		return
	}

	d.downgradeHandler(hand)
	d.ServeHTTPSecond(rw, r)
}

func (d *Domain) ErrorHandlerSecond(hand *Handler, rw http.ResponseWriter,
	r *http.Request, err error) {

	if _, ok := err.(*WebSocketBlock); ok {
		return
	}

	d.downgradeHandler(hand)
	d.ServeHTTPThird(rw, r)
}

func (d *Domain) ErrorHandlerThird(hand *Handler, rw http.ResponseWriter,
	r *http.Request, err error) {

	if _, ok := err.(*WebSocketBlock); ok {
		return
	}

	d.downgradeHandler(hand)
	rw.WriteHeader(http.StatusBadGateway)
}
