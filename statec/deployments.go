package statec

import (
	"fmt"
	"sync"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/unit"
	"github.com/sirupsen/logrus"
)

var (
	Deployments    = &DeploymentsState{}
	DeploymentsPkg = NewPackage(Deployments)
)

type DeploymentsResult struct {
	DeploymentIds []bson.ObjectID `bson:"deployment_ids"`
	PodIds        []bson.ObjectID `bson:"pod_ids"`
	UnitIds       []bson.ObjectID `bson:"unit_ids"`
	SpecIds       []bson.ObjectID `bson:"spec_ids"`

	Deployments []*deployment.Deployment `bson:"deployments"`
	Pods        []*pod.Pod               `bson:"pods"`
	Units       []*unit.Unit             `bson:"units"`
	Specs       []*spec.Spec             `bson:"specs"`

	SpecPodIds    []bson.ObjectID `bson:"spec_pod_ids"`
	SpecUnitIds   []bson.ObjectID `bson:"spec_unit_ids"`
	SpecSecretIds []bson.ObjectID `bson:"spec_secret_ids"`
	SpecCertIds   []bson.ObjectID `bson:"spec_cert_ids"`
	SpecDomainIds []bson.ObjectID `bson:"spec_domain_ids"`

	SpecIdUnits        []*unit.Unit               `bson:"spec_id_units"`
	SpecPodUnits       []*unit.Unit               `bson:"spec_pod_units"`
	SpecPods           []*pod.Pod                 `bson:"spec_pods"`
	SpecSecrets        []*secret.Secret           `bson:"spec_secrets"`
	SpecCerts          []*certificate.Certificate `bson:"spec_certs"`
	SpecDomains        []*domain.Domain           `bson:"spec_domains"`
	SpecDomainsRecords []*domain.Record           `bson:"spec_domains_records"`

	LinkedDeployments []*deployment.Deployment `bson:"linked_deployments"`
}

type idSet map[bson.ObjectID]struct{}

func (s idSet) has(id bson.ObjectID) bool {
	_, ok := s[id]
	return ok
}

func (s idSet) add(ids ...bson.ObjectID) {
	for _, id := range ids {
		s[id] = struct{}{}
	}
}

type deploymentsKnown struct {
	deployments idSet
	pods        idSet
	units       idSet
	specs       idSet
	secrets     idSet
	certs       idSet
	domains     idSet
	records     idSet
}

func newDeploymentsKnown(result *DeploymentsResult) (
	known *deploymentsKnown) {

	known = &deploymentsKnown{
		deployments: idSet{},
		pods:        idSet{},
		units:       idSet{},
		specs:       idSet{},
		secrets:     idSet{},
		certs:       idSet{},
		domains:     idSet{},
		records:     idSet{},
	}

	known.deployments.add(result.DeploymentIds...)
	for _, deply := range result.LinkedDeployments {
		known.deployments.add(deply.Id)
	}

	known.pods.add(result.PodIds...)
	known.pods.add(result.SpecPodIds...)

	known.units.add(result.UnitIds...)
	known.units.add(result.SpecUnitIds...)
	for _, unt := range result.SpecIdUnits {
		known.units.add(unt.Id)
	}
	for _, unt := range result.SpecPodUnits {
		known.units.add(unt.Id)
	}

	known.specs.add(result.SpecIds...)
	known.secrets.add(result.SpecSecretIds...)
	known.certs.add(result.SpecCertIds...)
	known.domains.add(result.SpecDomainIds...)

	for _, rec := range result.SpecDomainsRecords {
		known.records.add(rec.Id)
	}

	return
}

type DeploymentsState struct {
	lock       sync.Mutex
	dirty      bool
	loaded     bool
	running    bool
	generation uint64
	result     bson.Raw
	known      *deploymentsKnown

	podsMap                map[bson.ObjectID]*pod.Pod
	unitsMap               map[bson.ObjectID]*unit.Unit
	specsMap               map[bson.ObjectID]*spec.Spec
	specsPodsMap           map[bson.ObjectID]*pod.Pod
	specsPodUnitsMap       map[bson.ObjectID][]*unit.Unit
	specsUnitsMap          map[bson.ObjectID]*unit.Unit
	specsDeploymentsMap    map[bson.ObjectID]*deployment.Deployment
	specsDomainsMap        map[bson.ObjectID]*domain.Domain
	specsSecretsMap        map[bson.ObjectID]*secret.Secret
	specsCertsMap          map[bson.ObjectID]*certificate.Certificate
	deploymentsNode        map[bson.ObjectID]*deployment.Deployment
	deploymentsReservedMap map[bson.ObjectID]*deployment.Deployment
	deploymentsDeployedMap map[bson.ObjectID]*deployment.Deployment
	deploymentsInactiveMap map[bson.ObjectID]*deployment.Deployment
}

func (p *DeploymentsState) Pod(pdId bson.ObjectID) *pod.Pod {
	return p.podsMap[pdId]
}

func (p *DeploymentsState) PodsMap() map[bson.ObjectID]*pod.Pod {
	return p.podsMap
}

func (p *DeploymentsState) Unit(untId bson.ObjectID) *unit.Unit {
	return p.unitsMap[untId]
}

func (p *DeploymentsState) UnitsMap() map[bson.ObjectID]*unit.Unit {
	return p.unitsMap
}

func (p *DeploymentsState) Spec(commitId bson.ObjectID) *spec.Spec {
	return p.specsMap[commitId]
}

func (p *DeploymentsState) SpecsMap() map[bson.ObjectID]*spec.Spec {
	return p.specsMap
}

func (p *DeploymentsState) SpecPod(pdId bson.ObjectID) *pod.Pod {
	return p.specsPodsMap[pdId]
}

func (p *DeploymentsState) SpecPodUnits(pdId bson.ObjectID) []*unit.Unit {
	return p.specsPodUnitsMap[pdId]
}

func (p *DeploymentsState) SpecUnit(unitId bson.ObjectID) *unit.Unit {
	return p.specsUnitsMap[unitId]
}

func (p *DeploymentsState) SpecsUnitsMap() map[bson.ObjectID]*unit.Unit {
	return p.specsUnitsMap
}

func (p *DeploymentsState) SpecDomain(domnId bson.ObjectID) *domain.Domain {
	return p.specsDomainsMap[domnId]
}

func (p *DeploymentsState) SpecSecret(secrID bson.ObjectID) *secret.Secret {
	return p.specsSecretsMap[secrID]
}

func (p *DeploymentsState) SpecCert(
	certId bson.ObjectID) *certificate.Certificate {

	return p.specsCertsMap[certId]
}

func (p *DeploymentsState) SpecCertMap() map[bson.ObjectID]*certificate.Certificate {
	return p.specsCertsMap
}

func (p *DeploymentsState) DeploymentsNode() map[bson.ObjectID]*deployment.Deployment {
	return p.deploymentsNode
}

func (p *DeploymentsState) DeploymentReserved(deplyId bson.ObjectID) *deployment.Deployment {
	return p.deploymentsReservedMap[deplyId]
}

func (p *DeploymentsState) DeploymentsReserved() (
	deplys map[bson.ObjectID]*deployment.Deployment) {

	deplys = p.deploymentsReservedMap
	return
}

func (p *DeploymentsState) DeploymentDeployed(deplyId bson.ObjectID) *deployment.Deployment {
	return p.deploymentsDeployedMap[deplyId]
}

func (p *DeploymentsState) DeploymentsDeployed() (
	deplys map[bson.ObjectID]*deployment.Deployment) {

	deplys = p.deploymentsDeployedMap
	return
}

func (p *DeploymentsState) DeploymentsDestroy() (
	deplys map[bson.ObjectID]*deployment.Deployment) {

	deplys = p.deploymentsInactiveMap
	return
}

func (p *DeploymentsState) DeploymentInactive(deplyId bson.ObjectID) *deployment.Deployment {
	return p.deploymentsInactiveMap[deplyId]
}

func (p *DeploymentsState) DeploymentsInactive() (
	deplys map[bson.ObjectID]*deployment.Deployment) {

	deplys = p.deploymentsInactiveMap
	return
}

func (p *DeploymentsState) Deployment(deplyId bson.ObjectID) (
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

func (p *DeploymentsState) Resync() {
	p.lock.Lock()
	p.dirty = true
	p.lock.Unlock()
}

func (p *DeploymentsState) matchDeployment(evt *Event,
	known *deploymentsKnown) bool {

	ndeId, ok := evt.DocObjectId("node")
	if ok && ndeId == node.Self.Id {
		return true
	}

	untId, ok := evt.DocObjectId("unit")
	if ok && known.units.has(untId) {
		return true
	}

	pdId, ok := evt.DocObjectId("pod")
	if ok && known.pods.has(pdId) {
		return true
	}

	return false
}

func (p *DeploymentsState) matchUnit(evt *Event,
	known *deploymentsKnown) bool {

	pdId, ok := evt.DocObjectId("pod")
	return ok && known.pods.has(pdId)
}

func (p *DeploymentsState) matchRecord(evt *Event,
	known *deploymentsKnown) bool {

	domnId, ok := evt.DocObjectId("domain")
	return ok && known.domains.has(domnId)
}

func (p *DeploymentsState) reason(evt *Event, id bson.ObjectID,
	known *deploymentsKnown) string {

	pick := func(ids idSet, match func() bool, matchReason string) string {
		if ids.has(id) {
			return "known"
		}
		if match != nil && match() {
			return matchReason
		}
		return ""
	}

	switch evt.Ns.Coll {
	case "deployments":
		switch evt.OperationType {
		case opInsert, opReplace:
			return pick(known.deployments, func() bool {
				return p.matchDeployment(evt, known)
			}, "match")
		case opUpdate:
			return pick(known.deployments, func() bool {
				return evt.Touches("node", "unit", "pod")
			}, "scope_key")
		case opDelete:
			return pick(known.deployments, nil, "")
		}
	case "pods":
		return pick(known.pods, nil, "")
	case "units":
		switch evt.OperationType {
		case opInsert, opReplace:
			return pick(known.units, func() bool {
				return p.matchUnit(evt, known)
			}, "match")
		case opUpdate:
			return pick(known.units, func() bool {
				return evt.Touches("pod")
			}, "scope_key")
		case opDelete:
			return pick(known.units, nil, "")
		}
	case "specs":
		return pick(known.specs, nil, "")
	case "secrets":
		return pick(known.secrets, nil, "")
	case "certificates":
		return pick(known.certs, nil, "")
	case "domains":
		return pick(known.domains, nil, "")
	case "domains_records":
		switch evt.OperationType {
		case opInsert, opReplace:
			return pick(known.records, func() bool {
				return p.matchRecord(evt, known)
			}, "match")
		case opUpdate:
			return pick(known.records, func() bool {
				return evt.Touches("domain")
			}, "scope_key")
		case opDelete:
			return pick(known.records, nil, "")
		}
	}

	return ""
}

func (p *DeploymentsState) Handle(evt *Event) {
	if evt.IsStructural() {
		if settings.System.StateDebug {
			logrus.WithFields(logrus.Fields{
				"collection": evt.Ns.Coll,
				"operation":  evt.OperationType,
			}).Info("statec: Deployments resync")
		}
		p.Resync()
		return
	}

	id, ok := evt.DocId()
	if !ok {
		p.Resync()
		return
	}

	p.lock.Lock()

	reason := ""
	if p.running {
		reason = "running"
	} else if p.known == nil {
		reason = "unloaded"
	} else {
		reason = p.reason(evt, id, p.known)
	}

	if reason != "" {
		p.dirty = true
	}

	p.lock.Unlock()

	if reason != "" && settings.System.StateDebug {
		logrus.WithFields(logrus.Fields{
			"collection": evt.Ns.Coll,
			"operation":  evt.OperationType,
			"id":         id.Hex(),
			"reason":     reason,
		}).Info("statec: Deployments dirty")
	}
}

func (p *DeploymentsState) aggregate(db *database.Database) (
	raw bson.Raw, err error) {

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
				"foreignField": "_id",
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
				"spec_pods":            1,
				"spec_secrets":         1,
				"spec_certs":           1,
				"spec_domains":         1,
				"spec_domains_records": 1,
				"linked_deployments":   1,
			},
		},
	}

	cursor, err := db.Deployments().Aggregate(db, pipeline)
	if err != nil {
		return
	}
	defer cursor.Close(db)

	if cursor.Next(db) {
		raw = bson.Raw(append([]byte(nil), cursor.Current...))
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (p *DeploymentsState) Refresh(
	pkg *Package, db *database.Database) (err error) {

	healthy, generation := Watch.Status()

	p.lock.Lock()
	reason := ""
	switch {
	case !p.loaded:
		reason = "initial"
	case p.dirty:
		reason = "dirty"
	case !healthy:
		reason = "unhealthy"
	case generation != p.generation:
		reason = "generation"
	}
	run := reason != ""
	if run {
		p.dirty = false
		p.running = true
	}
	p.lock.Unlock()

	var raw bson.Raw
	start := time.Now()
	if run {
		raw, err = p.aggregate(db)
		if err != nil {
			p.lock.Lock()
			p.dirty = true
			p.running = false
			p.lock.Unlock()
			return
		}
	} else {
		p.lock.Lock()
		raw = p.result
		p.lock.Unlock()
	}

	result := &DeploymentsResult{}
	if raw != nil {
		err = bson.Unmarshal(raw, result)
		if err != nil {
			err = &errortypes.ParseError{
				errors.Wrap(err, "statec: Failed to decode deployments"),
			}

			if run {
				p.lock.Lock()
				p.dirty = true
				p.running = false
				p.lock.Unlock()
			}
			return
		}
	}

	if run {
		known := newDeploymentsKnown(result)

		p.lock.Lock()
		p.result = raw
		p.known = known
		p.loaded = true
		p.running = false
		p.generation = generation
		p.lock.Unlock()

		if settings.System.StateDebug {
			logrus.WithFields(logrus.Fields{
				"reason":      reason,
				"deployments": len(result.Deployments),
				"linked":      len(result.LinkedDeployments),
				"pods":        len(known.pods),
				"units":       len(known.units),
				"specs":       len(result.Specs),
				"domains":     len(known.domains),
				"generation":  int(generation),
				"duration":    fmt.Sprintf("%v", time.Since(start)),
			}).Info("statec: Deployments aggregated")
		}
	}

	deploymentsNode := map[bson.ObjectID]*deployment.Deployment{}
	deploymentsReservedMap := map[bson.ObjectID]*deployment.Deployment{}
	deploymentsDeployedMap := map[bson.ObjectID]*deployment.Deployment{}
	deploymentsInactiveMap := map[bson.ObjectID]*deployment.Deployment{}

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

	specsMap := map[bson.ObjectID]*spec.Spec{}
	for _, spec := range result.Specs {
		specsMap[spec.Id] = spec
	}
	p.specsMap = specsMap

	specsCertsMap := map[bson.ObjectID]*certificate.Certificate{}
	for _, specCert := range result.SpecCerts {
		specsCertsMap[specCert.Id] = specCert
	}
	p.specsCertsMap = specsCertsMap

	specsSecretsMap := map[bson.ObjectID]*secret.Secret{}
	for _, specSecret := range result.SpecSecrets {
		specsSecretsMap[specSecret.Id] = specSecret
	}
	p.specsSecretsMap = specsSecretsMap

	specsPodsMap := map[bson.ObjectID]*pod.Pod{}
	for _, specPod := range result.SpecPods {
		specsPodsMap[specPod.Id] = specPod
	}
	p.specsPodsMap = specsPodsMap

	specDomains := domain.PreloadedRecords(
		result.SpecDomains, result.SpecDomainsRecords)
	specsDomainsMap := map[bson.ObjectID]*domain.Domain{}
	for _, specDomain := range specDomains {
		specsDomainsMap[specDomain.Id] = specDomain
	}
	p.specsDomainsMap = specsDomainsMap

	specUnitsIds := set.NewSet()
	specsUnitsMap := map[bson.ObjectID]*unit.Unit{}
	specsPodUnitsMap := map[bson.ObjectID][]*unit.Unit{}
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

	podsMap := map[bson.ObjectID]*pod.Pod{}
	for _, pd := range result.Pods {
		podsMap[pd.Id] = pd
	}
	p.podsMap = podsMap

	unitsMap := map[bson.ObjectID]*unit.Unit{}
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
	st.SpecsMap = p.SpecsMap
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

func init() {
	for _, coll := range []string{
		"deployments",
		"pods",
		"units",
		"specs",
		"secrets",
		"certificates",
		"domains",
		"domains_records",
	} {
		Watch.Register(coll, Deployments)
	}
}
