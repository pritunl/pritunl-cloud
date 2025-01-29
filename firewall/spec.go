package firewall

import (
	"strings"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/vm"
)

func GetSpecRules(instances []*instance.Instance,
	deploymentsNode map[primitive.ObjectID]*deployment.Deployment,
	specsMap map[primitive.ObjectID]*spec.Commit,
	specsPodsUnitsMap map[primitive.ObjectID]*pod.Unit,
	deploymentsDeployedMap map[primitive.ObjectID]*deployment.Deployment) (
	firewalls map[string][]*Rule, err error) {

	firewalls = map[string][]*Rule{}
	for _, inst := range instances {
		if inst.Deployment.IsZero() {
			continue
		}

		deply := deploymentsNode[inst.Deployment]
		if deply == nil {
			continue
		}

		spc := specsMap[deply.Spec]
		if spc == nil {
			continue
		}

		if spc.Firewall == nil || spc.Firewall.Ingress == nil {
			continue
		}

		if !inst.IsActive() {
			continue
		}

		namespaces := []string{}
		for i := range inst.Virt.NetworkAdapters {
			namespaces = append(namespaces, vm.GetNamespace(inst.Id, i))
		}

		for _, specRule := range spc.Firewall.Ingress {
			rule := &Rule{
				Protocol:  specRule.Protocol,
				Port:      specRule.Port,
				SourceIps: specRule.SourceIps,
			}

			for _, ref := range specRule.Sources {
				if ref.Kind != spec.Unit {
					continue
				}

				ruleUnit := specsPodsUnitsMap[ref.Id]
				if ruleUnit == nil {
					continue
				}

				for _, ruleDeplyRec := range ruleUnit.Deployments {
					ruleDeply := deploymentsDeployedMap[ruleDeplyRec.Id]
					if ruleDeply == nil {
						continue
					}

					instData := ruleDeply.InstanceData
					if instData == nil {
						continue
					}

					if ref.Selector == "" || ref.Selector == "private_ips" {
						for _, ip := range instData.PrivateIps {
							rule.SourceIps = append(
								rule.SourceIps,
								strings.Split(ip, "/")[0]+"/32",
							)
						}
					}
				}
			}

			if len(rule.SourceIps) == 0 {
				continue
			}

			for _, namespace := range namespaces {
				firewalls[namespace] = append(firewalls[namespace], rule)
			}
		}
	}

	return
}
