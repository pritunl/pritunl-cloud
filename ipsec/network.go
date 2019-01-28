package ipsec

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/interfaces"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	networkStates     = map[primitive.ObjectID]bool{}
	networkStatesLock = sync.Mutex{}
	networkLock       = utils.NewMultiTimeoutLock(2 * time.Minute)
)

func networkConf(vc *vpc.Vpc,
	netAddr, netAddr6 string, netCidr int) (err error) {

	lockId := networkLock.Lock(vc.Id.Hex())
	defer networkLock.Unlock(vc.Id.Hex(), lockId)

	networkStatesLock.Lock()
	networkState := networkStates[vc.Id]
	networkStatesLock.Unlock()
	if networkState {
		return
	}

	logrus.WithFields(logrus.Fields{
		"vpc_id": vc.Id.Hex(),
	}).Info("ipsec: Configuring IPsec network namespace")

	namespace := vm.GetLinkNamespace(vc.Id, 0)
	ifaceExternalVirt := vm.GetLinkIfaceVirt(vc.Id, 0)
	ifaceInternalVirt := vm.GetLinkIfaceVirt(vc.Id, 1)
	ifaceVlan := vm.GetIfaceVlan(vc.Id, 0)
	ifaceExternal := vm.GetLinkIfaceExternal(vc.Id, 0)
	ifaceInternal := vm.GetLinkIfaceInternal(vc.Id, 0)
	pidPath := fmt.Sprintf("/var/run/dhclient-%s.pid", ifaceExternal)

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns",
		"add", namespace,
	)
	if err != nil {
		return
	}

	utils.ExecCombinedOutput("", "ip", "link",
		"set", ifaceExternalVirt, "down")
	utils.ExecCombinedOutput("", "ip", "link", "del", ifaceExternalVirt)
	utils.ExecCombinedOutput("", "ip", "link",
		"set", ifaceInternalVirt, "down")
	utils.ExecCombinedOutput("", "ip", "link", "del", ifaceInternalVirt)

	interfaces.RemoveVirtIface(ifaceExternalVirt)
	interfaces.RemoveVirtIface(ifaceInternalVirt)

	macAddrExternal := vm.GetMacAddrExternal(vc.Id, node.Self.Id)
	macAddrInternal := vm.GetMacAddrInternal(vc.Id, node.Self.Id)

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link",
		"add", ifaceExternalVirt,
		"type", "veth",
		"peer", "name", ifaceExternal,
		"addr", macAddrExternal,
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link",
		"add", ifaceInternalVirt,
		"type", "veth",
		"peer", "name", ifaceInternal,
		"addr", macAddrInternal,
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link",
		"set", "dev", ifaceExternalVirt, "up",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link",
		"set", "dev", ifaceInternalVirt, "up",
	)
	if err != nil {
		return
	}

	externalIface := interfaces.GetExternal(ifaceExternalVirt)
	internalIface := interfaces.GetInternal(ifaceInternalVirt)

	_, err = utils.ExecCombinedOutputLogged(
		nil, "sysctl", "-w",
		fmt.Sprintf("net.ipv6.conf.%s.accept_ra=2", externalIface),
	)
	if err != nil {
		return
	}
	if internalIface != externalIface {
		_, err = utils.ExecCombinedOutputLogged(
			nil, "sysctl", "-w",
			fmt.Sprintf("net.ipv6.conf.%s.accept_ra=2", internalIface),
		)
		if err != nil {
			return
		}
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"already a member of a bridge"},
		"brctl", "addif", externalIface, ifaceExternalVirt)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"already a member of a bridge"},
		"brctl", "addif", internalIface, ifaceInternalVirt)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "link",
		"set", "dev", ifaceExternal,
		"netns", namespace,
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "link",
		"set", "dev", ifaceInternal,
		"netns", namespace,
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"sysctl", "-w", "net.ipv6.conf.all.accept_ra=0",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"sysctl", "-w", "net.ipv6.conf.default.accept_ra=0",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"sysctl", "-w",
		fmt.Sprintf("net.ipv6.conf.%s.accept_ra=2", ifaceExternal),
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"sysctl", "-w", "net.ipv4.ip_forward=1",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"sysctl", "-w", "net.ipv6.conf.all.forwarding=1",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"sysctl", "-w", "net.ipv6.conf.default.forwarding=1",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ip", "link",
		"set", "dev", "lo", "up",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ip", "link",
		"set", "dev", ifaceExternal, "up",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ip", "link",
		"set", "dev", ifaceInternal, "up",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns", "exec", namespace,
		"ip", "link",
		"add", "link", ifaceInternal,
		"name", ifaceVlan,
		"type", "vlan",
		"id", strconv.Itoa(vc.VpcId),
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ip", "link",
		"set", "dev", ifaceVlan, "up",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"already exists"},
		"ip", "netns", "exec", namespace,
		"brctl", "addbr", "br0",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"already a member of a bridge"},
		"ip", "netns", "exec", namespace,
		"brctl", "addif", "br0", ifaceVlan,
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns", "exec", namespace,
		"ip", "addr",
		"add", fmt.Sprintf("%s/%d", netAddr, netCidr),
		"dev", "br0",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns", "exec", namespace,
		"ip", "-6", "addr",
		"add", netAddr6+"/64",
		"dev", "br0",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ip", "link",
		"set", "dev", "br0", "up",
	)
	if err != nil {
		return
	}

	networkStopDhClient(vc.Id)

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"dhclient", "-pf", pidPath,
		ifaceExternal,
	)
	if err != nil {
		return
	}

	time.Sleep(2 * time.Second)
	start := time.Now()

	pubAddr := ""
	pubAddr6 := ""
	for i := 0; i < 60; i++ {
		ipData, e := utils.ExecCombinedOutputLogged(
			[]string{
				"No such file or directory",
				"does not exist",
			},
			"ip", "netns", "exec", namespace,
			"ip", "-f", "inet", "-o", "addr",
			"show", "dev", ifaceExternal,
		)
		if e != nil {
			err = e
			return
		}

		fields := strings.Fields(ipData)
		if len(fields) > 3 {
			ipAddr := net.ParseIP(strings.Split(fields[3], "/")[0])
			pubAddr = ipAddr.String()
		}

		ipData, e = utils.ExecCombinedOutputLogged(
			[]string{
				"No such file or directory",
				"does not exist",
			},
			"ip", "netns", "exec", namespace,
			"ip", "-f", "inet6", "-o", "addr",
			"show", "dev", ifaceExternal,
		)
		if e != nil {
			err = e
			return
		}

		for _, line := range strings.Split(ipData, "\n") {
			if !strings.Contains(line, "global") {
				continue
			}

			fields = strings.Fields(ipData)
			if len(fields) > 3 {
				ipAddr := net.ParseIP(strings.Split(fields[3], "/")[0])
				pubAddr6 = ipAddr.String()
			}

			break
		}

		if pubAddr != "" && (pubAddr6 != "" ||
			time.Since(start) > 8*time.Second) {

			break
		}

		time.Sleep(250 * time.Millisecond)
	}

	networkStatesLock.Lock()
	networkStates[vc.Id] = true
	networkStatesLock.Unlock()

	return
}

func syncAddr(vc *vpc.Vpc) (addr, addr6 string, err error) {
	namespace := vm.GetLinkNamespace(vc.Id, 0)
	ifaceExternal := vm.GetLinkIfaceExternal(vc.Id, 0)

	ipData, err := utils.ExecCombinedOutputLogged(
		[]string{
			"No such file or directory",
			"does not exist",
			"setting the network namespace",
		},
		"ip", "netns", "exec", namespace,
		"ip", "-f", "inet", "-o", "addr",
		"show", "dev", ifaceExternal,
	)
	if err != nil {
		return
	}

	fields := strings.Fields(ipData)
	if len(fields) > 3 {
		ipAddr := net.ParseIP(strings.Split(fields[3], "/")[0])
		if ipAddr != nil && len(ipAddr) > 0 {
			addr = ipAddr.String()
		}
	}

	ipData, err = utils.ExecCombinedOutputLogged(
		[]string{
			"No such file or directory",
			"does not exist",
			"setting the network namespace",
		},
		"ip", "netns", "exec", namespace,
		"ip", "-f", "inet6", "-o", "addr",
		"show", "dev", ifaceExternal,
	)
	if err != nil {
		return
	}

	for _, line := range strings.Split(ipData, "\n") {
		if !strings.Contains(line, "global") {
			continue
		}

		fields = strings.Fields(ipData)
		if len(fields) > 3 {
			ipAddr := net.ParseIP(strings.Split(fields[3], "/")[0])
			if ipAddr != nil && len(ipAddr) > 0 {
				addr6 = ipAddr.String()
			}
		}

		break
	}

	return
}
