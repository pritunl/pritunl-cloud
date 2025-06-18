package state

import (
	"context"

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

type DeploymentsResult struct {
	DeploymentIds []primitive.ObjectID `bson:"deployment_ids"`
	PodIds        []primitive.ObjectID `bson:"pod_ids"`
	UnitIds       []primitive.ObjectID `bson:"unit_ids"`
	SpecIds       []primitive.ObjectID `bson:"spec_ids"`

	Deployments []*deployment.Deployment `bson:"deployments"`
	Pods        []*pod.Pod               `bson:"pods"`
	Units       []*unit.Unit             `bson:"units"`
	Specs       []*spec.Spec             `bson:"specs"`

	SpecPodIds    []primitive.ObjectID `bson:"spec_pod_ids"`
	SpecUnitIds   []primitive.ObjectID `bson:"spec_unit_ids"`
	SpecSecretIds []primitive.ObjectID `bson:"spec_secret_ids"`
	SpecCertIds   []primitive.ObjectID `bson:"spec_cert_ids"`
	SpecDomainIds []primitive.ObjectID `bson:"spec_domain_ids"`

	SpecIdUnits        []*unit.Unit               `bson:"spec_id_units"`
	SpecPodUnits       []*unit.Unit               `bson:"spec_pod_units"`
	SpecPods           []*pod.Pod                 `bson:"spec_pods"`
	SpecSecrets        []*secret.Secret           `bson:"spec_secrets"`
	SpecCerts          []*certificate.Certificate `bson:"spec_certs"`
	SpecDomains        []*domain.Domain           `bson:"spec_domains"`
	SpecDomainsRecords []*domain.Record           `bson:"spec_domains_records"`

	LinkedDeployments []*deployment.Deployment `bson:"linked_deployments"`
}

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

func (p *DeploymentsState) Refresh(pkg *Package, db *database.Database) (err error) {
	pipeline := bson.A{
		bson.M{
			"$match": bson.M{
				"node": node.Self.Id,
			},
		},

		bson.M{
			"$group": bson.M{
				"_id":            nil,
				"deployments":    bson.M{"$push": "$$ROOT"},
				"deployment_ids": bson.M{"$addToSet": "$_id"},
				"pod_ids":        bson.M{"$addToSet": "$pod"},
				"unit_ids":       bson.M{"$addToSet": "$unit"},
				"spec_ids":       bson.M{"$addToSet": "$spec"},
			},
		},

		bson.M{
			"$lookup": bson.M{
				"from":         "specs",
				"localField":   "spec_ids",
				"foreignField": "_id",
				"as":           "specs",
			},
		},

		bson.M{
			"$lookup": bson.M{
				"from":         "units",
				"localField":   "unit_ids",
				"foreignField": "_id",
				"as":           "units",
			},
		},

		bson.M{
			"$lookup": bson.M{
				"from":         "pods",
				"localField":   "pod_ids",
				"foreignField": "_id",
				"as":           "pods",
			},
		},

		bson.M{
			"$addFields": bson.M{
				"specs_data": bson.M{
					"$reduce": bson.M{
						"input": "$specs",
						"initialValue": bson.M{
							"pod_ids":    bson.A{},
							"secret_ids": bson.A{},
							"cert_ids":   bson.A{},
							"domain_ids": bson.A{},
							"unit_ids":   bson.A{},
						},
						"in": bson.M{
							"pod_ids": bson.M{
								"$concatArrays": bson.A{
									"$$value.pod_ids",
									bson.M{
										"$ifNull": bson.A{
											"$$this.instance.pods",
											bson.A{},
										},
									},
								},
							},
							"secret_ids": bson.M{
								"$concatArrays": bson.A{
									"$$value.secret_ids",
									bson.M{
										"$ifNull": bson.A{
											"$$this.instance.secrets",
											bson.A{},
										},
									},
								},
							},
							"cert_ids": bson.M{
								"$concatArrays": bson.A{
									"$$value.cert_ids",
									bson.M{
										"$ifNull": bson.A{
											"$$this.instance.certificates",
											bson.A{},
										},
									},
								},
							},
							"domain_ids": bson.M{
								"$concatArrays": bson.A{
									"$$value.domain_ids",
									bson.M{
										"$map": bson.M{
											"input": bson.M{
												"$ifNull": bson.A{
													"$$this.domain.records",
													bson.A{},
												},
											},
											"as": "record",
											"in": "$$record.domain",
										},
									},
								},
							},
							"unit_ids": bson.M{
								"$concatArrays": bson.A{
									"$$value.unit_ids",
									bson.M{
										"$reduce": bson.M{
											"input": bson.M{
												"$ifNull": bson.A{
													"$$this.firewall.ingress",
													bson.A{},
												},
											},
											"initialValue": bson.A{},
											"in": bson.M{
												"$concatArrays": bson.A{
													"$$value",
													bson.M{
														"$ifNull": bson.A{
															bson.M{
																"$map": bson.M{
																	"input": bson.M{
																		"$ifNull": bson.A{
																			"$$this.sources",
																			bson.A{},
																		},
																	},
																	"as": "source",
																	"in": "$$source.id",
																},
															},
															bson.A{},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},

		bson.M{
			"$addFields": bson.M{
				"spec_pod_ids":    "$specs_data.pod_ids",
				"spec_secret_ids": "$specs_data.secret_ids",
				"spec_cert_ids":   "$specs_data.cert_ids",
				"spec_domain_ids": "$specs_data.domain_ids",
				"spec_unit_ids":   "$specs_data.unit_ids",
			},
		},

		bson.M{
			"$lookup": bson.M{
				"from":         "units",
				"localField":   "spec_unit_ids",
				"foreignField": "_id",
				"as":           "spec_id_units",
			},
		},

		bson.M{
			"$lookup": bson.M{
				"from":         "units",
				"localField":   "spec_pod_ids",
				"foreignField": "pod",
				"as":           "spec_pod_units",
			},
		},

		bson.M{
			"$lookup": bson.M{
				"from":         "pods",
				"localField":   "spec_pod_ids",
				"foreignField": "pod",
				"as":           "spec_pods",
			},
		},

		bson.M{
			"$lookup": bson.M{
				"from":         "secrets",
				"localField":   "spec_secret_ids",
				"foreignField": "_id",
				"as":           "spec_secrets",
			},
		},

		bson.M{
			"$lookup": bson.M{
				"from":         "certificates",
				"localField":   "spec_cert_ids",
				"foreignField": "_id",
				"as":           "spec_certs",
			},
		},

		bson.M{
			"$lookup": bson.M{
				"from":         "domains",
				"localField":   "spec_domain_ids",
				"foreignField": "_id",
				"as":           "spec_domains",
			},
		},

		bson.M{
			"$lookup": bson.M{
				"from":         "domains_records",
				"localField":   "spec_domain_ids",
				"foreignField": "domain",
				"as":           "spec_domains_records",
			},
		},

		bson.M{
			"$addFields": bson.M{
				"spec_deployment_id_ids": bson.M{
					"$reduce": bson.M{
						"input":        "$spec_id_units",
						"initialValue": bson.A{},
						"in": bson.M{
							"$concatArrays": bson.A{
								"$$value",
								bson.M{
									"$ifNull": bson.A{
										"$$this.deployments",
										bson.A{},
									},
								},
							},
						},
					},
				},
				"spec_deployment_pod_ids": bson.M{
					"$reduce": bson.M{
						"input":        "$spec_pod_units",
						"initialValue": bson.A{},
						"in": bson.M{
							"$concatArrays": bson.A{
								"$$value",
								bson.M{
									"$ifNull": bson.A{
										"$$this.deployments",
										bson.A{},
									},
								},
							},
						},
					},
				},
				"pod_deployment_ids": bson.M{
					"$reduce": bson.M{
						"input":        "$units",
						"initialValue": bson.A{},
						"in": bson.M{
							"$setUnion": bson.A{
								"$$value",
								bson.M{
									"$ifNull": bson.A{
										"$$this.deployments",
										bson.A{},
									},
								},
							},
						},
					},
				},
			},
		},

		bson.M{
			"$addFields": bson.M{
				"linked_deployment_ids": bson.M{
					"$setDifference": bson.A{
						bson.M{
							"$setUnion": bson.A{
								"$spec_deployment_id_ids",
								"$spec_deployment_pod_ids",
								"$pod_deployment_ids",
							},
						},
						"$deployment_ids",
					},
				},
			},
		},

		bson.M{
			"$lookup": bson.M{
				"from":         "deployments",
				"localField":   "linked_deployment_ids",
				"foreignField": "_id",
				"as":           "linked_deployments",
			},
		},

		bson.M{
			"$project": bson.M{
				"deployment_ids":       1,
				"pod_ids":              1,
				"unit_ids":             1,
				"spec_ids":             1,
				"deployments":          1,
				"pods":                 1,
				"units":                1,
				"specs":                1,
				"spec_pod_ids":         1,
				"spec_unit_ids":        1,
				"spec_secret_ids":      1,
				"spec_cert_ids":        1,
				"spec_domain_ids":      1,
				"spec_id_units":        1,
				"spec_pod_units":       1,
				"spec_secrets":         1,
				"spec_certs":           1,
				"spec_domains":         1,
				"spec_domains_records": 1,
				"linked_deployments":   1,
			},
		},
	}

	ctx := context.Background()
	cursor, err := db.Deployments().Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	result := &DeploymentsResult{}
	if cursor.Next(ctx) {
		err = cursor.Decode(result)
		if err != nil {
			return err
		}
	}

	deploymentsNode := map[primitive.ObjectID]*deployment.Deployment{}
	deploymentsReservedMap := map[primitive.ObjectID]*deployment.Deployment{}
	deploymentsDeployedMap := map[primitive.ObjectID]*deployment.Deployment{}
	deploymentsInactiveMap := map[primitive.ObjectID]*deployment.Deployment{}

	for _, deply := range result.Deployments {
		deploymentsNode[deply.Id] = deply

		switch deply.State {
		case deployment.Reserved:
			deploymentsReservedMap[deply.Id] = deply
		case deployment.Deployed:
			switch deply.Action {
			case deployment.Destroy, deployment.Archive, deployment.Restore:
				deploymentsInactiveMap[deply.Id] = deply
			default:
				deploymentsDeployedMap[deply.Id] = deply
			}
		case deployment.Archived:
			deploymentsInactiveMap[deply.Id] = deply
		}
	}
	p.deploymentsNode = deploymentsNode

	specsMap := map[primitive.ObjectID]*spec.Spec{}
	for _, spec := range result.Specs {
		specsMap[spec.Id] = spec
	}
	p.specsMap = specsMap

	specsCertsMap := map[primitive.ObjectID]*certificate.Certificate{}
	for _, specCert := range result.SpecCerts {
		specsCertsMap[specCert.Id] = specCert
	}
	p.specsCertsMap = specsCertsMap

	specsSecretsMap := map[primitive.ObjectID]*secret.Secret{}
	for _, specSecret := range result.SpecSecrets {
		specsSecretsMap[specSecret.Id] = specSecret
	}
	p.specsSecretsMap = specsSecretsMap

	specsPodsMap := map[primitive.ObjectID]*pod.Pod{}
	for _, specPod := range result.SpecPods {
		specsPodsMap[specPod.Id] = specPod
	}
	p.specsPodsMap = specsPodsMap

	specDomains := domain.PreloadedRecords(
		result.SpecDomains, result.SpecDomainsRecords)
	specsDomainsMap := map[primitive.ObjectID]*domain.Domain{}
	for _, specDomain := range specDomains {
		specsDomainsMap[specDomain.Id] = specDomain
	}
	p.specsDomainsMap = specsDomainsMap

	specUnitsIds := set.NewSet()
	specsUnitsMap := map[primitive.ObjectID]*unit.Unit{}
	specsPodUnitsMap := map[primitive.ObjectID][]*unit.Unit{}
	for _, specUnit := range result.SpecIdUnits {
		if specUnitsIds.Contains(specUnit.Id) {
			continue
		}
		specUnitsIds.Add(specUnit.Id)

		specsUnitsMap[specUnit.Id] = specUnit
		specsPodUnitsMap[specUnit.Pod] = append(
			specsPodUnitsMap[specUnit.Pod], specUnit)
	}
	for _, specUnit := range result.SpecPodUnits {
		if specUnitsIds.Contains(specUnit.Id) {
			continue
		}
		specUnitsIds.Add(specUnit.Id)

		specsUnitsMap[specUnit.Id] = specUnit
		specsPodUnitsMap[specUnit.Pod] = append(
			specsPodUnitsMap[specUnit.Pod], specUnit)
	}
	p.specsUnitsMap = specsUnitsMap
	p.specsPodUnitsMap = specsPodUnitsMap

	for _, deply := range result.LinkedDeployments {
		switch deply.State {
		case deployment.Reserved:
			deploymentsReservedMap[deply.Id] = deply
		case deployment.Deployed:
			switch deply.Action {
			case deployment.Destroy, deployment.Archive, deployment.Restore:
				deploymentsInactiveMap[deply.Id] = deply
			default:
				deploymentsDeployedMap[deply.Id] = deply
			}
		case deployment.Archived:
			deploymentsInactiveMap[deply.Id] = deply
		}
	}
	p.deploymentsReservedMap = deploymentsReservedMap
	p.deploymentsDeployedMap = deploymentsDeployedMap
	p.deploymentsInactiveMap = deploymentsInactiveMap

	podsMap := map[primitive.ObjectID]*pod.Pod{}
	for _, pd := range result.Pods {
		podsMap[pd.Id] = pd
	}
	p.podsMap = podsMap

	unitsMap := map[primitive.ObjectID]*unit.Unit{}
	for _, unt := range result.Units {
		unitsMap[unt.Id] = unt
	}
	p.unitsMap = unitsMap

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
