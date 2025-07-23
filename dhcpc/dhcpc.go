package dhcpc

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pritunl/tools/logger"
)

type Dhcpc struct {
	ImdsAddress string
	ImdsPort    int
	ImdsSecret  string
	DhcpIface   string
	DhcpIface6  string
	DhcpIp      *net.IPNet
	DhcpIp6     *net.IPNet
	lease       *Lease
	syncTrigger chan struct{}
}

func (d *Dhcpc) startSync() {
	im := &Imds{
		Address: d.ImdsAddress,
		Port:    d.ImdsPort,
		Secret:  d.ImdsSecret,
	}

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		e := im.Sync(d.lease)
		if e != nil {
			logger.WithFields(logger.Fields{
				"error": e,
			}).Error("dhcpc: Failed to sync lease with imds")
			time.Sleep(1 * time.Second)
		}

		logger.WithFields(logger.Fields{
			"interface": d.DhcpIface,
			"address":   d.lease.Address.String(),
			"gateway":   d.lease.Gateway.String(),
			"server":    d.lease.ServerAddress.String(),
			"time":      d.lease.LeaseTime.String(),
		}).Info("dhcpc: Synced")

		select {
		case <-ticker.C:
		case <-d.syncTrigger:
			ticker.Reset(60 * time.Second)
		}
	}
}

func (d *Dhcpc) sync() {
	select {
	case d.syncTrigger <- struct{}{}:
	default:
	}
}

func (d *Dhcpc) run4() (err error) {
	d.lease.Gateway = nil
	d.lease.ServerAddress = nil
	d.lease.LeaseTime = 0
	d.lease.TransactionId = ""

	for {
		for {
			ok, e := d.lease.Exchange4()
			if e != nil {
				logger.WithFields(logger.Fields{
					"interface": d.DhcpIface,
					"address":   d.lease.Address.String(),
					"error":     e,
				}).Error("dhcpc: Failed to exchange lease4")
				time.Sleep(500 * time.Millisecond)
				continue
			}

			if !ok {
				logger.WithFields(logger.Fields{
					"interface": d.DhcpIface,
				}).Error("dhcpc: Failed to receive lease4")
				time.Sleep(1000 * time.Millisecond)
			}

			break
		}

		logger.WithFields(logger.Fields{
			"interface": d.DhcpIface,
			"address":   d.lease.Address.String(),
			"gateway":   d.lease.Gateway.String(),
			"server":    d.lease.ServerAddress.String(),
			"time":      d.lease.LeaseTime.String(),
		}).Info("dhcpc: Exchanged ipv4")

		d.sync()

		ready4 := false
		for i := 0; i < 20; i++ {
			ready4, _ = d.lease.IfaceReady()
			if ready4 {
				break
			}

			time.Sleep(500 * time.Millisecond)
		}

		if !ready4 {
			logger.WithFields(logger.Fields{
				"interface": d.DhcpIface,
				"address":   d.lease.Address.String(),
			}).Error("dhcpc: Interface4 ready timeout")
		}

		for {
			time.Sleep(60 * time.Second)

			ready4, _ := d.lease.IfaceReady()
			if !ready4 {
				logger.WithFields(logger.Fields{
					"interface": d.DhcpIface,
					"address":   d.lease.Address.String(),
				}).Error("dhcpc: Interface4 not ready")
				break
			}

			ok, e := d.lease.Renew4()
			if e != nil {
				logger.WithFields(logger.Fields{
					"interface": d.DhcpIface,
					"address":   d.lease.Address.String(),
					"error":     e,
				}).Error("dhcpc: Failed to renew lease4")
				break
			}

			if !ok {
				logger.WithFields(logger.Fields{
					"interface": d.DhcpIface,
					"address":   d.lease.Address.String(),
					"error":     err,
				}).Error("dhcpc: Failed to receive lease4 renewal")
				break
			}

			logger.WithFields(logger.Fields{
				"interface": d.DhcpIface,
				"address":   d.lease.Address.String(),
				"gateway":   d.lease.Gateway.String(),
				"server":    d.lease.ServerAddress.String(),
				"time":      d.lease.LeaseTime.String(),
			}).Info("dhcpc: Renewed ipv4")

			d.sync()
		}
	}
}

func (d *Dhcpc) run6() (err error) {
	d.lease.ServerAddress6 = nil
	d.lease.LeaseTime6 = 0
	d.lease.TransactionId6 = ""
	d.lease.IaId6 = [4]byte{}
	d.lease.ServerId6 = nil

	for {
		for {
			ok, e := d.lease.Exchange6()
			if e != nil {
				logger.WithFields(logger.Fields{
					"interface6": d.DhcpIface6,
					"address6":   d.lease.Address6.String(),
					"error":      e,
				}).Error("dhcpc: Failed to exchange lease6")
				time.Sleep(500 * time.Millisecond)
				continue
			}

			if !ok {
				logger.WithFields(logger.Fields{
					"interface6": d.DhcpIface6,
				}).Error("dhcpc: Failed to receive lease6")
				time.Sleep(1000 * time.Millisecond)
			}

			break
		}

		logger.WithFields(logger.Fields{
			"interface6": d.DhcpIface6,
			"address6":   d.lease.Address6.String(),
			"server6":    d.lease.ServerAddress6.String(),
			"time6":      d.lease.LeaseTime6.String(),
		}).Info("dhcpc: Exchanged ipv6")

		d.sync()

		ready6 := false
		for i := 0; i < 20; i++ {
			_, ready6 = d.lease.IfaceReady()
			if ready6 {
				break
			}

			time.Sleep(500 * time.Millisecond)
		}

		if !ready6 {
			logger.WithFields(logger.Fields{
				"interface6": d.DhcpIface6,
				"address6":   d.lease.Address6.String(),
			}).Error("dhcpc: Interface6 ready timeout")
		}

		for {
			time.Sleep(60 * time.Second)

			_, ready6 := d.lease.IfaceReady()
			if !ready6 {
				logger.WithFields(logger.Fields{
					"interface6": d.DhcpIface6,
					"address6":   d.lease.Address6.String(),
				}).Error("dhcpc: Interface6 not ready")
				break
			}

			ok, e := d.lease.Renew4()
			if e != nil {
				logger.WithFields(logger.Fields{
					"interface6": d.DhcpIface6,
					"address6":   d.lease.Address6.String(),
					"error":      e,
				}).Error("dhcpc: Failed to renew lease6")
				break
			}

			if !ok {
				logger.WithFields(logger.Fields{
					"interface6": d.DhcpIface6,
					"address6":   d.lease.Address6.String(),
					"error":      err,
				}).Error("dhcpc: Failed to receive lease6 renewal")
				break
			}

			logger.WithFields(logger.Fields{
				"interface6": d.DhcpIface6,
				"address6":   d.lease.Address6.String(),
				"server6":    d.lease.ServerAddress6.String(),
				"time6":      d.lease.LeaseTime6.String(),
			}).Info("dhcpc: Renewed ipv6")

			d.sync()
		}
	}
}

func (d *Dhcpc) Run(ip4, ip6 bool) {
	d.lease = &Lease{
		Iface:    d.DhcpIface,
		Iface6:   d.DhcpIface6,
		Address:  d.DhcpIp,
		Address6: d.DhcpIp6,
	}

	go d.startSync()

	waiters := &sync.WaitGroup{}
	if ip4 {
		waiters.Add(1)
		go func() {
			defer waiters.Done()
			for {
				err := d.run4()
				if err != nil {
					logger.WithFields(logger.Fields{
						"interface": d.DhcpIface,
						"address":   d.DhcpIp,
						"address6":  d.DhcpIp6,
						"error":     err,
					}).Error("dhcpc: Run error")
				}

				time.Sleep(3 * time.Second)
			}
		}()
	}

	if ip6 {
		waiters.Add(1)
		go func() {
			defer waiters.Done()
			for {
				err := d.run6()
				if err != nil {
					logger.WithFields(logger.Fields{
						"interface": d.DhcpIface,
						"address":   d.DhcpIp,
						"address6":  d.DhcpIp6,
						"error":     err,
					}).Error("dhcpc: Run error")
				}

				time.Sleep(3 * time.Second)
			}
		}()
	}

	waiters.Wait()
}

func Main() (err error) {
	imdsAddress := os.Getenv("IMDS_ADDRESS")
	imdsPort := os.Getenv("IMDS_PORT")
	imdsSecret := os.Getenv("IMDS_SECRET")
	dhcpIface := os.Getenv("DHCP_IFACE")
	dhcpIface6 := os.Getenv("DHCP_IFACE6")
	dhcpIp := os.Getenv("DHCP_IP")
	dhcpIp6 := os.Getenv("DHCP_IP6")
	os.Unsetenv("IMDS_ADDRESS")
	os.Unsetenv("IMDS_PORT")
	os.Unsetenv("IMDS_SECRET")
	os.Unsetenv("DHCP_IFACE")
	os.Unsetenv("DHCP_IFACE6")
	os.Unsetenv("DHCP_IP")
	os.Unsetenv("DHCP_IP6")

	logger.Init(
		logger.SetTimeFormat(""),
	)

	logLock := sync.Mutex{}
	logger.AddHandler(func(record *logger.Record) {
		logLock.Lock()
		fmt.Print(record.String())
		logLock.Unlock()
	})

	ip4 := false
	flag.BoolVar(&ip4, "ip4", false, "Enable IPv4")

	ip6 := false
	flag.BoolVar(&ip6, "ip6", false, "Enable IPv6")

	flag.Parse()

	imdsPortInt, _ := strconv.Atoi(imdsPort)

	client := &Dhcpc{
		ImdsAddress: strings.Split(imdsAddress, "/")[0],
		ImdsPort:    imdsPortInt,
		ImdsSecret:  imdsSecret,
		DhcpIface:   dhcpIface,
		DhcpIface6:  dhcpIface6,
		syncTrigger: make(chan struct{}, 1),
	}

	if dhcpIp != "" {
		ip, ipnet, _ := net.ParseCIDR(dhcpIp)
		if ip != nil && ipnet != nil {
			ipnet.IP = ip
			client.DhcpIp = ipnet
		}
	}

	if dhcpIp6 != "" {
		ip, ipnet, _ := net.ParseCIDR(dhcpIp6)
		if ip != nil && ipnet != nil {
			ipnet.IP = ip
			client.DhcpIp6 = ipnet
		}
	}

	client.Run(ip4, ip6)

	return
}
