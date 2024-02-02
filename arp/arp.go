package arp

import (
	"encoding/json"
	"fmt"
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/instance"
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
		"ip", "-4", "--json",
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
		if ent.Dev != "br0" {
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

		records.Add(Record{
			Ip:  ent.Dst,
			Mac: mac,
		})
	}

	return
}

func BuildState(instances []*instance.Instance,
	vpcIpsMap map[primitive.ObjectID][]*vpc.VpcIp) (
	recrds map[string]set.Set) {

	recrds = map[string]set.Set{}

	for _, inst := range instances {
		if !inst.IsActive() {
			continue
		}

		for i, adapter := range inst.Virt.NetworkAdapters {
			namespace := vm.GetNamespace(inst.Id, i)
			vpcIps := vpcIpsMap[adapter.Vpc]

			newRecrds := set.NewSet()

			for _, vpcIp := range vpcIps {
				if vpcIp.Instance.IsZero() {
					continue
				}

				newRecrds.Add(Record{
					Ip:  vpcIp.GetIp().String(),
					Mac: vm.GetMacAddr(vpcIp.Instance, adapter.Vpc),
				})
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

		fmt.Printf("[DEL][%s] %s -> %s\n", namespace, recrd.Ip, recrd.Mac)

		utils.ExecCombinedOutputLogged(
			[]string{
				"No such file",
			},
			"ip", "netns", "exec", namespace,
			"ip", "neighbor",
			"del", recrd.Ip,
			"dev", "br0",
		)
	}

	for recordInf := range addRecords.Iter() {
		recrd := recordInf.(Record)
		changed = true

		fmt.Printf("[ADD][%s] %s -> %s\n", namespace, recrd.Ip, recrd.Mac)

		utils.ExecCombinedOutputLogged(
			[]string{
				"No such file",
			},
			"ip", "netns", "exec", namespace,
			"ip", "neighbor",
			"del", recrd.Ip,
			"dev", "br0",
		)

		_, err = utils.ExecCombinedOutputLogged(
			[]string{
				"File exists",
			},
			"ip", "netns", "exec", namespace,
			"ip", "neighbor",
			"add", recrd.Ip,
			"lladdr", recrd.Mac,
			"dev", "br0",
			"nud", "permanent",
		)
		if err != nil {
			return
		}
	}

	return
}
