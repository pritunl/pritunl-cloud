package dnss

import (
	"context"
	"time"

	"github.com/coredns/coredns/plugin/forward"
	"github.com/coredns/coredns/plugin/pkg/proxy"
	"github.com/coredns/coredns/plugin/pkg/transport"
	"github.com/miekg/dns"
)

type ForwardMulti struct {
	primary   *forward.Forward
	secondary *forward.Forward
}

func (f *ForwardMulti) Name() string {
	return "pritunl-cloud"
}

func (f *ForwardMulti) ServeDNS(ctx context.Context,
	w dns.ResponseWriter, r *dns.Msg) (int, error) {

	if f.secondary == nil {
		return f.primary.ServeDNS(ctx, w, r)
	}

	rw := &Response{
		ResponseWriter: w,
	}
	code, err := f.primary.ServeDNS(ctx, rw, r)

	if err == nil && rw.msg != nil &&
		rw.msg.Rcode != dns.RcodeServerFailure {

		w.WriteMsg(rw.msg)
		return code, nil
	}

	return f.secondary.ServeDNS(ctx, w, r)
}

func (f *ForwardMulti) Shutdown() {
	if f.primary != nil {
		f.primary.OnShutdown()
	}
	if f.secondary != nil {
		f.secondary.OnShutdown()
	}
}

func NewForwardMulti(primary, secondary []string) *ForwardMulti {
	if len(primary) == 0 {
		return nil
	}

	primaryFwd := forward.New()
	for _, upstream := range primary {
		prxy := proxy.NewProxy(upstream, upstream, transport.DNS)
		prxy.SetReadTimeout(2 * time.Second)
		prxy.Start(60 * time.Second)
		primaryFwd.SetProxy(prxy)
	}

	var secondaryFwd *forward.Forward
	if len(secondary) > 0 {
		secondaryFwd = forward.New()
		for _, upstream := range secondary {
			prxy := proxy.NewProxy(upstream, upstream, transport.DNS)
			prxy.SetReadTimeout(2 * time.Second)
			prxy.Start(60 * time.Second)
			secondaryFwd.SetProxy(prxy)
		}
	}

	return &ForwardMulti{
		primary:   primaryFwd,
		secondary: secondaryFwd,
	}
}
