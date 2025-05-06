package setup

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/ipset"
	"github.com/pritunl/pritunl-cloud/iptables"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vpc"
)

func Iptables() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	namespaces, err := utils.GetNamespaces()
	if err != nil {
		return
	}

	nodeDatacenter, err := node.Self.GetDatacenter(db)
	if err != nil {
		return
	}

	vpcs := []*vpc.Vpc{}
	if !nodeDatacenter.IsZero() {
		vpcs, err = vpc.GetDatacenter(db, nodeDatacenter)
		if err != nil {
			return
		}
	}

	instances, err := instance.GetAllVirt(db, &bson.M{
		"node": node.Self.Id,
	}, nil, nil)
	if err != nil {
		return
	}

	specRules, nodePortsMap, err := firewall.GetSpecRulesSlow(
		db, node.Self.Id, instances)
	if err != nil {
		return
	}

	nodeFirewall, firewalls, firewallMaps, _, err := firewall.GetAllIngress(
		db, node.Self, instances, specRules, nodePortsMap)
	if err != nil {
		return
	}

	err = ipset.Init(namespaces, instances, nodeFirewall, firewalls)
	if err != nil {
		return
	}

	err = iptables.Init(namespaces, vpcs, instances, nodeFirewall,
		firewalls, firewallMaps)
	if err != nil {
		return
	}

	err = ipset.InitNames(namespaces, instances, nodeFirewall, firewalls)
	if err != nil {
		return
	}

	return
}
