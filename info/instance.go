package info

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/tools/set"
)

func NewInstance(stat *state.State, inst *instance.Instance) (
	inf *instance.Info) {

	inf = &instance.Info{
		Disks:         []string{},
		FirewallRules: map[string]string{},
		Authorities:   []string{},
		OracleSubnets: []*node.OracleSubnet{},
	}

	nde := stat.Node()
	if inst.Node != nde.Id {
		return
	}

	inf.Node = nde.Name
	if len(nde.PublicIps) > 0 {
		inf.NodePublicIp = nde.PublicIps[0]
	}
	inf.Iscsi = nde.Iscsi
	inf.Isos = nde.LocalIsos
	inf.OracleSubnets = nde.GetOracleSubnetsName()
	if nde.UsbPassthrough {
		inf.UsbDevices = nde.UsbDevices
	}
	if nde.PciDevices != nil {
		inf.PciDevices = nde.PciDevices
	}
	if nde.InstanceDrives != nil {
		inf.DriveDevices = nde.InstanceDrives
	}

	dc := stat.GetDatacenter(inst.Datacenter)

	if dc != nil {
		inf.Mtu = dc.GetInstanceMtu()
	}

	instDisks := stat.GetInstaceDisks(inst.Id)
	for _, dsk := range instDisks {
		inf.Disks = append(
			inf.Disks,
			fmt.Sprintf("%s: %s", dsk.Index, dsk.Name),
		)
	}

	firewallRulesKeys := []string{}
	firewallRules := map[string]set.Set{}
	namespaces := stat.GetInstanceNamespaces(inst.Id)
	firewalls := stat.Firewalls()
	for _, namespace := range namespaces {
		for _, rule := range firewalls[namespace] {
			key := rule.Protocol
			if rule.Port != "" {
				key += ":" + rule.Port
			}

			rules := firewallRules[key]
			if rules == nil {
				rules = set.NewSet()
				firewallRules[key] = rules
				firewallRulesKeys = append(
					firewallRulesKeys,
					key,
				)
			}

			for _, sourceIp := range rule.SourceIps {
				rules.Add(sourceIp)
			}
		}
	}

	sort.Strings(firewallRulesKeys)
	for _, key := range firewallRulesKeys {
		rules := firewallRules[key]

		vals := []string{}
		for rule := range rules.Iter() {
			vals = append(vals, rule.(string))
		}
		sort.Strings(vals)

		inf.FirewallRules[key] = strings.Join(vals, ", ")
	}

	authrs := stat.GetInstaceAuthorities(inst.NetworkRoles)
	for _, authr := range authrs {
		inf.Authorities = append(inf.Authorities, authr.Name)
	}
	sort.Strings(inf.Authorities)

	return
}
