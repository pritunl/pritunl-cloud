package firewall

import (
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/nodeport"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/unit"
	"github.com/pritunl/pritunl-cloud/vm"
)

func GetSpecRules(instances []*instance.Instance,
	deploymentsNode map[primitive.ObjectID]*deployment.Deployment,
	specsMap map[primitive.ObjectID]*spec.Spec,
	specsUnitsMap map[primitive.ObjectID]*unit.Unit,
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

				ruleUnit := specsUnitsMap[ref.Id]
				if ruleUnit == nil {
					continue
				}

				for _, ruleDeplyId := range ruleUnit.Deployments {
					ruleDeply := deploymentsDeployedMap[ruleDeplyId]
					if ruleDeply == nil {
						continue
					}

					instData := ruleDeply.InstanceData
					if instData == nil {
						continue
					}

					switch ref.Selector {
					case "", "private_ips":
						for _, ip := range instData.PrivateIps {
							rule.SourceIps = append(
								rule.SourceIps,
								strings.Split(ip, "/")[0]+"/32",
							)
						}
					case "private_ips6":
						for _, ip := range instData.PrivateIps6 {
							rule.SourceIps = append(
								rule.SourceIps,
								strings.Split(ip, "/")[0]+"/32",
							)
						}
					case "public_ips":
						for _, ip := range instData.PublicIps {
							rule.SourceIps = append(
								rule.SourceIps,
								strings.Split(ip, "/")[0]+"/32",
							)
						}
					case "public_ips6":
						for _, ip := range instData.PublicIps6 {
							rule.SourceIps = append(
								rule.SourceIps,
								strings.Split(ip, "/")[0]+"/32",
							)
						}
					case "oracle_private_ips":
						for _, ip := range instData.OraclePrivateIps {
							rule.SourceIps = append(
								rule.SourceIps,
								strings.Split(ip, "/")[0]+"/32",
							)
						}
					case "oracle_public_ips":
						for _, ip := range instData.OraclePublicIps {
							rule.SourceIps = append(
								rule.SourceIps,
								strings.Split(ip, "/")[0]+"/32",
							)
						}
					case "oracle_public_ips6":
						for _, ip := range instData.OraclePublicIps6 {
							rule.SourceIps = append(
								rule.SourceIps,
								strings.Split(ip, "/")[0]+"/32",
							)
						}
					case "host_ips":
						for _, ip := range instData.HostIps {
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

func GetSpecRulesSlow(db *database.Database,
	nodeId primitive.ObjectID, instances []*instance.Instance) (
	firewalls map[string][]*Rule, nodePortsMap map[string][]*nodeport.Mapping,
	err error) {

	deployments, err := deployment.GetAll(db, &bson.M{
		"node": nodeId,
	})
	if err != nil {
		return
	}

	deploymentsNode := map[primitive.ObjectID]*deployment.Deployment{}
	deploymentsDeployedMap := map[primitive.ObjectID]*deployment.Deployment{}
	deploymentsIdSet := set.NewSet()
	podIdsSet := set.NewSet()
	unitIdsSet := set.NewSet()
	specIdsSet := set.NewSet()
	for _, deply := range deployments {
		deploymentsNode[deply.Id] = deply
		deploymentsIdSet.Add(deply.Id)

		if deply.State == deployment.Deployed {
			deploymentsDeployedMap[deply.Id] = deply
		}

		podIdsSet.Add(deply.Pod)
		unitIdsSet.Add(deply.Unit)
		specIdsSet.Add(deply.Spec)
	}

	specIds := []primitive.ObjectID{}
	for specId := range specIdsSet.Iter() {
		specIds = append(specIds, specId.(primitive.ObjectID))
	}

	specs, err := spec.GetAll(db, &bson.M{
		"_id": &bson.M{
			"$in": specIds,
		},
	})
	if err != nil {
		return
	}

	specUnitsSet := set.NewSet()
	specsMap := map[primitive.ObjectID]*spec.Spec{}
	for _, spc := range specs {
		specsMap[spc.Id] = spc

		if spc.Firewall != nil {
			for _, rule := range spc.Firewall.Ingress {
				for _, ref := range rule.Sources {
					specUnitsSet.Add(ref.Id)
				}
			}
		}
	}

	specUnitIds := []primitive.ObjectID{}
	for unitId := range specUnitsSet.Iter() {
		specUnitIds = append(specUnitIds, unitId.(primitive.ObjectID))
	}

	specUnits, err := unit.GetAll(db, &bson.M{
		"_id": &bson.M{
			"$in": specUnitIds,
		},
	})
	if err != nil {
		return
	}

	specDeploymentsSet := set.NewSet()
	specsUnitsMap := map[primitive.ObjectID]*unit.Unit{}
	for _, specUnit := range specUnits {
		specsUnitsMap[specUnit.Id] = specUnit

		for _, deplyId := range specUnit.Deployments {
			specDeploymentsSet.Add(deplyId)
		}
	}

	specDeploymentIds := []primitive.ObjectID{}
	for deplyIdInf := range specDeploymentsSet.Iter() {
		deplyId := deplyIdInf.(primitive.ObjectID)
		if !deploymentsIdSet.Contains(deplyId) {
			specDeploymentIds = append(specDeploymentIds, deplyId)
		}
	}

	specDeployments, err := deployment.GetAll(db, &bson.M{
		"_id": &bson.M{
			"$in": specDeploymentIds,
		},
	})
	if err != nil {
		return
	}

	for _, specDeployment := range specDeployments {
		deploymentsIdSet.Add(specDeployment.Id)

		if specDeployment.State == deployment.Deployed {
			deploymentsDeployedMap[specDeployment.Id] = specDeployment
		}
	}

	podIds := []primitive.ObjectID{}
	for podId := range podIdsSet.Iter() {
		podIds = append(podIds, podId.(primitive.ObjectID))
	}

	pods, err := pod.GetAll(db, &bson.M{
		"_id": &bson.M{
			"$in": podIds,
		},
	})
	if err != nil {
		return
	}

	podsMap := map[primitive.ObjectID]*pod.Pod{}
	for _, pd := range pods {
		podsMap[pd.Id] = pd
	}

	unitIds := []primitive.ObjectID{}
	for unitId := range unitIdsSet.Iter() {
		unitIds = append(unitIds, unitId.(primitive.ObjectID))
	}

	units := []*unit.Unit{}
	if len(unitIds) > 0 {
		units, err = unit.GetAll(db, &bson.M{
			"_id": &bson.M{
				"$in": unitIds,
			},
		})
		if err != nil {
			return
		}
	}

	unitsMap := map[primitive.ObjectID]*unit.Unit{}
	podDeploymentsSet := set.NewSet()
	for _, unt := range units {
		unitsMap[unt.Id] = unt

		for _, deplyId := range unt.Deployments {
			podDeploymentsSet.Add(deplyId)
		}
	}

	podDeploymentIds := []primitive.ObjectID{}
	for deplyIdInf := range podDeploymentsSet.Iter() {
		deplyId := deplyIdInf.(primitive.ObjectID)
		if !deploymentsIdSet.Contains(deplyId) {
			podDeploymentIds = append(podDeploymentIds, deplyId)
		}
	}

	podDeployments, err := deployment.GetAll(db, &bson.M{
		"_id": &bson.M{
			"$in": podDeploymentIds,
		},
	})
	if err != nil {
		return
	}

	for _, podDeployment := range podDeployments {
		deploymentsIdSet.Add(podDeployment.Id)

		if podDeployment.State == deployment.Deployed {
			deploymentsDeployedMap[podDeployment.Id] = podDeployment
		}
	}

	nodePortsMap = map[string][]*nodeport.Mapping{}
	firewalls = map[string][]*Rule{}
	for _, inst := range instances {
		nodePortsMap[inst.NetworkNamespace] = append(
			nodePortsMap[inst.NetworkNamespace], inst.NodePorts...)

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

				ruleUnit := specsUnitsMap[ref.Id]
				if ruleUnit == nil {
					continue
				}

				for _, ruleDeplyId := range ruleUnit.Deployments {
					ruleDeply := deploymentsDeployedMap[ruleDeplyId]
					if ruleDeply == nil {
						continue
					}

					instData := ruleDeply.InstanceData
					if instData == nil {
						continue
					}

					switch ref.Selector {
					case "", "private_ips":
						for _, ip := range instData.PrivateIps {
							rule.SourceIps = append(
								rule.SourceIps,
								strings.Split(ip, "/")[0]+"/32",
							)
						}
					case "private_ips6":
						for _, ip := range instData.PrivateIps6 {
							rule.SourceIps = append(
								rule.SourceIps,
								strings.Split(ip, "/")[0]+"/32",
							)
						}
					case "public_ips":
						for _, ip := range instData.PublicIps {
							rule.SourceIps = append(
								rule.SourceIps,
								strings.Split(ip, "/")[0]+"/32",
							)
						}
					case "public_ips6":
						for _, ip := range instData.PublicIps6 {
							rule.SourceIps = append(
								rule.SourceIps,
								strings.Split(ip, "/")[0]+"/32",
							)
						}
					case "oracle_private_ips":
						for _, ip := range instData.OraclePrivateIps {
							rule.SourceIps = append(
								rule.SourceIps,
								strings.Split(ip, "/")[0]+"/32",
							)
						}
					case "oracle_public_ips":
						for _, ip := range instData.OraclePublicIps {
							rule.SourceIps = append(
								rule.SourceIps,
								strings.Split(ip, "/")[0]+"/32",
							)
						}
					case "oracle_public_ips6":
						for _, ip := range instData.OraclePublicIps6 {
							rule.SourceIps = append(
								rule.SourceIps,
								strings.Split(ip, "/")[0]+"/32",
							)
						}
					case "host_ips":
						for _, ip := range instData.HostIps {
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
