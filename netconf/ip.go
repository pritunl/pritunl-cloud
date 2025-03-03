package netconf

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/iproute"
	"github.com/pritunl/pritunl-cloud/iptables"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/store"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

func (n *NetConf) ipStartDhClient(db *database.Database) (err error) {
	if len(n.Virt.NetworkAdapters) == 0 {
		err = &errortypes.NotFoundError{
			errors.New("qemu: Missing network interfaces"),
		}
		return
	}

	pid := ""
	pidData, _ := ioutil.ReadFile(n.DhcpPidPath)
	if pidData != nil {
		pid = strings.TrimSpace(string(pidData))
	}

	if pid != "" {
		_, _ = utils.ExecCombinedOutput("", "kill", pid)
	}

	_ = utils.RemoveAll(n.DhcpPidPath)

	pid = ""
	pidData, _ = ioutil.ReadFile(n.Dhcp6PidPath)
	if pidData != nil {
		pid = strings.TrimSpace(string(pidData))
	}

	if pid != "" {
		_, _ = utils.ExecCombinedOutput("", "kill", pid)
	}

	_ = utils.RemoveAll(n.Dhcp6PidPath)

	return
}

func (n *NetConf) ipExternal(db *database.Database) (err error) {
	if n.NetworkMode == node.Static {
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
		)
		if err != nil {
			return
		}
	} else if n.NetworkMode == node.Dhcp {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"unshare", "--mount",
			"sh", "-c", fmt.Sprintf(
				"mount -t tmpfs none /etc && dhclient -4 -pf %s -lf %s %s",
				n.DhcpPidPath, n.DhcpLeasePath, n.SpaceExternalIface),
		)
		if err != nil {
			return
		}
	}

	if n.NetworkMode6 == node.Static {
		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "netns", "exec", n.Namespace,
			"ip", "-6", "addr",
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
	} else if n.NetworkMode6 == node.Dhcp || n.NetworkMode6 == node.DhcpSlaac {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"unshare", "--mount",
			"sh", "-c", fmt.Sprintf(
				"mount -t tmpfs none /etc && dhclient -6 -pf %s -lf %s %s",
				n.Dhcp6PidPath, n.Dhcp6LeasePath, n.SpaceExternalIface),
		)
		if err != nil {
			return
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
	time.Sleep(2 * time.Second)

	ipTimeout := settings.Hypervisor.IpTimeout * 4
	ipTimeout6 := settings.Hypervisor.IpTimeout6 * 4

	pubAddr := ""
	pubAddr6 := ""
	if n.NetworkMode != node.Disabled && n.NetworkMode != node.Oracle {
		for i := 0; i < ipTimeout; i++ {
			address, address6, e := iproute.AddressGetIface(
				n.Namespace, n.SpaceExternalIface)
			if e != nil {
				err = e
				return
			}

			if n.NetworkMode6 != node.Disabled &&
				n.NetworkMode6 != node.Oracle {

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
		n.NetworkMode6 != node.Oracle {

		for i := 0; i < ipTimeout6; i++ {
			_, address6, e := iproute.AddressGetIface(
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
	n.PublicAddress6 = pubAddr6

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

		err = coll.UpdateId(n.Virt.Deployment, &bson.M{
			"$set": &bson.M{
				"instance_data.private_ips": []string{
					n.InternalAddr.String(),
				},
				"instance_data.private_ips6": []string{
					n.InternalAddr6.String(),
				},
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
	if n.NetworkMode6 != node.Disabled && n.NetworkMode6 != node.Oracle &&
		n.PublicAddress6 != "" && !settings.Hypervisor.NoIpv6PingInit {

		_, e := utils.DnsLookup("2001:4860:4860::8888", "app6.pritunl.com")
		if e != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": n.Virt.Id.Hex(),
				"namespace":   n.Namespace,
				"address6":    n.PublicAddress6,
				"error":       e,
			}).Warn("netconf: Failed to initialize IPv6 network DNS lookup")
		}

		output, e := utils.ExecCombinedOutput(
			"",
			"ip", "netns", "exec", n.Namespace,
			"ping6", "-c", "3", "-i", "0.5", "-w", "6",
			"2001:4860:4860::8888",
		)
		if e != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": n.Virt.Id.Hex(),
				"namespace":   n.Namespace,
				"address6":    n.PublicAddress6,
				"output":      output,
			}).Warn("netconf: Failed to initialize IPv6 network ping")
		}
	}

	return
}

func (n *NetConf) ipInit6Alt(db *database.Database) (err error) {
	if n.NetworkMode6 != node.Disabled && n.NetworkMode6 != node.Oracle &&
		n.PublicAddress6 != "" && !settings.Hypervisor.NoIpv6PingInit {

		addrs, e := utils.DnsLookup("2001:4860:4860::8888", "app6.pritunl.com")
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
	err = n.ipStartDhClient(db)
	if err != nil {
		return
	}

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

	err = n.ipInit6(db)
	if err != nil {
		return
	}

	return
}
