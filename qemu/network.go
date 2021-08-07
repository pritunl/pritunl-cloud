package qemu

import (
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/interfaces"
	"github.com/pritunl/pritunl-cloud/iproute"
	"github.com/pritunl/pritunl-cloud/iptables"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/store"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
	"github.com/sirupsen/logrus"
)

func networkStopDhClient(virt *vm.VirtualMachine) (err error) {
	if len(virt.NetworkAdapters) == 0 {
		err = &errortypes.NotFoundError{
			errors.New("qemu: Missing network interfaces"),
		}
		return
	}

	ifaceExternal := vm.GetIfaceExternal(virt.Id, 0)
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

func NetworkConfClear(virt *vm.VirtualMachine) (err error) {
	if len(virt.NetworkAdapters) == 0 {
		err = &errortypes.NotFoundError{
			errors.New("qemu: Missing network interfaces"),
		}
		return
	}

	err = networkStopDhClient(virt)
	if err != nil {
		return
	}

	ifaceExternalVirt := vm.GetIfaceVirt(virt.Id, 0)
	ifaceExternalVirt6 := vm.GetIfaceVirt(virt.Id, 3)
	ifaceInternalVirt := vm.GetIfaceVirt(virt.Id, 1)
	ifaceHostVirt := vm.GetIfaceVirt(virt.Id, 2)

	_, _ = utils.ExecCombinedOutput(
		"", "ip", "link", "set", ifaceExternalVirt, "down")
	_, _ = utils.ExecCombinedOutput(
		"", "ip", "link", "set", ifaceExternalVirt6, "down")
	_, _ = utils.ExecCombinedOutput(
		"", "ip", "link", "del", ifaceExternalVirt)
	_, _ = utils.ExecCombinedOutput(
		"", "ip", "link", "del", ifaceExternalVirt6)
	_, _ = utils.ExecCombinedOutput(
		"", "ip", "link", "set", ifaceInternalVirt, "down")
	_, _ = utils.ExecCombinedOutput(
		"", "ip", "link", "del", ifaceInternalVirt)
	_, _ = utils.ExecCombinedOutput(
		"", "ip", "link", "set", ifaceHostVirt, "down")
	_, _ = utils.ExecCombinedOutput(
		"", "ip", "link", "del", ifaceHostVirt)

	interfaces.RemoveVirtIface(ifaceExternalVirt)
	interfaces.RemoveVirtIface(ifaceExternalVirt6)
	interfaces.RemoveVirtIface(ifaceInternalVirt)

	store.RemAddress(virt.Id)
	store.RemRoutes(virt.Id)

	return
}

func NetworkConf(db *database.Database,
	virt *vm.VirtualMachine) (err error) {

	ifaceNames := set.NewSet()

	if len(virt.NetworkAdapters) == 0 {
		err = &errortypes.NotFoundError{
			errors.New("qemu: Missing network interfaces"),
		}
		return
	}

	for i := range virt.NetworkAdapters {
		ifaceNames.Add(vm.GetIface(virt.Id, i))
	}

	for i := 0; i < 100; i++ {
		ifaces, e := net.Interfaces()
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "qemu: Failed to get network interfaces"),
			}
			return
		}

		for _, iface := range ifaces {
			if ifaceNames.Contains(iface.Name) {
				ifaceNames.Remove(iface.Name)
			}
		}

		if ifaceNames.Len() == 0 {
			break
		}

		time.Sleep(250 * time.Millisecond)
	}

	if ifaceNames.Len() != 0 {
		err = &errortypes.ReadError{
			errors.New("qemu: Failed to find network interfaces"),
		}
		return
	}

	zne, err := zone.Get(db, node.Self.Zone)
	if err != nil {
		return
	}

	vxlan := false
	if zne.NetworkMode == zone.VxlanVlan {
		vxlan = true
	}

	nodeNetworkMode := node.Self.NetworkMode
	if nodeNetworkMode == "" {
		nodeNetworkMode = node.Dhcp
	}
	nodeNetworkMode6 := node.Self.NetworkMode6
	jumboFrames := node.Self.JumboFrames
	iface := vm.GetIface(virt.Id, 0)
	ifaceExternalVirt := vm.GetIfaceVirt(virt.Id, 0)
	ifaceInternalVirt := vm.GetIfaceVirt(virt.Id, 1)
	ifaceHostVirt := vm.GetIfaceVirt(virt.Id, 2)
	ifaceExternal := vm.GetIfaceExternal(virt.Id, 0)
	ifaceInternal := vm.GetIfaceInternal(virt.Id, 0)
	ifaceHost := vm.GetIfaceHost(virt.Id, 0)
	ifaceVlan := vm.GetIfaceVlan(virt.Id, 0)
	namespace := vm.GetNamespace(virt.Id, 0)
	pidPath := fmt.Sprintf("/var/run/dhclient-%s.pid", ifaceExternal)
	leasePath := paths.GetLeasePath(virt.Id)
	adapter := virt.NetworkAdapters[0]

	ifaceExternal6 := ifaceExternal
	ifaceExternalVirt6 := ifaceExternalVirt
	if nodeNetworkMode6 != "" && (nodeNetworkMode != nodeNetworkMode6 ||
		(nodeNetworkMode6 == node.Static)) {

		ifaceExternal6 = vm.GetIfaceExternal(virt.Id, 1)
		ifaceExternalVirt6 = vm.GetIfaceVirt(virt.Id, 3)
	}

	externalNetwork := true
	if virt.NoPublicAddress || nodeNetworkMode == node.Internal {
		externalNetwork = false
	}

	externalNetwork6 := externalNetwork
	if !virt.NoPublicAddress && nodeNetworkMode6 != "" {
		externalNetwork6 = true
	}

	hostNetwork := false
	if !virt.NoHostAddress && !node.Self.HostBlock.IsZero() {
		hostNetwork = true
	}

	updateMtuInternal := ""
	updateMtuExternal := ""
	updateMtuInstance := ""
	updateMtuHost := ""
	if jumboFrames || vxlan {
		mtuSize := 0
		if jumboFrames {
			mtuSize = settings.Hypervisor.JumboMtu
		} else {
			mtuSize = settings.Hypervisor.NormalMtu
		}

		updateMtuExternal = strconv.Itoa(mtuSize)
		updateMtuHost = strconv.Itoa(mtuSize)

		if vxlan {
			mtuSize -= 50
		}

		updateMtuInternal = strconv.Itoa(mtuSize)

		if vxlan {
			mtuSize -= 4
		}

		updateMtuInstance = strconv.Itoa(mtuSize)
	}

	err = utils.ExistsMkdir(paths.GetLeasesPath(), 0755)
	if err != nil {
		return
	}

	vc, err := vpc.Get(db, adapter.Vpc)
	if err != nil {
		return
	}

	vcNet, err := vc.GetNetwork()
	if err != nil {
		return
	}

	addr, gatewayAddr, err := vc.GetIp(db, adapter.Subnet, virt.Id)
	if err != nil {
		return
	}

	addr6 := vc.GetIp6(addr)
	gatewayAddr6 := vc.GetIp6(gatewayAddr)

	cidr, _ := vcNet.Mask.Size()
	gatewayCidr := fmt.Sprintf("%s/%d", gatewayAddr.String(), cidr)

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns",
		"add", namespace,
	)
	if err != nil {
		return
	}

	_, _ = utils.ExecCombinedOutput(
		"", "ip", "link", "set", ifaceExternalVirt, "down")
	_, _ = utils.ExecCombinedOutput(
		"", "ip", "link", "del", ifaceExternalVirt)
	_, _ = utils.ExecCombinedOutput(
		"", "ip", "link", "set", ifaceInternalVirt, "down")
	_, _ = utils.ExecCombinedOutput(
		"", "ip", "link", "del", ifaceInternalVirt)
	_, _ = utils.ExecCombinedOutput(
		"", "ip", "link", "set", ifaceHostVirt, "down")
	_, _ = utils.ExecCombinedOutput(
		"", "ip", "link", "del", ifaceHostVirt)

	interfaces.RemoveVirtIface(ifaceExternalVirt)
	interfaces.RemoveVirtIface(ifaceInternalVirt)

	macAddrExternal := vm.GetMacAddrExternal(virt.Id, vc.Id)
	macAddrInternal := vm.GetMacAddrInternal(virt.Id, vc.Id)
	macAddrHost := vm.GetMacAddrHost(virt.Id, vc.Id)

	macAddrExternal6 := macAddrExternal
	if ifaceExternal != ifaceExternal6 {
		macAddrExternal6 = vm.GetMacAddrExternal6(virt.Id, vc.Id)
	}

	if externalNetwork || (externalNetwork6 &&
		ifaceExternal == ifaceExternal6) {

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
	}
	if externalNetwork6 && ifaceExternal != ifaceExternal6 {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"add", ifaceExternalVirt6,
			"type", "veth",
			"peer", "name", ifaceExternal6,
			"addr", macAddrExternal6,
		)
		if err != nil {
			return
		}
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

	if hostNetwork {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"add", ifaceHostVirt,
			"type", "veth",
			"peer", "name", ifaceHost,
			"addr", macAddrHost,
		)
		if err != nil {
			return
		}
	}

	if (externalNetwork || (externalNetwork6 &&
		ifaceExternal == ifaceExternal6)) && updateMtuExternal != "" {

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
	if externalNetwork6 && ifaceExternal != ifaceExternal6 &&
		updateMtuExternal != "" {

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", ifaceExternalVirt6,
			"mtu", updateMtuExternal,
		)
		if err != nil {
			return
		}

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", ifaceExternal6,
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

	if hostNetwork && updateMtuHost != "" {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", ifaceHostVirt,
			"mtu", updateMtuHost,
		)
		if err != nil {
			return
		}

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", ifaceHost,
			"mtu", updateMtuHost,
		)
		if err != nil {
			return
		}
	}

	if externalNetwork || (externalNetwork6 &&
		ifaceExternal == ifaceExternal6) {

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", ifaceExternalVirt, "up",
		)
		if err != nil {
			return
		}
	}
	if externalNetwork6 && ifaceExternal != ifaceExternal6 {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", ifaceExternalVirt6, "up",
		)
		if err != nil {
			return
		}
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link",
		"set", "dev", ifaceInternalVirt, "up",
	)
	if err != nil {
		return
	}

	if hostNetwork {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link",
			"set", "dev", ifaceHostVirt, "up",
		)
		if err != nil {
			return
		}
	}

	internalIface := interfaces.GetInternal(ifaceInternalVirt, vxlan)
	if internalIface == "" {
		err = &errortypes.NotFoundError{
			errors.New("qemu: Failed to get internal interface"),
		}
		return
	}

	var externalIface string
	var blck *block.Block
	var staticAddr net.IP

	if externalNetwork {
		if nodeNetworkMode == node.Static {
			blck, staticAddr, externalIface, err = node.Self.GetStaticAddr(
				db, virt.Id)
			if err != nil {
				return
			}
		} else if nodeNetworkMode == node.Dhcp {
			externalIface = interfaces.GetExternal(ifaceExternalVirt)
		}
		if externalIface == "" {
			err = &errortypes.NotFoundError{
				errors.New("qemu: Failed to get external interface"),
			}
			return
		}
	}

	var externalIface6 string
	var staticAddr6 net.IP
	var staticCidr6 int
	var blck6 *block.Block

	if externalNetwork6 {
		if nodeNetworkMode6 == node.Static {
			blck6, staticAddr6, staticCidr6, externalIface6,
				err = node.Self.GetStaticAddr6(db, virt.Id, vc.VpcId)
			if err != nil {
				return
			}
		} else if nodeNetworkMode6 == node.Dhcp {
			if nodeNetworkMode == node.Dhcp {
				externalIface6 = externalIface
			} else {
				externalIface6 = interfaces.GetExternal(ifaceExternalVirt6)
			}
		} else {
			externalIface6 = externalIface
		}
		if externalIface6 == "" {
			err = &errortypes.NotFoundError{
				errors.New("qemu: Failed to get external interface6"),
			}
			return
		}
	}

	hostIface := settings.Hypervisor.HostNetworkName
	var hostBlck *block.Block
	var hostStaticAddr net.IP
	if hostNetwork {
		hostBlck, hostStaticAddr, err = node.Self.GetStaticHostAddr(
			db, virt.Id)
		if err != nil {
			return
		}
	}

	if externalNetwork6 {
		_, err = utils.ExecCombinedOutputLogged(
			nil, "sysctl", "-w",
			fmt.Sprintf("net.ipv6.conf.%s.accept_ra=2", externalIface6),
		)
		if err != nil {
			return
		}
	}
	if !externalNetwork6 || internalIface != externalIface6 {
		_, err = utils.ExecCombinedOutputLogged(
			nil, "sysctl", "-w",
			fmt.Sprintf("net.ipv6.conf.%s.accept_ra=2", internalIface),
		)
		if err != nil {
			return
		}
	}

	if externalNetwork || (externalNetwork6 &&
		ifaceExternalVirt == ifaceExternalVirt6) {

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link", "set",
			ifaceExternalVirt, "master", externalIface,
		)
		if err != nil {
			return
		}
	}
	if externalNetwork6 && ifaceExternalVirt != ifaceExternalVirt6 {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link", "set",
			ifaceExternalVirt6, "master", externalIface6,
		)
		if err != nil {
			return
		}
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link", "set",
		ifaceInternalVirt, "master", internalIface,
	)
	if err != nil {
		return
	}

	if hostNetwork {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "link", "set",
			ifaceHostVirt, "master", hostIface,
		)
		if err != nil {
			return
		}
	}

	if externalNetwork || (externalNetwork6 &&
		ifaceExternal == ifaceExternal6) {

		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "link",
			"set", "dev", ifaceExternal,
			"netns", namespace,
		)
		if err != nil {
			return
		}
	}
	if externalNetwork6 && ifaceExternal != ifaceExternal6 {
		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "link",
			"set", "dev", ifaceExternal6,
			"netns", namespace,
		)
		if err != nil {
			return
		}
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

	if hostNetwork {
		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "link",
			"set", "dev", ifaceHost,
			"netns", namespace,
		)
		if err != nil {
			return
		}
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

	if externalNetwork6 {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", namespace,
			"sysctl", "-w",
			fmt.Sprintf("net.ipv6.conf.%s.accept_ra=2", ifaceExternal6),
		)
		if err != nil {
			return
		}
	}

	if externalNetwork {
		iptables.Lock()
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", namespace,
			"iptables",
			"-I", "FORWARD", "1",
			"!", "-d", addr.String()+"/32",
			"-i", ifaceExternal,
			"-j", "DROP",
		)
		iptables.Unlock()
		if err != nil {
			return
		}
	}
	if externalNetwork6 {
		iptables.Lock()
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", namespace,
			"ip6tables",
			"-I", "FORWARD", "1",
			"!", "-d", addr6.String()+"/128",
			"-i", ifaceExternal6,
			"-j", "DROP",
		)
		iptables.Unlock()
		if err != nil {
			return
		}
	}

	if hostNetwork {
		iptables.Lock()
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", namespace,
			"iptables",
			"-I", "FORWARD", "1",
			"!", "-d", addr.String()+"/32",
			"-i", ifaceHost,
			"-j", "DROP",
		)
		iptables.Unlock()
		if err != nil {
			return
		}
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
		[]string{"File exists"},
		"ip", "link",
		"set", "dev", iface,
		"netns", namespace,
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

	if externalNetwork || (externalNetwork6 &&
		ifaceExternal == ifaceExternal6) {

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", namespace,
			"ip", "link",
			"set", "dev", ifaceExternal, "up",
		)
		if err != nil {
			return
		}
	}
	if externalNetwork6 && ifaceExternal != ifaceExternal6 {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", namespace,
			"ip", "link",
			"set", "dev", ifaceExternal6, "up",
		)
		if err != nil {
			return
		}
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

	if hostNetwork {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", namespace,
			"ip", "link",
			"set", "dev", ifaceHost, "up",
		)
		if err != nil {
			return
		}
	}

	if updateMtuInstance != "" {
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", namespace,
			"ip", "link",
			"set", "dev", iface,
			"mtu", updateMtuInstance,
		)
		if err != nil {
			return
		}
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ip", "link",
		"set", "dev", iface, "up",
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

	err = iproute.BridgeAdd(namespace, "br0")
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
		nil,
		"ip", "netns", "exec", namespace,
		"ip", "link", "set",
		iface, "master", "br0",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns", "exec", namespace,
		"ip", "addr",
		"add", gatewayCidr,
		"dev", "br0",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		[]string{"File exists"},
		"ip", "netns", "exec", namespace,
		"ip", "-6", "addr",
		"add", gatewayAddr6.String()+"/64",
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

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"bridge", "link",
		"set", "dev", iface, "hairpin", "on",
	)
	if err != nil {
		return
	}

	_ = networkStopDhClient(virt)

	if externalNetwork {
		if nodeNetworkMode == node.Static {
			staticGateway := blck.GetGateway()
			staticMask := blck.GetMask()
			if staticGateway == nil || staticMask == nil {
				err = &errortypes.ParseError{
					errors.New("qemu: Invalid block gateway cidr"),
				}
				return
			}

			staticSize, _ := staticMask.Size()
			staticCidr := fmt.Sprintf(
				"%s/%d", staticAddr.String(), staticSize)

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
	}
	if externalNetwork6 && nodeNetworkMode6 == node.Static {
		staticCidr6 := fmt.Sprintf(
			"%s/%d", staticAddr6.String(), staticCidr6)

		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "netns", "exec", namespace,
			"ip", "-6", "addr",
			"add", staticCidr6,
			"dev", ifaceExternal6,
		)
		if err != nil {
			return
		}

		gateway6 := blck6.GetGateway6()
		if gateway6 == nil {
			err = &errortypes.ParseError{
				errors.New("qemu: Invalid block gateway6"),
			}
			return
		}

		if gateway6 != nil {
			_, err = utils.ExecCombinedOutputLogged(
				[]string{"File exists"},
				"ip", "netns", "exec", namespace,
				"ip", "-6", "route",
				"add", "default",
				"via", gateway6.String(),
				"dev", ifaceExternal6,
			)
			if err != nil {
				return
			}
		}
	}

	if hostNetwork {
		hostStaticGateway := hostBlck.GetGateway()
		hostStaticMask := hostBlck.GetMask()
		if hostStaticGateway == nil || hostStaticMask == nil {
			err = &errortypes.ParseError{
				errors.New("qemu: Invalid block gateway cidr"),
			}
			return
		}

		hostStaticSize, _ := hostStaticMask.Size()
		hostStaticCidr := fmt.Sprintf(
			"%s/%d", hostStaticAddr.String(), hostStaticSize)

		_, err = utils.ExecCombinedOutputLogged(
			[]string{"File exists"},
			"ip", "netns", "exec", namespace,
			"ip", "addr",
			"add", hostStaticCidr,
			"dev", ifaceHost,
		)
		if err != nil {
			return
		}

		if !externalNetwork {
			_, err = utils.ExecCombinedOutputLogged(
				[]string{"File exists"},
				"ip", "netns", "exec", namespace,
				"ip", "route",
				"add", "default",
				"via", hostStaticGateway.String(),
			)
			if err != nil {
				return
			}
		}

	}

	time.Sleep(2 * time.Second)
	start := time.Now()

	pubAddr := ""
	pubAddr6 := ""
	if externalNetwork || (externalNetwork6 &&
		ifaceExternal == ifaceExternal6) {

		for i := 0; i < 60; i++ {
			address, address6, e := iproute.AddressGetIface(
				namespace, ifaceExternal)
			if e != nil {
				err = e
				return
			}

			if address != nil && (ifaceExternal != ifaceExternal6 ||
				address6 != nil || time.Since(start) > 8*time.Second) {

				pubAddr = address.Local
				if address6 != nil {
					pubAddr6 = address6.Local
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

		//iptables.Lock()
		//_, err = utils.ExecCombinedOutputLogged(
		//	nil,
		//	"ip", "netns", "exec", namespace,
		//	"iptables", "-t", "nat",
		//	"-A", "POSTROUTING",
		//	"-s", addr.String()+"/32",
		//	"-o", ifaceExternal,
		//	"-m", "comment",
		//	"--comment", "pritunl_cloud_nat",
		//	"-j", "MASQUERADE",
		//)
		//iptables.Unlock()
		//if err != nil {
		//	return
		//}
		//
		//iptables.Lock()
		//_, err = utils.ExecCombinedOutputLogged(
		//	nil,
		//	"ip", "netns", "exec", namespace,
		//	"iptables", "-t", "nat",
		//	"-A", "PREROUTING",
		//	"-d", pubAddr,
		//	"-m", "comment",
		//	"--comment", "pritunl_cloud_nat",
		//	"-j", "DNAT",
		//	"--to-destination", addr.String(),
		//)
		//iptables.Unlock()
		//if err != nil {
		//	return
		//}

		if externalNetwork6 && ifaceExternal == ifaceExternal6 {
			if pubAddr6 != "" {
				//iptables.Lock()
				//_, err = utils.ExecCombinedOutputLogged(
				//	nil,
				//	"ip", "netns", "exec", namespace,
				//	"ip6tables", "-t", "nat",
				//	"-A", "POSTROUTING",
				//	"-s", addr6.String()+"/128",
				//	"-o", ifaceExternal,
				//	"-m", "comment",
				//	"--comment", "pritunl_cloud_nat",
				//	"-j", "MASQUERADE",
				//)
				//iptables.Unlock()
				//if err != nil {
				//	return
				//}
				//
				//iptables.Lock()
				//_, err = utils.ExecCombinedOutputLogged(
				//	nil,
				//	"ip", "netns", "exec", namespace,
				//	"ip6tables", "-t", "nat",
				//	"-A", "PREROUTING",
				//	"-d", pubAddr6,
				//	"-m", "comment",
				//	"--comment", "pritunl_cloud_nat",
				//	"-j", "DNAT",
				//	"--to-destination", addr6.String(),
				//)
				//iptables.Unlock()
				//if err != nil {
				//	return
				//}

				//iptables.Lock()
				//_, err = utils.ExecCombinedOutputLogged(
				//	nil,
				//	"ip", "netns", "exec", namespace,
				//	"ip6tables", "-t", "nat",
				//	"-A", "POSTROUTING",
				//	"-s", addr6.String(),
				//	"-m", "comment",
				//	"--comment", "pritunl_cloud_nat",
				//	"-j", "SNAT",
				//	"--to-source", pubAddr6,
				//)
				//iptables.Unlock()
				//if err != nil {
				//	return
				//}
			} else {
				logrus.WithFields(logrus.Fields{
					"instance_id":   virt.Id.Hex(),
					"net_namespace": namespace,
				}).Warning("qemu: Instance missing IPv6 address")
			}
		}
	}
	if externalNetwork6 && ifaceExternal != ifaceExternal6 {
		for i := 0; i < 60; i++ {
			_, address6, e := iproute.AddressGetIface(
				namespace, ifaceExternal6)
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

		//iptables.Lock()
		//_, err = utils.ExecCombinedOutputLogged(
		//	nil,
		//	"ip", "netns", "exec", namespace,
		//	"ip6tables", "-t", "nat",
		//	"-A", "POSTROUTING",
		//	"-s", addr6.String()+"/128",
		//	"-o", ifaceExternal6,
		//	"-m", "comment",
		//	"--comment", "pritunl_cloud_nat",
		//	"-j", "MASQUERADE",
		//)
		//iptables.Unlock()
		//if err != nil {
		//	return
		//}
		//
		//iptables.Lock()
		//_, err = utils.ExecCombinedOutputLogged(
		//	nil,
		//	"ip", "netns", "exec", namespace,
		//	"ip6tables", "-t", "nat",
		//	"-A", "PREROUTING",
		//	"-d", pubAddr6,
		//	"-m", "comment",
		//	"--comment", "pritunl_cloud_nat",
		//	"-j", "DNAT",
		//	"--to-destination", addr6.String(),
		//)
		//iptables.Unlock()
		//if err != nil {
		//	return
		//}

		//iptables.Lock()
		//_, err = utils.ExecCombinedOutputLogged(
		//	nil,
		//	"ip", "netns", "exec", namespace,
		//	"ip6tables", "-t", "nat",
		//	"-A", "POSTROUTING",
		//	"-s", addr6.String(),
		//  "-m", "comment",
		//  "--comment", "pritunl_cloud_nat",
		//	"-j", "SNAT",
		//	"--to-source", pubAddr6,
		//)
		//iptables.Unlock()
		//if err != nil {
		//	return
		//}
	}

	if hostNetwork {
		iptables.Lock()
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", namespace,
			"iptables", "-t", "nat",
			"-A", "POSTROUTING",
			"-s", addr.String()+"/32",
			"-o", ifaceHost,
			"-j", "MASQUERADE",
		)
		iptables.Unlock()
		if err != nil {
			return
		}

		iptables.Lock()
		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"ip", "netns", "exec", namespace,
			"iptables", "-t", "nat",
			"-A", "PREROUTING",
			"-d", hostStaticAddr.String(),
			"-j", "DNAT",
			"--to-destination", addr.String(),
		)
		iptables.Unlock()
		if err != nil {
			return
		}
	}

	store.RemAddress(virt.Id)
	store.RemRoutes(virt.Id)

	hostIps := []string{}
	if hostStaticAddr != nil {
		hostIps = append(hostIps, hostStaticAddr.String())
	}

	coll := db.Instances()
	err = coll.UpdateId(virt.Id, &bson.M{
		"$set": &bson.M{
			"private_ips":       []string{addr.String()},
			"private_ips6":      []string{addr6.String()},
			"gateway_ips":       []string{gatewayCidr},
			"gateway_ips6":      []string{gatewayAddr6.String() + "/64"},
			"network_namespace": namespace,
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
