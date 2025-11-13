package dnss

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coredns/coredns/plugin/forward"
	"github.com/coredns/coredns/plugin/pkg/proxy"
	"github.com/coredns/coredns/plugin/pkg/transport"
	"github.com/miekg/dns"
)

func Run() {
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

	serverUdp := &dns.Server{
		Addr:    "169.254.169.254:53",
		Net:     "udp",
		Handler: mux,
	}

	serverTcp := &dns.Server{
		Addr:    "169.254.169.254:53",
		Net:     "tcp",
		Handler: mux,
	}

	go func() {
		err := serverUdp.ListenAndServe()
		if err != nil {
			log.Fatalf("Failed to start udp server: %s", err)
		}
	}()
	go func() {
		err := serverTcp.ListenAndServe()
		if err != nil {

			log.Fatalf("Failed to start tcp server: %s", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down...")
	serverUdp.Shutdown()
	serverTcp.Shutdown()
}
