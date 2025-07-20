package dhcpc

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
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
}

func (d *Dhcpc) startSync() {
	im := &Imds{
		Address: d.ImdsAddress,
		Port:    d.ImdsPort,
		Secret:  d.ImdsSecret,
	}

	for {
		e := im.Sync(d.lease)
		if e != nil {
			logger.WithFields(logger.Fields{
				"error": e,
			}).Error("dhcpc: Failed to sync lease with imds")
			time.Sleep(1 * time.Second)
		}

		time.Sleep(60 * time.Second)
	}
}

func (d *Dhcpc) run() (err error) {
	d.lease = &Lease{
		Iface:   d.DhcpIface,
		Address: d.DhcpIp,
	}

	go d.startSync()

	for {
		for {
			ok, e := d.lease.Exchange()
			if e != nil {
				logger.WithFields(logger.Fields{
					"interface": d.DhcpIface,
					"address":   d.lease.Address.String(),
					"error":     e,
				}).Error("dhcpc: Failed to exchange lease")
				time.Sleep(500 * time.Millisecond)
				continue
			}

			if !ok {
				logger.WithFields(logger.Fields{
					"interface": d.DhcpIface,
				}).Error("dhcpc: Failed to receive lease")
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
		}).Info("dhcpc: Exchanged")

		ready := false
		for i := 0; i < 20; i++ {
			ready = d.lease.IfaceReady()
			if ready {
				break
			}

			time.Sleep(500 * time.Millisecond)
		}

		if !ready {
			logger.WithFields(logger.Fields{
				"interface": d.DhcpIface,
				"address":   d.lease.Address.String(),
			}).Error("dhcpc: Interface ready timeout")
		}

		for {
			time.Sleep(60 * time.Second)

			ready := d.lease.IfaceReady()
			if !ready {
				logger.WithFields(logger.Fields{
					"interface": d.DhcpIface,
					"address":   d.lease.Address.String(),
				}).Error("dhcpc: Interface not ready")
				break
			}

			ok, e := d.lease.Renew()
			if e != nil {
				logger.WithFields(logger.Fields{
					"interface": d.DhcpIface,
					"address":   d.lease.Address.String(),
					"error":     e,
				}).Error("dhcpc: Failed to renew lease")
				break
			}

			if !ok {
				logger.WithFields(logger.Fields{
					"interface": d.DhcpIface,
					"address":   d.lease.Address.String(),
					"error":     err,
				}).Error("dhcpc: Failed to receive lease renewal")
				break
			}

			logger.WithFields(logger.Fields{
				"interface": d.DhcpIface,
				"address":   d.lease.Address.String(),
				"gateway":   d.lease.Gateway.String(),
				"server":    d.lease.ServerAddress.String(),
				"time":      d.lease.LeaseTime.String(),
			}).Info("dhcpc: Renewed")
		}
	}
}

func (d *Dhcpc) Run() {
	for {
		err := d.run()
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

	logger.AddHandler(func(record *logger.Record) {
		fmt.Print(record.String())
	})

	ip4 := false
	flag.BoolVar(&ip4, "ip4", false, "Enable IPv4")

	ip6 := false
	flag.BoolVar(&ip6, "ip6", false, "Enable IPv6")

	flag.Parse()

	imdsPortInt, _ := strconv.Atoi(imdsPort)

	client := &Dhcpc{
		ImdsAddress: imdsAddress,
		ImdsPort:    imdsPortInt,
		ImdsSecret:  imdsSecret,
		DhcpIface:   dhcpIface,
		DhcpIface6:  dhcpIface6,
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

	client.Run()

	return
}
