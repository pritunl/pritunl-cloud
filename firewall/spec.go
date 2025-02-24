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
	"github.com/pritunl/pritunl-cloud/vm"
)

func GetSpecRules(instances []*instance.Instance,
	deploymentsNode map[primitive.ObjectID]*deployment.Deployment,
	specsMap map[primitive.ObjectID]*spec.Spec,
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
	unitIds := set.NewSet()
	specIdsSet := set.NewSet()
	for _, deply := range deployments {
		deploymentsNode[deply.Id] = deply
		deploymentsIdSet.Add(deply.Id)

		if deply.State == deployment.Deployed ||
			deply.State == deployment.Migrate {

			deploymentsDeployedMap[deply.Id] = deply
		}

		podIdsSet.Add(deply.Pod)
		unitIds.Add(deply.Unit)
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

	specPodsSet := set.NewSet()
	specsMap := map[primitive.ObjectID]*spec.Spec{}
	for _, spc := range specs {
		specsMap[spc.Id] = spc

		if spc.Firewall != nil {
			for _, rule := range spc.Firewall.Ingress {
				for _, ref := range rule.Sources {
					specPodsSet.Add(ref.Realm)
				}
			}
		}
	}

	specPodIds := []primitive.ObjectID{}
	for pdId := range specPodsSet.Iter() {
		specPodIds = append(specPodIds, pdId.(primitive.ObjectID))
	}

	specPods, err := pod.GetAll(db, &bson.M{
		"_id": &bson.M{
			"$in": specPodIds,
		},
	})
	if err != nil {
		return
	}

	specDeploymentsSet := set.NewSet()
	specsPodsMap := map[primitive.ObjectID]*pod.Pod{}
	specsPodsUnitsMap := map[primitive.ObjectID]*pod.Unit{}
	for _, specPod := range specPods {
		specsPodsMap[specPod.Id] = specPod

		for _, unit := range specPod.Units {
			specsPodsUnitsMap[unit.Id] = unit
			for _, deply := range unit.Deployments {
				specDeploymentsSet.Add(deply.Id)
			}
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

		if specDeployment.State == deployment.Deployed ||
			specDeployment.State == deployment.Migrate {

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

	nodePortsDeployments := map[primitive.ObjectID][]*nodeport.Mapping{}
	podDeploymentsSet := set.NewSet()
	podsMap := map[primitive.ObjectID]*pod.Pod{}
	podsUnitsMap := map[primitive.ObjectID]*pod.Unit{}
	for _, pd := range pods {
		podsMap[pd.Id] = pd

		for _, unit := range pd.Units {
			if !unitIds.Contains(unit.Id) {
				continue
			}
			podsUnitsMap[unit.Id] = unit

			for _, deply := range unit.Deployments {
				nodePortsDeployments[deply.Id] = append(
					nodePortsDeployments[deply.Id], unit.NodePorts...)

				podDeploymentsSet.Add(deply.Id)
			}
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

		if podDeployment.State == deployment.Deployed ||
			podDeployment.State == deployment.Migrate {

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
		} else {
			nodePortsMap[inst.NetworkNamespace] = append(
				nodePortsMap[inst.NetworkNamespace],
				nodePortsDeployments[inst.Deployment]...,
			)
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
