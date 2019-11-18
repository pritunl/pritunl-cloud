package ipsec

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/interfaces"
	"github.com/pritunl/pritunl-cloud/iptables"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
)

var (
	networkStates     = map[primitive.ObjectID]bool{}
	networkStatesLock = sync.Mutex{}
)

func networkConf(db *database.Database, vc *vpc.Vpc,
	netAddr, netAddr6 string, netCidr int) (err error) {

	networkStatesLock.Lock()
	networkState := networkStates[vc.Id]
	networkStatesLock.Unlock()
	if networkState {
		return
	}

	jumboFrames := node.Self.JumboFrames
	namespace := vm.GetLinkNamespace(vc.Id, 0)
	ifaceExternalVirt := vm.GetLinkIfaceVirt(vc.Id, 0)
	ifaceInternalVirt := vm.GetLinkIfaceVirt(vc.Id, 1)
	ifaceVlan := vm.GetIfaceVlan(vc.Id, 0)
	ifaceExternal := vm.GetLinkIfaceExternal(vc.Id, 0)
	ifaceInternal := vm.GetLinkIfaceInternal(vc.Id, 0)
	pidPath := fmt.Sprintf("/var/run/dhclient-%s.pid", ifaceExternal)
	namespacePth := fmt.Sprintf("/etc/netns/%s", namespace)
	vcIdPth := fmt.Sprintf("%s/vpc.id", namespacePth)
	leasePath := paths.GetLinkLeasePath()

	logrus.WithFields(logrus.Fields{
		"vpc_id":    vc.Id.Hex(),
		"namespace": namespace,
	}).Info("ipsec: Configuring ipsec network namespace")

	zne, err := zone.Get(db, node.Self.Zone)
	if err != nil {
		return
	}

	vxlan := false
	if zne.NetworkMode == zone.VxlanVlan {
		vxlan = true
	}

	updateMtuInternal := ""
	updateMtuExternal := ""
	if jumboFrames || vxlan {
		mtuSize := 0
		if jumboFrames {
			mtuSize = settings.Hypervisor.JumboMtu
		} else {
			mtuSize = settings.Hypervisor.NormalMtu
		}

		updateMtuExternal = strconv.Itoa(mtuSize)

		if vxlan {
			mtuSize -= 50
		}

		updateMtuInternal = strconv.Itoa(mtuSize)
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns",
		"add", namespace,
	)
	if err != nil {
		return
	}

	err = utils.ExistsMkdir(namespacePth, 0755)
	if err != nil {
		return
	}

	err = utils.CreateWrite(vcIdPth, vc.Id.Hex(), 0644)
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

	if updateMtuExternal != "" {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", ifaceExternalVirt,
			"mtu", updateMtuExternal,
		)
		if err != nil {
			return
		}

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", ifaceExternal,
			"mtu", updateMtuExternal,
		)
		if err != nil {
			return
		}
	}
	if updateMtuInternal != "" {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", ifaceInternalVirt,
			"mtu", updateMtuInternal,
		)
		if err != nil {
			return
		}

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", ifaceInternal,
			"mtu", updateMtuInternal,
		)
		if err != nil {
			return
		}
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

	internalIface := interfaces.GetInternal(ifaceInternalVirt, vxlan)
	if internalIface == "" {
		err = &errortypes.NotFoundError{
			errors.New("ipsec: Failed to get internal interface"),
		}
		return
	}

	var externalIface string
	var blck *block.Block
	var staticAddr net.IP

	if node.Self.NetworkMode == node.Static {
		blck, staticAddr, externalIface, err = node.Self.GetStaticAddr(
			db, vc.Id)
		if err != nil {
			return
		}
	} else {
		externalIface = interfaces.GetExternal(ifaceExternalVirt)
	}
	if externalIface == "" {
		err = &errortypes.NotFoundError{
			errors.New("ipsec: Failed to get external interface"),
		}
		return
	}

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
		nil,
		"ip", "link", "set",
		ifaceExternalVirt, "master", externalIface,
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link", "set",
		ifaceInternalVirt, "master", internalIface,
	)
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

	iptables.Lock()
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"iptables",
		"-A", "FORWARD",
		"-i", ifaceExternal,
		"-j", "DROP",
	)
	iptables.Unlock()
	if err != nil {
		return
	}

	iptables.Lock()
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ip6tables",
		"-A", "FORWARD",
		"-i", ifaceExternal,
		"-j", "DROP",
	)
	iptables.Unlock()
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

	if updateMtuInternal != "" {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", namespace,
			"ip", "link",
			"set", "dev", ifaceVlan,
			"mtu", updateMtuInternal,
		)
		if err != nil {
			return
		}
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
		[]string{
			"File exists",
		},
		"ip", "netns", "exec", namespace,
		"ip", "link", "add",
		"br0", "type", "bridge",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ip", "link", "set",
		ifaceVlan, "master", "br0",
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

	if node.Self.NetworkMode == node.Static {
		staticGateway := blck.GetGateway()
		staticMask := blck.GetMask()
		if staticGateway == nil || staticMask == nil {
			err = &errortypes.ParseError{
				errors.New("qemu: Invalid block gateway cidr"),
			}
			return
		}

		staticSize, _ := staticMask.Size()
		staticCidr := fmt.Sprintf("%s/%d", staticAddr.String(), staticSize)

		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "netns", "exec", namespace,
			"ip", "addr",
			"add", staticCidr,
			"dev", ifaceExternal,
		)
		if err != nil {
			return
		}

		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "netns", "exec", namespace,
			"ip", "route",
			"add", "default",
			"via", staticGateway.String(),
		)
		if err != nil {
			return
		}
	} else {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", namespace,
			"dhclient",
			"-pf", pidPath,
			"-lf", leasePath,
			ifaceExternal,
		)
		if err != nil {
			return
		}
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

		for _, line := range strings.Split(ipData, "\n") {
			if !strings.Contains(line, "global") {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) > 3 {
				ipAddr := net.ParseIP(strings.Split(fields[3], "/")[0])
				if ipAddr != nil && len(ipAddr) > 0 {
					pubAddr = ipAddr.String()
				}
			}

			break
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

			fields := strings.Fields(line)
			if len(fields) > 3 {
				ipAddr := net.ParseIP(strings.Split(fields[3], "/")[0])
				if ipAddr != nil && len(ipAddr) > 0 {
					pubAddr6 = ipAddr.String()
				}
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

func networkConfClear(vcId primitive.ObjectID) (err error) {
	networkStatesLock.Lock()
	delete(networkStates, vcId)
	networkStatesLock.Unlock()

	logrus.WithFields(logrus.Fields{
		"vpc_id": vcId.Hex(),
	}).Info("ipsec: Removing ipsec network namespace")

	err = networkStopDhClient(vcId)
	if err != nil {
		return
	}

	ifaceExternalVirt := vm.GetLinkIfaceVirt(vcId, 0)
	ifaceInternalVirt := vm.GetLinkIfaceVirt(vcId, 1)

	utils.ExecCombinedOutput("", "ip", "link",
		"set", ifaceExternalVirt, "down")
	utils.ExecCombinedOutput("", "ip", "link", "del", ifaceExternalVirt)
	utils.ExecCombinedOutput("", "ip", "link",
		"set", ifaceInternalVirt, "down")
	utils.ExecCombinedOutput("", "ip", "link", "del", ifaceInternalVirt)

	interfaces.RemoveVirtIface(ifaceExternalVirt)
	interfaces.RemoveVirtIface(ifaceInternalVirt)

	return
}

func getAddr(vc *vpc.Vpc) (addr, addr6 string, err error) {
	namespace := vm.GetLinkNamespace(vc.Id, 0)
	ifaceExternal := vm.GetLinkIfaceExternal(vc.Id, 0)

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

	for _, line := range strings.Split(ipData, "\n") {
		if !strings.Contains(line, "global") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) > 3 {
			ipAddr := net.ParseIP(strings.Split(fields[3], "/")[0])
			if ipAddr != nil && len(ipAddr) > 0 {
				addr = ipAddr.String()
			}
		}

		break
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

		fields := strings.Fields(line)
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
