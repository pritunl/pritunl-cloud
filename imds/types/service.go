package types

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/unit"
)

type Pod struct {
	Id    primitive.ObjectID `json:"id"`
	Name  string             `json:"name"`
	Units []*Unit            `json:"units"`
}

type Unit struct {
	Id                        primitive.ObjectID `json:"id"`
	Name                      string             `json:"name"`
	Kind                      string             `json:"kind"`
	Count                     int                `json:"count"`
	PublicIps                 []string           `json:"public_ips"`
	PublicIps6                []string           `json:"public_ips6"`
	HealthyPublicIps          []string           `json:"healthy_public_ips"`
	HealthyPublicIps6         []string           `json:"healthy_public_ips6"`
	UnhealthyPublicIps        []string           `json:"unhealthy_public_ips"`
	UnhealthyPublicIps6       []string           `json:"unhealthy_public_ips6"`
	PrivateIps                []string           `json:"private_ips"`
	PrivateIps6               []string           `json:"private_ips6"`
	HealthyPrivateIps         []string           `json:"healthy_private_ips"`
	HealthyPrivateIps6        []string           `json:"healthy_private_ips6"`
	UnhealthyPrivateIps       []string           `json:"unhealthy_private_ips"`
	UnhealthyPrivateIps6      []string           `json:"unhealthy_private_ips6"`
	OraclePublicIps           []string           `json:"oracle_public_ips"`
	OraclePublicIps6          []string           `json:"oracle_public_ips6"`
	OraclePrivateIps          []string           `json:"oracle_private_ips"`
	HealthyOraclePublicIps    []string           `json:"healthy_oracle_public_ips"`
	HealthyOraclePublicIps6   []string           `json:"healthy_oracle_public_ips6"`
	HealthyOraclePrivateIps   []string           `json:"healthy_oracle_private_ips"`
	UnhealthyOraclePublicIps  []string           `json:"unhealthy_oracle_public_ips"`
	UnhealthyOraclePublicIps6 []string           `json:"unhealthy_oracle_public_ips6"`
	UnhealthyOraclePrivateIps []string           `json:"unhealthy_oracle_private_ips"`
	Deployments               []*Deployment      `json:"deployments"`
}

type Deployment struct {
	Id                        primitive.ObjectID `json:"id"`
	Pod                       primitive.ObjectID `json:"pod"`
	Unit                      primitive.ObjectID `json:"unit"`
	Spec                      primitive.ObjectID `json:"spec"`
	Kind                      string             `json:"kind"`
	State                     string             `json:"state"`
	Action                    string             `json:"action"`
	Node                      primitive.ObjectID `json:"node"`
	Instance                  primitive.ObjectID `json:"instance"`
	PublicIps                 []string           `json:"public_ips"`
	PublicIps6                []string           `json:"public_ips6"`
	HealthyPublicIps          []string           `json:"healthy_public_ips"`
	HealthyPublicIps6         []string           `json:"healthy_public_ips6"`
	UnhealthyPublicIps        []string           `json:"unhealthy_public_ips"`
	UnhealthyPublicIps6       []string           `json:"unhealthy_public_ips6"`
	PrivateIps                []string           `json:"private_ips"`
	PrivateIps6               []string           `json:"private_ips6"`
	HealthyPrivateIps         []string           `json:"healthy_private_ips"`
	HealthyPrivateIps6        []string           `json:"healthy_private_ips6"`
	UnhealthyPrivateIps       []string           `json:"unhealthy_private_ips"`
	UnhealthyPrivateIps6      []string           `json:"unhealthy_private_ips6"`
	OraclePublicIps           []string           `json:"oracle_public_ips"`
	OraclePublicIps6          []string           `json:"oracle_public_ips6"`
	OraclePrivateIps          []string           `json:"oracle_private_ips"`
	HealthyOraclePublicIps    []string           `json:"healthy_oracle_public_ips"`
	HealthyOraclePublicIps6   []string           `json:"healthy_oracle_public_ips6"`
	HealthyOraclePrivateIps   []string           `json:"healthy_oracle_private_ips"`
	UnhealthyOraclePublicIps  []string           `json:"unhealthy_oracle_public_ips"`
	UnhealthyOraclePublicIps6 []string           `json:"unhealthy_oracle_public_ips6"`
	UnhealthyOraclePrivateIps []string           `json:"unhealthy_oracle_private_ips"`
}

func NewPods(pods []*pod.Pod, podUnitsMap map[primitive.ObjectID][]*unit.Unit,
	deployments map[primitive.ObjectID]*deployment.Deployment) []*Pod {

	datas := []*Pod{}

	for _, pd := range pods {
		if pd == nil {
			continue
		}

		units := []*Unit{}
		for _, pdUnit := range podUnitsMap[pd.Id] {
			unit := &Unit{
				Id:                        pdUnit.Id,
				Name:                      pdUnit.Name,
				Kind:                      pdUnit.Kind,
				Count:                     pdUnit.Count,
				PublicIps:                 []string{},
				PublicIps6:                []string{},
				HealthyPublicIps:          []string{},
				HealthyPublicIps6:         []string{},
				UnhealthyPublicIps:        []string{},
				UnhealthyPublicIps6:       []string{},
				PrivateIps:                []string{},
				PrivateIps6:               []string{},
				HealthyPrivateIps:         []string{},
				HealthyPrivateIps6:        []string{},
				UnhealthyPrivateIps:       []string{},
				UnhealthyPrivateIps6:      []string{},
				OraclePublicIps:           []string{},
				OraclePublicIps6:          []string{},
				OraclePrivateIps:          []string{},
				HealthyOraclePublicIps:    []string{},
				HealthyOraclePublicIps6:   []string{},
				HealthyOraclePrivateIps:   []string{},
				UnhealthyOraclePublicIps:  []string{},
				UnhealthyOraclePublicIps6: []string{},
				UnhealthyOraclePrivateIps: []string{},
				Deployments:               []*Deployment{},
			}

			for _, unitDeplyId := range pdUnit.Deployments {
				deply := deployments[unitDeplyId]

				if deply != nil {
					data := deply.InstanceData
					if data == nil {
						data = &deployment.InstanceData{}
					}

					publicIps := data.PublicIps
					if publicIps == nil {
						publicIps = []string{}
					}
					publicIps6 := data.PublicIps6
					if publicIps6 == nil {
						publicIps6 = []string{}
					}
					healthyPublicIps := []string{}
					unhealthyPublicIps := []string{}
					healthyPublicIps6 := []string{}
					unhealthyPublicIps6 := []string{}

					privateIps := data.PrivateIps
					if privateIps == nil {
						privateIps = []string{}
					}
					privateIps6 := data.PrivateIps6
					if privateIps6 == nil {
						privateIps6 = []string{}
					}
					healthyPrivateIps := []string{}
					unhealthyPrivateIps := []string{}
					healthyPrivateIps6 := []string{}
					unhealthyPrivateIps6 := []string{}

					oraclePublicIps := data.OraclePublicIps
					if oraclePublicIps == nil {
						oraclePublicIps = []string{}
					}
					oraclePublicIps6 := data.OraclePublicIps6
					if oraclePublicIps6 == nil {
						oraclePublicIps6 = []string{}
					}
					oraclePrivateIps := data.OraclePrivateIps
					if oraclePrivateIps == nil {
						oraclePrivateIps = []string{}
					}
					healthyOraclePublicIps := []string{}
					unhealthyOraclePublicIps := []string{}
					healthyOraclePublicIps6 := []string{}
					unhealthyOraclePublicIps6 := []string{}
					healthyOraclePrivateIps := []string{}
					unhealthyOraclePrivateIps := []string{}

					if deply.IsHealthy() {
						healthyPublicIps = publicIps
						healthyPublicIps6 = publicIps6
						healthyPrivateIps = privateIps
						healthyPrivateIps6 = privateIps6
						healthyOraclePublicIps = oraclePublicIps
						healthyOraclePublicIps6 = oraclePublicIps6
						healthyOraclePrivateIps = oraclePrivateIps
					} else {
						unhealthyPublicIps = publicIps
						unhealthyPublicIps6 = publicIps6
						unhealthyPrivateIps = privateIps
						unhealthyPrivateIps6 = privateIps6
						unhealthyOraclePublicIps = oraclePublicIps
						unhealthyOraclePublicIps6 = oraclePublicIps6
						unhealthyOraclePrivateIps = oraclePrivateIps
					}

					unit.PublicIps = append(
						unit.PublicIps, publicIps...)
					unit.PublicIps6 = append(
						unit.PublicIps6, publicIps6...)
					unit.HealthyPublicIps = append(
						unit.HealthyPublicIps, healthyPublicIps...)
					unit.HealthyPublicIps6 = append(
						unit.HealthyPublicIps6, healthyPublicIps6...)
					unit.UnhealthyPublicIps = append(
						unit.UnhealthyPublicIps, unhealthyPublicIps...)
					unit.UnhealthyPublicIps6 = append(
						unit.UnhealthyPublicIps6, unhealthyPublicIps6...)

					unit.PrivateIps = append(
						unit.PrivateIps, privateIps...)
					unit.PrivateIps6 = append(
						unit.PrivateIps6, privateIps6...)
					unit.HealthyPrivateIps = append(
						unit.HealthyPrivateIps, healthyPrivateIps...)
					unit.HealthyPrivateIps6 = append(
						unit.HealthyPrivateIps6, healthyPrivateIps6...)
					unit.UnhealthyPrivateIps = append(
						unit.UnhealthyPrivateIps, unhealthyPrivateIps...)
					unit.UnhealthyPrivateIps6 = append(
						unit.UnhealthyPrivateIps6, unhealthyPrivateIps6...)

					unit.OraclePublicIps = append(
						unit.OraclePublicIps, oraclePublicIps...)
					unit.HealthyOraclePublicIps = append(
						unit.HealthyOraclePublicIps,
						healthyOraclePublicIps...,
					)
					unit.UnhealthyOraclePublicIps = append(
						unit.UnhealthyOraclePublicIps,
						unhealthyOraclePublicIps...,
					)
					unit.OraclePublicIps6 = append(
						unit.OraclePublicIps6, oraclePublicIps6...)
					unit.HealthyOraclePublicIps6 = append(
						unit.HealthyOraclePublicIps6,
						healthyOraclePublicIps6...,
					)
					unit.UnhealthyOraclePublicIps6 = append(
						unit.UnhealthyOraclePublicIps6,
						unhealthyOraclePublicIps6...,
					)
					unit.OraclePrivateIps = append(
						unit.OraclePrivateIps, oraclePrivateIps...)
					unit.HealthyOraclePrivateIps = append(
						unit.HealthyOraclePrivateIps,
						healthyOraclePrivateIps...,
					)
					unit.UnhealthyOraclePrivateIps = append(
						unit.UnhealthyOraclePrivateIps,
						unhealthyOraclePrivateIps...,
					)

					unit.Deployments = append(unit.Deployments, &Deployment{
						Id:                        deply.Id,
						Pod:                       deply.Pod,
						Unit:                      deply.Unit,
						Spec:                      deply.Spec,
						Kind:                      deply.Kind,
						State:                     deply.State,
						Action:                    deply.Action,
						Node:                      deply.Node,
						Instance:                  deply.Instance,
						PublicIps:                 publicIps,
						PublicIps6:                publicIps6,
						HealthyPublicIps:          healthyPublicIps,
						HealthyPublicIps6:         healthyPublicIps6,
						UnhealthyPublicIps:        unhealthyPublicIps,
						UnhealthyPublicIps6:       unhealthyPublicIps6,
						PrivateIps:                privateIps,
						PrivateIps6:               privateIps6,
						HealthyPrivateIps:         healthyPrivateIps,
						HealthyPrivateIps6:        healthyPrivateIps6,
						UnhealthyPrivateIps:       unhealthyPrivateIps,
						UnhealthyPrivateIps6:      unhealthyPrivateIps6,
						OraclePublicIps:           oraclePublicIps,
						HealthyOraclePublicIps:    healthyOraclePublicIps,
						UnhealthyOraclePublicIps:  unhealthyOraclePublicIps,
						OraclePublicIps6:          oraclePublicIps6,
						HealthyOraclePublicIps6:   healthyOraclePublicIps6,
						UnhealthyOraclePublicIps6: unhealthyOraclePublicIps6,
						OraclePrivateIps:          oraclePrivateIps,
						HealthyOraclePrivateIps:   healthyOraclePrivateIps,
						UnhealthyOraclePrivateIps: unhealthyOraclePrivateIps,
					})
				}
			}

			units = append(units, unit)
		}

		data := &Pod{
			Id:    pd.Id,
			Name:  pd.Name,
			Units: units,
		}

		datas = append(datas, data)
	}

	return datas
}
