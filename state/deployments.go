package state

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/unit"
)

var (
	Deployments    = &DeploymentsState{}
	DeploymentsPkg = NewPackage(Deployments)
)

type DeploymentsState struct {
	podsMap                map[primitive.ObjectID]*pod.Pod
	unitsMap               map[primitive.ObjectID]*unit.Unit
	specsMap               map[primitive.ObjectID]*spec.Spec
	specsPodsMap           map[primitive.ObjectID]*pod.Pod
	specsPodUnitsMap       map[primitive.ObjectID][]*unit.Unit
	specsUnitsMap          map[primitive.ObjectID]*unit.Unit
	specsDeploymentsMap    map[primitive.ObjectID]*deployment.Deployment
	specsDomainsMap        map[primitive.ObjectID]*domain.Domain
	specsSecretsMap        map[primitive.ObjectID]*secret.Secret
	specsCertsMap          map[primitive.ObjectID]*certificate.Certificate
	deploymentsNode        map[primitive.ObjectID]*deployment.Deployment
	deploymentsReservedMap map[primitive.ObjectID]*deployment.Deployment
	deploymentsDeployedMap map[primitive.ObjectID]*deployment.Deployment
	deploymentsInactiveMap map[primitive.ObjectID]*deployment.Deployment
}

func (p *DeploymentsState) Pod(pdId primitive.ObjectID) *pod.Pod {
	return p.podsMap[pdId]
}

func (p *DeploymentsState) PodsMap() map[primitive.ObjectID]*pod.Pod {
	return p.podsMap
}

func (p *DeploymentsState) Unit(pdId primitive.ObjectID) *unit.Unit {
	return p.unitsMap[pdId]
}

func (p *DeploymentsState) UnitsMap() map[primitive.ObjectID]*unit.Unit {
	return p.unitsMap
}

func (p *DeploymentsState) Spec(commitId primitive.ObjectID) *spec.Spec {
	return p.specsMap[commitId]
}

func (p *DeploymentsState) SpecsMap() map[primitive.ObjectID]*spec.Spec {
	return p.specsMap
}

func (p *DeploymentsState) SpecPod(pdId primitive.ObjectID) *pod.Pod {
	return p.specsPodsMap[pdId]
}

func (p *DeploymentsState) SpecPodUnits(pdId primitive.ObjectID) []*unit.Unit {
	return p.specsPodUnitsMap[pdId]
}

func (p *DeploymentsState) SpecUnit(unitId primitive.ObjectID) *unit.Unit {
	return p.specsUnitsMap[unitId]
}

func (p *DeploymentsState) SpecsUnitsMap() map[primitive.ObjectID]*unit.Unit {
	return p.specsUnitsMap
}

func (p *DeploymentsState) SpecDomain(domnId primitive.ObjectID) *domain.Domain {
	return p.specsDomainsMap[domnId]
}

func (p *DeploymentsState) SpecSecret(secrID primitive.ObjectID) *secret.Secret {
	return p.specsSecretsMap[secrID]
}

func (p *DeploymentsState) SpecCert(
	certId primitive.ObjectID) *certificate.Certificate {

	return p.specsCertsMap[certId]
}

func (p *DeploymentsState) SpecCertMap() map[primitive.ObjectID]*certificate.Certificate {
	return p.specsCertsMap
}

func (p *DeploymentsState) DeploymentsNode() map[primitive.ObjectID]*deployment.Deployment {
	return p.deploymentsNode
}

func (p *DeploymentsState) DeploymentReserved(deplyId primitive.ObjectID) *deployment.Deployment {
	return p.deploymentsReservedMap[deplyId]
}

func (p *DeploymentsState) DeploymentsReserved() (
	deplys map[primitive.ObjectID]*deployment.Deployment) {

	deplys = p.deploymentsReservedMap
	return
}

func (p *DeploymentsState) DeploymentDeployed(deplyId primitive.ObjectID) *deployment.Deployment {
	return p.deploymentsDeployedMap[deplyId]
}

func (p *DeploymentsState) DeploymentsDeployed() (
	deplys map[primitive.ObjectID]*deployment.Deployment) {

	deplys = p.deploymentsDeployedMap
	return
}

func (p *DeploymentsState) DeploymentsDestroy() (
	deplys map[primitive.ObjectID]*deployment.Deployment) {

	deplys = p.deploymentsInactiveMap
	return
}

func (p *DeploymentsState) DeploymentInactive(deplyId primitive.ObjectID) *deployment.Deployment {
	return p.deploymentsInactiveMap[deplyId]
}

func (p *DeploymentsState) DeploymentsInactive() (
	deplys map[primitive.ObjectID]*deployment.Deployment) {

	deplys = p.deploymentsInactiveMap
	return
}

func (p *DeploymentsState) Deployment(deplyId primitive.ObjectID) (
	deply *deployment.Deployment) {

	deply = p.deploymentsDeployedMap[deplyId]
	if deply != nil {
		return
	}

	deply = p.deploymentsReservedMap[deplyId]
	if deply != nil {
		return
	}

	deply = p.deploymentsInactiveMap[deplyId]
	if deply != nil {
		return
	}

	return
}

func (p *DeploymentsState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	deployments, err := deployment.GetAll(db, &bson.M{
		"node": node.Self.Id,
	})
	if err != nil {
		return
	}

	deploymentsNode := map[primitive.ObjectID]*deployment.Deployment{}
	deploymentsReservedMap := map[primitive.ObjectID]*deployment.Deployment{}
	deploymentsDeployedMap := map[primitive.ObjectID]*deployment.Deployment{}
	deploymentsInactiveMap := map[primitive.ObjectID]*deployment.Deployment{}
	deploymentsIdSet := set.NewSet()
	podIdsSet := set.NewSet()
	unitIdsSet := set.NewSet()
	specIdsSet := set.NewSet()
	for _, deply := range deployments {
		deploymentsNode[deply.Id] = deply

		deploymentsIdSet.Add(deply.Id)
		switch deply.State {
		case deployment.Reserved:
			deploymentsReservedMap[deply.Id] = deply
			break
		case deployment.Deployed:
			switch deply.Action {
			case deployment.Destroy, deployment.Archive,
				deployment.Restore:

				deploymentsInactiveMap[deply.Id] = deply
				break
			default:
				deploymentsDeployedMap[deply.Id] = deply
			}
			break
		case deployment.Archived:
			deploymentsInactiveMap[deply.Id] = deply
			break
		}

		podIdsSet.Add(deply.Pod)
		unitIdsSet.Add(deply.Unit)
		specIdsSet.Add(deply.Spec)
	}
	p.deploymentsNode = deploymentsNode

	podIds := []primitive.ObjectID{}
	for podId := range podIdsSet.Iter() {
		podIds = append(podIds, podId.(primitive.ObjectID))
	}

	unitIds := []primitive.ObjectID{}
	for unitId := range unitIdsSet.Iter() {
		unitIds = append(unitIds, unitId.(primitive.ObjectID))
	}

	specIds := []primitive.ObjectID{}
	for specId := range specIdsSet.Iter() {
		specIds = append(specIds, specId.(primitive.ObjectID))
	}

	specs := []*spec.Spec{}
	if len(specIds) > 0 {
		specs, err = spec.GetAll(db, &bson.M{
			"_id": &bson.M{
				"$in": specIds,
			},
		})
		if err != nil {
			return
		}
	}

	specSecretsSet := set.NewSet()
	specCertsSet := set.NewSet()
	specPodsSet := set.NewSet()
	specUnitsSet := set.NewSet()
	specDomainsSet := set.NewSet()
	specsMap := map[primitive.ObjectID]*spec.Spec{}
	for _, spc := range specs {
		specsMap[spc.Id] = spc

		if spc.Instance != nil {
			if spc.Instance.Pods != nil {
				for _, pdId := range spc.Instance.Pods {
					specPodsSet.Add(pdId)
				}
			}

			if spc.Instance.Secrets != nil {
				for _, secrId := range spc.Instance.Secrets {
					specSecretsSet.Add(secrId)
				}
			}

			if spc.Instance.Certificates != nil {
				for _, certId := range spc.Instance.Certificates {
					specCertsSet.Add(certId)
				}
			}
		}

		if spc.Firewall != nil {
			for _, rule := range spc.Firewall.Ingress {
				for _, ref := range rule.Sources {
					specUnitsSet.Add(ref.Id)
				}
			}
		}

		if spc.Domain != nil {
			for _, record := range spc.Domain.Records {
				specDomainsSet.Add(record.Domain)
			}
		}
	}
	p.specsMap = specsMap

	specCertIds := []primitive.ObjectID{}
	for certId := range specCertsSet.Iter() {
		specCertIds = append(specCertIds, certId.(primitive.ObjectID))
	}

	specSecretIds := []primitive.ObjectID{}
	for secrId := range specSecretsSet.Iter() {
		specSecretIds = append(specSecretIds, secrId.(primitive.ObjectID))
	}

	specPodIds := []primitive.ObjectID{}
	for pdId := range specPodsSet.Iter() {
		specPodIds = append(specPodIds, pdId.(primitive.ObjectID))
	}

	specUnitIds := []primitive.ObjectID{}
	for unitId := range specUnitsSet.Iter() {
		specUnitIds = append(specUnitIds, unitId.(primitive.ObjectID))
	}

	specDomainIds := []primitive.ObjectID{}
	for pdId := range specDomainsSet.Iter() {
		specDomainIds = append(specDomainIds, pdId.(primitive.ObjectID))
	}

	specCerts := []*certificate.Certificate{}
	if len(specCertIds) > 0 {
		specCerts, err = certificate.GetAll(db, &bson.M{
			"_id": &bson.M{
				"$in": specCertIds,
			},
		})
		if err != nil {
			return
		}
	}

	specsCertsMap := map[primitive.ObjectID]*certificate.Certificate{}
	for _, specCert := range specCerts {
		specsCertsMap[specCert.Id] = specCert
	}
	p.specsCertsMap = specsCertsMap

	specSecrets := []*secret.Secret{}
	if len(specSecretIds) > 0 {
		specSecrets, err = secret.GetAll(db, &bson.M{
			"_id": &bson.M{
				"$in": specSecretIds,
			},
		})
		if err != nil {
			return
		}
	}

	specsSecretsMap := map[primitive.ObjectID]*secret.Secret{}
	for _, specSecret := range specSecrets {
		specsSecretsMap[specSecret.Id] = specSecret
	}
	p.specsSecretsMap = specsSecretsMap

	specPods := []*pod.Pod{}
	if len(specPodIds) > 0 {
		specPods, err = pod.GetAll(db, &bson.M{
			"_id": &bson.M{
				"$in": specPodIds,
			},
		})
		if err != nil {
			return
		}
	}

	specsPodsMap := map[primitive.ObjectID]*pod.Pod{}
	for _, specPod := range specPods {
		specsPodsMap[specPod.Id] = specPod
	}
	p.specsPodsMap = specsPodsMap

	specUnits := []*unit.Unit{}
	if len(specUnitIds) > 0 || len(specPodIds) > 0 {
		specUnits, err = unit.GetAll(db, &bson.M{
			"$or": []*bson.M{
				&bson.M{
					"_id": &bson.M{
						"$in": specUnitIds,
					},
				},
				&bson.M{
					"pod": &bson.M{
						"$in": specPodIds,
					},
				},
			},
		})
		if err != nil {
			return
		}
	}

	specDeploymentsSet := set.NewSet()
	specsUnitsMap := map[primitive.ObjectID]*unit.Unit{}
	specsPodUnitsMap := map[primitive.ObjectID][]*unit.Unit{}
	for _, specUnit := range specUnits {
		specsUnitsMap[specUnit.Id] = specUnit

		specsPodUnitsMap[specUnit.Pod] = append(
			specsPodUnitsMap[specUnit.Pod], specUnit)

		for _, deplyId := range specUnit.Deployments {
			specDeploymentsSet.Add(deplyId)
		}
	}
	p.specsUnitsMap = specsUnitsMap
	p.specsPodUnitsMap = specsPodUnitsMap

	specDomains, err := domain.GetLoadedAllIds(db, specDomainIds)
	if err != nil {
		return
	}

	specsDomainsMap := map[primitive.ObjectID]*domain.Domain{}
	for _, specDomain := range specDomains {
		specsDomainsMap[specDomain.Id] = specDomain
	}
	p.specsDomainsMap = specsDomainsMap

	specDeploymentIds := []primitive.ObjectID{}
	for deplyIdInf := range specDeploymentsSet.Iter() {
		deplyId := deplyIdInf.(primitive.ObjectID)
		if !deploymentsIdSet.Contains(deplyId) {
			specDeploymentIds = append(specDeploymentIds, deplyId)
		}
	}

	specDeployments := []*deployment.Deployment{}
	if len(specDeploymentIds) > 0 {
		specDeployments, err = deployment.GetAll(db, &bson.M{
			"_id": &bson.M{
				"$in": specDeploymentIds,
			},
		})
		if err != nil {
			return
		}
	}

	for _, specDeployment := range specDeployments {
		deploymentsIdSet.Add(specDeployment.Id)

		switch specDeployment.State {
		case deployment.Reserved:
			deploymentsReservedMap[specDeployment.Id] = specDeployment
			break
		case deployment.Deployed:
			switch specDeployment.Action {
			case deployment.Destroy, deployment.Archive,
				deployment.Restore:

				deploymentsInactiveMap[specDeployment.Id] = specDeployment
				break
			default:
				deploymentsDeployedMap[specDeployment.Id] = specDeployment
			}
			break
		case deployment.Archived:
			deploymentsInactiveMap[specDeployment.Id] = specDeployment
			break
		}
	}

	pods := []*pod.Pod{}
	if len(podIds) > 0 {
		pods, err = pod.GetAll(db, &bson.M{
			"_id": &bson.M{
				"$in": podIds,
			},
		})
		if err != nil {
			return
		}
	}

	podsMap := map[primitive.ObjectID]*pod.Pod{}
	for _, pd := range pods {
		podsMap[pd.Id] = pd
	}
	p.podsMap = podsMap

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
	p.unitsMap = unitsMap

	podDeploymentIds := []primitive.ObjectID{}
	for deplyIdInf := range podDeploymentsSet.Iter() {
		deplyId := deplyIdInf.(primitive.ObjectID)
		if !deploymentsIdSet.Contains(deplyId) {
			podDeploymentIds = append(podDeploymentIds, deplyId)
		}
	}

	podDeployments := []*deployment.Deployment{}
	if len(podDeploymentIds) > 0 {
		podDeployments, err = deployment.GetAll(db, &bson.M{
			"_id": &bson.M{
				"$in": podDeploymentIds,
			},
		})
		if err != nil {
			return
		}
	}

	for _, podDeployment := range podDeployments {
		deploymentsIdSet.Add(podDeployment.Id)

		switch podDeployment.State {
		case deployment.Reserved:
			deploymentsReservedMap[podDeployment.Id] = podDeployment
			break
		case deployment.Deployed:
			switch podDeployment.Action {
			case deployment.Destroy, deployment.Archive,
				deployment.Restore:

				deploymentsInactiveMap[podDeployment.Id] = podDeployment
				break
			default:
				deploymentsDeployedMap[podDeployment.Id] = podDeployment
			}
			break
		case deployment.Archived:
			deploymentsInactiveMap[podDeployment.Id] = podDeployment
			break
		}
	}
	p.deploymentsReservedMap = deploymentsReservedMap
	p.deploymentsDeployedMap = deploymentsDeployedMap
	p.deploymentsInactiveMap = deploymentsInactiveMap

	return
}

func (p *DeploymentsState) Apply(st *State) {
	st.Pod = p.Pod
	st.PodsMap = p.PodsMap
	st.Unit = p.Unit
	st.UnitsMap = p.UnitsMap
	st.Spec = p.Spec
	st.SpecPod = p.SpecPod
	st.SpecPodUnits = p.SpecPodUnits
	st.SpecUnit = p.SpecUnit
	st.SpecsUnitsMap = p.SpecsUnitsMap
	st.SpecDomain = p.SpecDomain
	st.SpecSecret = p.SpecSecret
	st.SpecCert = p.SpecCert
	st.DeploymentsNode = p.DeploymentsNode
	st.DeploymentReserved = p.DeploymentReserved
	st.DeploymentsReserved = p.DeploymentsReserved
	st.DeploymentDeployed = p.DeploymentDeployed
	st.DeploymentsDeployed = p.DeploymentsDeployed
	st.DeploymentsDestroy = p.DeploymentsDestroy
	st.DeploymentInactive = p.DeploymentInactive
	st.DeploymentsInactive = p.DeploymentsInactive
	st.Deployment = p.Deployment
}
