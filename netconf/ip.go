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
	"github.com/pritunl/pritunl-cloud/store"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

func (n *NetConf) ipStartDhClient(db *database.Database) (err error) {
	if len(n.Virt.NetworkAdapters) == 0 {
		err = &errortypes.NotFoundError{
			errors.New("qemu: Missing network interfaces"),
		}
		return
	}

	ifaceExternal := vm.GetIfaceExternal(n.Virt.Id, 0)
	pidPath := fmt.Sprintf("/var/run/dhclient-%s.pid", ifaceExternal)

	pid := ""
	pidData, _ := ioutil.ReadFile(pidPath)
	if pidData != nil {
		pid = strings.TrimSpace(string(pidData))
	}

	if pid != "" {
		_, _ = utils.ExecCombinedOutput("", "kill", pid)
	}

	_ = utils.RemoveAll(pidPath)

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
			"dhclient",
			"-pf", n.DhcpPidPath,
			"-lf", n.DhcpLeasePath,
			n.SpaceExternalIface,
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
			"dev", n.SpaceExternalIface6,
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
				"dev", n.SpaceExternalIface6,
			)
			if err != nil {
				return
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

func (n *NetConf) ipDetect(db *database.Database) (err error) {
	time.Sleep(2 * time.Second)
	start := time.Now()

	pubAddr := ""
	pubAddr6 := ""
	if n.NetworkMode != node.Disabled && n.NetworkMode != node.Oracle {
		for i := 0; i < 120; i++ {
			address, address6, e := iproute.AddressGetIface(
				n.Namespace, n.SpaceExternalIface)
			if e != nil {
				err = e
				return
			}

			if n.NetworkMode6 != node.Disabled &&
				n.SpaceExternalIface == n.SpaceExternalIface6 &&
				n.NetworkMode6 != node.Oracle {

				if (address != nil && address6 != nil) ||
					time.Since(start) > 8*time.Second {

					if address != nil {
						pubAddr = address.Local
					}
					if address6 != nil {
						pubAddr6 = address6.Local
					}
					break
				}
			} else if address != nil || time.Since(start) > 8*time.Second {
				if address != nil {
					pubAddr = address.Local
				}
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

		if n.NetworkMode6 != node.Disabled &&
			n.SpaceExternalIface == n.SpaceExternalIface6 &&
			n.NetworkMode6 != node.Oracle {

			if pubAddr6 == "" {
				logrus.WithFields(logrus.Fields{
					"instance_id":   n.Virt.Id.Hex(),
					"net_namespace": n.Namespace,
				}).Warning("qemu: Instance missing IPv6 address")
			}
		}
	}

	if n.NetworkMode6 != node.Disabled &&
		n.SpaceExternalIface != n.SpaceExternalIface6 &&
		n.NetworkMode6 != node.Oracle {

		for i := 0; i < 120; i++ {
			_, address6, e := iproute.AddressGetIface(
				n.Namespace, n.SpaceExternalIface6)
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
			"-j", "SNAT",
			"--to", n.HostAddr.String(),
		)
		iptables.Unlock()
		if err != nil {
			return
		}

		iptables.Lock()
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"iptables", "-t", "nat",
			"-A", "POSTROUTING",
			"-s", n.InternalAddr.String()+"/32",
			"-o", n.SpaceHostIface,
			"-j", "MASQUERADE",
		)
		iptables.Unlock()
		if err != nil {
			return
		}

		iptables.Lock()
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", n.Namespace,
			"iptables", "-t", "nat",
			"-A", "PREROUTING",
			"-d", n.HostAddr.String()+"/32",
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

	hostIps := []string{}
	if n.HostAddr != nil {
		hostIps = append(hostIps, n.HostAddr.String())
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

	return
}
