package instanceinfo

import (
	"fmt"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"gopkg.in/mgo.v2/bson"
)

type Info struct {
	Instance      bson.ObjectId `json:"instance"`
	FirewallRules []string      `json:"firewall_rules"`
}

func Get(db *database.Database, instId bson.ObjectId) (
	info *Info, err error) {

	info = &Info{
		Instance:      instId,
		FirewallRules: []string{},
	}

	inst, err := instance.Get(db, instId)
	if err != nil {
		return
	}

	fires, err := firewall.GetOrgRoles(db,
		inst.Organization, inst.NetworkRoles)
	if err != nil {
		return
	}

	for _, fire := range fires {
		for _, rule := range fire.Ingress {
			for _, sourceIp := range rule.SourceIps {
				if rule.Port == "" {
					info.FirewallRules = append(
						info.FirewallRules,
						fmt.Sprintf(
							"%s - %s",
							rule.Protocol,
							sourceIp,
						),
					)
				} else {
					info.FirewallRules = append(
						info.FirewallRules,
						fmt.Sprintf(
							"%s:%s - %s",
							rule.Protocol,
							rule.Port,
							sourceIp,
						),
					)
				}
			}
		}
	}

	return
}
