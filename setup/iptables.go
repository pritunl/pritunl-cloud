package setup

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/ipset"
	"github.com/pritunl/pritunl-cloud/iptables"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Iptables() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	namespaces, err := utils.GetNamespaces()
	if err != nil {
		return
	}

	disks, err := disk.GetNode(db, node.Self.Id)
	if err != nil {
		return
	}

	instances, err := instance.GetAllVirt(db, &bson.M{
		"node": node.Self.Id,
	}, disks)
	if err != nil {
		return
	}

	nodeFirewall, firewalls, err := firewall.GetAllIngress(
		db, node.Self, instances)
	if err != nil {
		return
	}

	err = ipset.Init(namespaces, instances, nodeFirewall, firewalls)
	if err != nil {
		return
	}

	err = iptables.Init(namespaces, instances, nodeFirewall, firewalls)
	if err != nil {
		return
	}

	err = ipset.InitNames(namespaces, instances, nodeFirewall, firewalls)
	if err != nil {
		return
	}

	return
}
