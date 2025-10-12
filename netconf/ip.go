package netconf

import (
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/dhcpc"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/imds"
	"github.com/pritunl/pritunl-cloud/iproute"
	"github.com/pritunl/pritunl-cloud/iptables"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/store"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/tools/commander"
	"github.com/sirupsen/logrus"
)

func (n *NetConf) ipExternal(db *database.Database) (err error) {
	if n.NetworkMode == node.Dhcp || n.NetworkMode6 == node.Dhcp ||
		n.NetworkMode6 == node.DhcpSlaac {

		err = dhcpc.Start(
			db,
			n.Virt,
			n.SpaceExternalIface,
			n.SpaceExternalIface,
			n.NetworkMode == node.Dhcp,
			n.NetworkMode6 == node.Dhcp || n.NetworkMode6 == node.DhcpSlaac,
		)
		if err != nil {
			return
		}

		var imdsErr error
		ip4 := false
		ip6 := false
		ipTimeout := settings.Hypervisor.IpTimeout * 4
		for i := 0; i < ipTimeout; i++ {
			stat, e := imds.State(db, n.Virt.Id, n.Virt.ImdsHostSecret)
			if e != nil {
				imdsErr = e
				time.Sleep(250 * time.Millisecond)
				continue
			}

			if stat == nil {
				time.Sleep(250 * time.Millisecond)
				continue
			}

			if stat.DhcpIp != nil && stat.DhcpGateway != nil {
				_, err = utils.ExecCombinedOutputLogged(
					[]string{"File exists", "already assigned"},
					"ip", "netns", "exec", n.Namespace,
					"ip", "addr",
					"add", stat.DhcpIp.String(),
					"dev", n.SpaceExternalIface,
				)
				if err != nil {
					return
				}

				_, err = utils.ExecCombinedOutputLogged(
					[]string{"File exists"},
					"ip", "netns", "exec", n.Namespace,
					"ip", "route",
					"add", "default",
					"via", stat.DhcpGateway.String(),
					"dev", n.SpaceExternalIface,
				)
				if err != nil {
					return
				}

				ip4 = true
			}

			if stat.DhcpIp6 != nil {
				_, err = utils.ExecCombinedOutputLogged(
					[]string{"File exists", "already assigned"},
					"ip", "netns", "exec", n.Namespace,
					"ip", "addr",
					"add", stat.DhcpIp6.String(),
					"dev", n.SpaceExternalIface,
				)
				if err != nil {
					return
				}

				ip6 = true
			}

			if (n.NetworkMode != node.Dhcp || ip4) &&
				((n.NetworkMode6 != node.Dhcp &&
					n.NetworkMode6 != node.DhcpSlaac) || ip6) {

				break
			}

			time.Sleep(250 * time.Millisecond)
		}

		if !ip4 && n.NetworkMode == node.Dhcp {
			if imdsErr != nil {
				logrus.WithFields(logrus.Fields{
					"instance": n.Virt.Id.Hex(),
					"dhcp4":    ip4,
					"dhcp6":    ip6,
					"error":    imdsErr,
				}).Error("netconf: DHCP IPv4 timeout")
			} else {
				logrus.WithFields(logrus.Fields{
					"instance": n.Virt.Id.Hex(),
				}).Error("netconf: DHCP IPv4 timeout")
			}
		}

		if !ip6 && (n.NetworkMode6 == node.Dhcp ||
			n.NetworkMode6 == node.DhcpSlaac) {

			if imdsErr != nil {
				logrus.WithFields(logrus.Fields{
					"instance": n.Virt.Id.Hex(),
					"dhcp4":    ip4,
					"dhcp6":    ip6,
					"error":    imdsErr,
				}).Error("netconf: DHCP IPv6 timeout")
			} else {
				logrus.WithFields(logrus.Fields{
					"instance": n.Virt.Id.Hex(),
					"dhcp4":    ip4,
					"dhcp6":    ip6,
				}).Error("netconf: DHCP IPv6 timeout")
			}
		}
	}

	if n.NetworkMode == node.Static {
		if n.SpaceExternalIfaceMod != "" {
			_, err = utils.ExecCombinedOutputLogged(
				[]string{"File exists"},
				"ip", "netns", "exec", n.Namespace,
				"ip", "addr",
				"add", n.ExternalAddrCidr,
				"dev", n.SpaceExternalIfaceMod,
			)
			if err != nil {
				return
			}

			_, err = utils.ExecCombinedOutputLogged(
				[]string{"File exists"},
				"ip", "netns", "exec", n.Namespace,
				"ip", "route",
				"add", "default",
				"via", n.ExternalGatewayAddr.String(),
				"dev", n.SpaceExternalIfaceMod,
			)
			if err != nil {
				return
			}
		} else {
			_, err = utils.ExecCombinedOutputLogged(
				[]string{"File exists"},
				"ip", "netns", "exec", n.Namespace,
				"ip", "addr",
				"add", n.ExternalAddrCidr,
				"dev", n.SpaceExternalIface,
			)
			if err != nil {
				return
			}

			_, err = utils.ExecCombinedOutputLogged(
				[]string{"File exists"},
				"ip", "netns", "exec", n.Namespace,
				"ip", "route",
				"add", "default",
				"via", n.ExternalGatewayAddr.String(),
				"dev", n.SpaceExternalIface,
			)
			if err != nil {
				return
			}
		}
	}

	if n.NetworkMode6 == node.Static {
		if n.SpaceExternalIfaceMod6 != "" {
			_, err = utils.ExecCombinedOutputLogged(
				[]string{"File exists"},
				"ip", "netns", "exec", n.Namespace,
				"ip", "addr",
				"add", n.ExternalAddrCidr6,
				"dev", n.SpaceExternalIfaceMod6,
			)
			if err != nil {
				return
			}

			if n.ExternalGatewayAddr6 != nil {
				_, err = utils.ExecCombinedOutputLogged(
					[]string{"File exists"},
					"ip", "netns", "exec", n.Namespace,
					"ip", "-6", "route",
					"add", "default",
					"via", n.ExternalGatewayAddr6.String(),
					"dev", n.SpaceExternalIfaceMod6,
				)
				if err != nil {
					return
				}
			}
		} else {
			_, err = utils.ExecCombinedOutputLogged(
				[]string{"File exists"},
				"ip", "netns", "exec", n.Namespace,
				"ip", "addr",
				"add", n.ExternalAddrCidr6,
				"dev", n.SpaceExternalIface,
			)
			if err != nil {
				return
			}

			if n.ExternalGatewayAddr6 != nil {
				_, err = utils.ExecCombinedOutputLogged(
					[]string{"File exists"},
					"ip", "netns", "exec", n.Namespace,
					"ip", "-6", "route",
					"add", "default",
					"via", n.ExternalGatewayAddr6.String(),
					"dev", n.SpaceExternalIface,
				)
				if err != nil {
					return
				}
			}
		}
	}

	return
}

func (n *NetConf) ipHost(db *database.Database) (err error) {
	if n.HostNetwork {
		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "netns", "exec", n.Namespace,
			"ip", "addr",
			"add", n.HostAddrCidr,
			"dev", n.SpaceHostIface,
		)
		if err != nil {
			return
		}

		if n.NetworkMode == node.Disabled {
			_, err = utils.ExecCombinedOutputLogged(
				[]string{"File exists"},
				"ip", "netns", "exec", n.Namespace,
				"ip", "route",
				"add", "default",
				"via", n.HostGatewayAddr.String(),
			)
			if err != nil {
				return
			}
		}
	}

	return
}

func (n *NetConf) ipNodePort(db *database.Database) (err error) {
	if n.NodePortNetwork {
		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "netns", "exec", n.Namespace,
			"ip", "addr",
			"add", n.NodePortAddrCidr,
			"dev", n.SpaceNodePortIface,
		)
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) ipDetect(db *database.Database) (err error) {
	time.Sleep(250 * time.Millisecond)

	ipTimeout := settings.Hypervisor.IpTimeout * 4
	ipTimeout6 := settings.Hypervisor.IpTimeout6 * 4

	pubAddr := ""
	pubAddr6 := ""
	if n.NetworkMode != node.Disabled && n.NetworkMode != node.Cloud {
		for i := 0; i < ipTimeout; i++ {
			address, address6, e := iproute.AddressGetIfaceMod(
				n.Namespace, n.SpaceExternalIface)
			if e != nil {
				err = e
				return
			}

			if n.NetworkMode6 != node.Disabled &&
				n.NetworkMode6 != node.Cloud {

				if address != nil {
					pubAddr = address.Local
				}

				if address != nil && address6 != nil {
					if address6 != nil {
						pubAddr6 = address6.Local
					}
					break
				}
			} else if address != nil {
				pubAddr = address.Local
				break
			}

			time.Sleep(250 * time.Millisecond)
		}

		if pubAddr == "" {
			err = &errortypes.NetworkError{
				errors.New("qemu: Instance missing IPv4 address"),
			}
			return
		}
	} else if n.NetworkMode6 != node.Disabled &&
		n.NetworkMode6 != node.Cloud {

		for i := 0; i < ipTimeout6; i++ {
			_, address6, e := iproute.AddressGetIfaceMod(
				n.Namespace, n.SpaceExternalIface)
			if e != nil {
				err = e
				return
			}

			if address6 != nil {
				pubAddr6 = address6.Local
				break
			}

			time.Sleep(250 * time.Millisecond)
		}

		if pubAddr6 == "" {
			err = &errortypes.NetworkError{
				errors.New("qemu: Instance missing IPv6 address"),
			}
			return
		}
	}

	n.PublicAddress = pubAddr
	if n.ExternalAddrCidr6 != "" {
		n.PublicAddress6 = n.ExternalAddrCidr6
	} else {
		n.PublicAddress6 = pubAddr6
	}

	return
}

func (n *NetConf) ipHostIptables(db *database.Database) (err error) {
	if n.HostNetwork {
		iptables.Lock()
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"iptables", "-t", "nat",
			"-A", "POSTROUTING",
			"-s", n.InternalAddr.String()+"/32",
			"-d", n.InternalAddr.String()+"/32",
			"-m", "comment",
			"--comment", "pritunl_cloud_host_nat",
			"-j", "SNAT",
			"--to", n.HostAddr.String(),
		)
		iptables.Unlock()
		if err != nil {
			return
		}

		if n.HostNat {
			iptables.Lock()
			_, err = utils.ExecCombinedOutputLogged(
				nil,
				"ip", "netns", "exec", n.Namespace,
				"iptables", "-t", "nat",
				"-A", "POSTROUTING",
				"-s", n.InternalAddr.String()+"/32",
				"-o", n.SpaceHostIface,
				"-m", "comment",
				"--comment", "pritunl_cloud_host_nat",
				"-j", "MASQUERADE",
			)
			iptables.Unlock()
			if err != nil {
				return
			}
		} else {
			iptables.Lock()
			_, err = utils.ExecCombinedOutputLogged(
				nil,
				"ip", "netns", "exec", n.Namespace,
				"iptables", "-t", "nat",
				"-A", "POSTROUTING",
				"-s", n.InternalAddr.String()+"/32",
				"-d", n.HostSubnet,
				"-o", n.SpaceHostIface,
				"-m", "comment",
				"--comment", "pritunl_cloud_host_nat",
				"-j", "MASQUERADE",
			)
			iptables.Unlock()
			if err != nil {
				return
			}
		}

		iptables.Lock()
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"iptables", "-t", "nat",
			"-A", "PREROUTING",
			"-d", n.HostAddr.String()+"/32",
			"-m", "comment",
			"--comment", "pritunl_cloud_host_nat",
			"-j", "DNAT",
			"--to-destination", n.InternalAddr.String(),
		)
		iptables.Unlock()
		if err != nil {
			return
		}
	}

	return
}

func (n *NetConf) ipDatabase(db *database.Database) (err error) {
	store.RemAddress(n.Virt.Id)
	store.RemRoutes(n.Virt.Id)
	store.RemArp(n.Virt.Id)

	hostIps := []string{}
	if n.HostAddr != nil {
		hostIps = append(hostIps, n.HostAddr.String())
	}

	nodePortIps := []string{}
	if n.NodePortAddr != nil {
		nodePortIps = append(nodePortIps, n.NodePortAddr.String())
	}

	coll := db.Instances()
	err = coll.UpdateId(n.Virt.Id, &bson.M{
		"$set": &bson.M{
			"private_ips":  []string{n.InternalAddr.String()},
			"private_ips6": []string{n.InternalAddr6.String()},
			"gateway_ips":  []string{n.InternalGatewayAddrCidr},
			"gateway_ips6": []string{
				n.InternalGatewayAddr6.String() + "/64"},
			"network_namespace": n.Namespace,
			"host_ips":          hostIps,
			"node_port_ips":     nodePortIps,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		} else {
			return
		}
	}

	if !n.Virt.Deployment.IsZero() {
		coll = db.Deployments()

		hostIps := []string{}
		if n.HostAddr != nil {
			hostIps = append(hostIps, n.HostAddr.String())
		}
		privateIps := []string{}
		if n.InternalAddr != nil {
			privateIps = append(privateIps, n.InternalAddr.String())
		}
		privateIps6 := []string{}
		if n.InternalAddr6 != nil {
			privateIps6 = append(privateIps6, n.InternalAddr6.String())
		}

		err = coll.UpdateId(n.Virt.Deployment, &bson.M{
			"$set": &bson.M{
				"instance_data.host_ips":     hostIps,
				"instance_data.private_ips":  privateIps,
				"instance_data.private_ips6": privateIps6,
			},
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}
	}

	return
}

func (n *NetConf) ipInit6(db *database.Database) (err error) {
	if n.NetworkMode6 != node.Disabled && n.NetworkMode6 != node.Cloud &&
		n.PublicAddress6 != "" && !settings.Hypervisor.NoIpv6PingInit {

		for i := 0; i < 3; i++ {
			time.Sleep(200 * time.Millisecond)

			resp, e := commander.Exec(&commander.Opt{
				Name: "ip",
				Args: []string{
					"netns", "exec", n.Namespace, "dig",
					"@" + settings.Hypervisor.DnsServerPrimary6,
					"app6.pritunl.com",
					"AAAA",
				},
				Timeout: 5 * time.Second,
				PipeOut: true,
				PipeErr: true,
			})
			if e != nil {
				output := ""
				if resp != nil {
					output = string(resp.Output)
				}

				logrus.WithFields(logrus.Fields{
					"instance_id": n.Virt.Id.Hex(),
					"namespace":   n.Namespace,
					"address6":    n.PublicAddress6,
					"output":      output,
				}).Warn("netconf: IPv6 network DNS lookup test failed")
				continue
			}

			resp, e = commander.Exec(&commander.Opt{
				Name: "ip",
				Args: []string{
					"netns", "exec", n.Namespace, "ping6",
					"-c", "3", "-i", "0.5", "-w", "6",
					settings.Hypervisor.Ipv6PingHost,
				},
				Timeout: 6 * time.Second,
				PipeOut: true,
				PipeErr: true,
			})
			if e != nil {
				output := ""
				if resp != nil {
					output = string(resp.Output)
				}

				logrus.WithFields(logrus.Fields{
					"instance_id": n.Virt.Id.Hex(),
					"namespace":   n.Namespace,
					"address6":    n.PublicAddress6,
					"output":      output,
				}).Warn("netconf: IPv6 network DNS lookup test failed")
				continue
			}

			break
		}
	}

	return
}

func (n *NetConf) ipArp(db *database.Database) (err error) {
	if n.NetworkMode == node.Static {
		addr := strings.Split(n.ExternalAddrCidr, "/")[0]
		iface := n.SpaceExternalIfaceMod
		if iface == "" {
			iface = n.SpaceExternalIface
		}

		_, _ = commander.Exec(&commander.Opt{
			Name: "ip",
			Args: []string{
				"netns", "exec", n.Namespace, "arping",
				"-U", "-I", iface, "-c", "3", addr,
			},
			Timeout: 6 * time.Second,
			PipeOut: true,
			PipeErr: true,
		})

		_, _ = commander.Exec(&commander.Opt{
			Name: "ip",
			Args: []string{
				"netns", "exec", n.Namespace, "arping",
				"-I", iface, "-c", "3",
				n.ExternalGatewayAddr.String(),
			},
			Timeout: 6 * time.Second,
			PipeOut: true,
			PipeErr: true,
		})
	}

	if n.NetworkMode6 == node.Static {
		addr := strings.Split(n.ExternalAddrCidr6, "/")[0]
		iface := n.SpaceExternalIfaceMod6
		if iface == "" {
			iface = n.SpaceExternalIface
		}

		_, _ = commander.Exec(&commander.Opt{
			Name: "ip",
			Args: []string{
				"netns", "exec", n.Namespace, "ndisc6",
				"-r", "3", addr, iface,
			},
			Timeout: 6 * time.Second,
			PipeOut: true,
			PipeErr: true,
		})

		if n.ExternalGatewayAddr6 != nil {
			_, _ = commander.Exec(&commander.Opt{
				Name: "ip",
				Args: []string{
					"netns", "exec", n.Namespace, "ping6",
					"-c", "3", "-i", "0.5", "-w", "6", "-I", iface,
					n.ExternalGatewayAddr6.String(),
				},
				Timeout: 8 * time.Second,
				PipeOut: true,
				PipeErr: true,
			})
		}
	}

	return
}

func (n *NetConf) ipInit6Alt(db *database.Database) (err error) {
	if n.NetworkMode6 != node.Disabled && n.NetworkMode6 != node.Cloud &&
		n.PublicAddress6 != "" && !settings.Hypervisor.NoIpv6PingInit {

		addrs, e := utils.DnsLookup(
			settings.Hypervisor.DnsServerPrimary6,
			"app6.pritunl.com",
		)
		if e != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": n.Virt.Id.Hex(),
				"namespace":   n.Namespace,
				"address6":    n.PublicAddress6,
				"error":       e,
			}).Warn("netconf: Failed to initialize IPv6 network DNS lookup")
		} else if addrs != nil && len(addrs) > 0 {
			output, e := utils.ExecCombinedOutput(
				"",
				"ip", "netns", "exec", n.Namespace,
				"ping6", "-c", "3", "-i", "0.5", "-w", "6", addrs[0],
			)
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"instance_id": n.Virt.Id.Hex(),
					"namespace":   n.Namespace,
					"address6":    n.PublicAddress6,
					"output":      output,
				}).Warn("netconf: Failed to initialize IPv6 network ping")
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"instance_id": n.Virt.Id.Hex(),
				"namespace":   n.Namespace,
				"address6":    n.PublicAddress6,
				"lookup":      addrs,
			}).Warn("netconf: Failed to initialize IPv6 network DNS lookup")
		}
	}

	return
}

func (n *NetConf) Ip(db *database.Database) (err error) {
	err = n.ipExternal(db)
	if err != nil {
		return
	}

	err = n.ipHost(db)
	if err != nil {
		return
	}

	err = n.ipNodePort(db)
	if err != nil {
		return
	}

	err = n.ipDetect(db)
	if err != nil {
		return
	}

	err = n.ipHostIptables(db)
	if err != nil {
		return
	}

	err = n.ipDatabase(db)
	if err != nil {
		return
	}

	err = n.ipArp(db)
	if err != nil {
		return
	}

	err = n.ipInit6(db)
	if err != nil {
		return
	}

	return
}
