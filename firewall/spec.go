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
	deploymentsMap map[primitive.ObjectID]*deployment.Deployment,
	specsMap map[primitive.ObjectID]*spec.Commit,
	unitsMap map[primitive.ObjectID]*pod.Unit,
	specDeployments map[primitive.ObjectID]*deployment.Deployment) (
	firewalls map[string][]*Rule, err error) {

	firewalls = map[string][]*Rule{}
	for _, inst := range instances {
		if inst.Deployment.IsZero() {
			continue
		}

		deply := deploymentsMap[inst.Deployment]
		if deply == nil {
			continue
		}

		spec := specsMap[deply.Spec]
		if spec == nil {
			continue
		}

		if spec.Firewall == nil || spec.Firewall.Ingress == nil {
			continue
		}

		if !inst.IsActive() {
			continue
		}

		namespaces := []string{}
		for i := range inst.Virt.NetworkAdapters {
			namespaces = append(namespaces, vm.GetNamespace(inst.Id, i))
		}

		for _, specRule := range spec.Firewall.Ingress {
			for _, ruleUnitId := range specRule.Units {
				ruleUnit := unitsMap[ruleUnitId]
				if ruleUnit == nil {
					continue
				}

				rule := &Rule{
					Protocol: specRule.Protocol,
					Port:     specRule.Port,
				}

				for _, ruleDeplyRec := range ruleUnit.Deployments {
					ruleDeply := specDeployments[ruleDeplyRec.Id]
					if ruleDeply == nil {
						continue
					}

					instData := ruleDeply.InstanceData
					if instData == nil {
						continue
					}

					for _, ip := range instData.PrivateIps {
						rule.SourceIps = append(
							rule.SourceIps,
							strings.Split(ip, "/")[0]+"/32",
						)
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
	}

	return
}
