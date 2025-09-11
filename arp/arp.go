package arp

import (
	"encoding/json"
	"net"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
)

type Record struct {
	Ip  string
	Mac string
}

type entry struct {
	Dst    string   `json:"dst"`
	Dev    string   `json:"dev"`
	Lladdr string   `json:"lladdr,omitempty"`
	State  []string `json:"state"`
	Router string   `json:"router,omitempty"`
}

func GetRecords(namespace string) (records set.Set, err error) {
	records = set.NewSet()

	output, _ := utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ip", "--json",
		"neighbor",
	)

	if output == "" {
		return
	}

	var entries []*entry
	err = json.Unmarshal([]byte(output), &entries)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "arp: Failed to process arp output"),
		}
		return
	}

	for _, ent := range entries {
		if ent.Dev != settings.Hypervisor.BridgeIfaceName {
			continue
		}

		mac := ""
		if ent.State != nil {
			for _, state := range ent.State {
				if state == "PERMANENT" {
					mac = ent.Lladdr
				}
			}
		}

		ip := net.ParseIP(ent.Dst)
		if ip != nil {
			records.Add(Record{
				Ip:  ip.String(),
				Mac: mac,
			})
		}
	}

	return
}

func BuildState(instances []*instance.Instance,
	vpcsMap map[bson.ObjectID]*vpc.Vpc,
	vpcIpsMap map[bson.ObjectID][]*vpc.VpcIp) (
	recrds map[string]set.Set) {

	recrds = map[string]set.Set{}

	for _, inst := range instances {
		if !inst.IsActive() {
			continue
		}

		for i, adapter := range inst.Virt.NetworkAdapters {
			namespace := vm.GetNamespace(inst.Id, i)
			vc := vpcsMap[adapter.Vpc]
			vpcIps := vpcIpsMap[adapter.Vpc]

			newRecrds := set.NewSet()

			for _, vpcIp := range vpcIps {
				if vpcIp.Instance.IsZero() {
					continue
				}

				addr := vpcIp.GetIp()

				newRecrds.Add(Record{
					Ip:  vpc.GetIp6(vpcIp.Vpc, vpcIp.Instance).String(),
					Mac: vm.GetMacAddr(vpcIp.Instance, adapter.Vpc),
				})

				newRecrds.Add(Record{
					Ip:  addr.String(),
					Mac: vm.GetMacAddr(vpcIp.Instance, adapter.Vpc),
				})
			}

			if vc != nil && vc.Arps != nil {
				for _, ap := range vc.Arps {
					newRecrds.Add(Record{
						Ip:  ap.Ip,
						Mac: ap.Mac,
					})
				}
			}

			recrds[namespace] = newRecrds
		}
	}

	return
}

func ApplyState(namespace string, oldState, newState set.Set) (
	changed bool, err error) {

	addRecords := newState.Copy()
	remRecords := oldState.Copy()

	addRecords.Subtract(oldState)
	remRecords.Subtract(newState)

	for recordInf := range remRecords.Iter() {
		recrd := recordInf.(Record)
		changed = true

		utils.ExecCombinedOutputLogged(
			[]string{
				"No such file",
			},
			"ip", "netns", "exec", namespace,
			"ip", "neighbor",
			"del", recrd.Ip,
			"dev", settings.Hypervisor.BridgeIfaceName,
		)
	}

	for recordInf := range addRecords.Iter() {
		recrd := recordInf.(Record)
		changed = true

		utils.ExecCombinedOutputLogged(
			[]string{
				"No such file",
			},
			"ip", "netns", "exec", namespace,
			"ip", "neighbor",
			"del", recrd.Ip,
			"dev", settings.Hypervisor.BridgeIfaceName,
		)

		_, err = utils.ExecCombinedOutputLogged(
			[]string{
				"File exists",
			},
			"ip", "netns", "exec", namespace,
			"ip", "neighbor",
			"replace", recrd.Ip,
			"lladdr", recrd.Mac,
			"dev", settings.Hypervisor.BridgeIfaceName,
			"nud", "permanent",
		)
		if err != nil {
			return
		}
	}

	return
}
