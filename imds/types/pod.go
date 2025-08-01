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
	CloudPublicIps           []string           `json:"cloud_public_ips"`
	CloudPublicIps6          []string           `json:"cloud_public_ips6"`
	CloudPrivateIps          []string           `json:"cloud_private_ips"`
	HealthyCloudPublicIps    []string           `json:"healthy_cloud_public_ips"`
	HealthyCloudPublicIps6   []string           `json:"healthy_cloud_public_ips6"`
	HealthyCloudPrivateIps   []string           `json:"healthy_cloud_private_ips"`
	UnhealthyCloudPublicIps  []string           `json:"unhealthy_cloud_public_ips"`
	UnhealthyCloudPublicIps6 []string           `json:"unhealthy_cloud_public_ips6"`
	UnhealthyCloudPrivateIps []string           `json:"unhealthy_cloud_private_ips"`
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
				CloudPublicIps:           []string{},
				CloudPublicIps6:          []string{},
				CloudPrivateIps:          []string{},
				HealthyCloudPublicIps:    []string{},
				HealthyCloudPublicIps6:   []string{},
				HealthyCloudPrivateIps:   []string{},
				UnhealthyCloudPublicIps:  []string{},
				UnhealthyCloudPublicIps6: []string{},
				UnhealthyCloudPrivateIps: []string{},
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

					cloudPublicIps := data.CloudPublicIps
					if cloudPublicIps == nil {
						cloudPublicIps = []string{}
					}
					cloudPublicIps6 := data.CloudPublicIps6
					if cloudPublicIps6 == nil {
						cloudPublicIps6 = []string{}
					}
					cloudPrivateIps := data.CloudPrivateIps
					if cloudPrivateIps == nil {
						cloudPrivateIps = []string{}
					}
					healthyCloudPublicIps := []string{}
					unhealthyCloudPublicIps := []string{}
					healthyCloudPublicIps6 := []string{}
					unhealthyCloudPublicIps6 := []string{}
					healthyCloudPrivateIps := []string{}
					unhealthyCloudPrivateIps := []string{}

					if deply.IsHealthy() {
						healthyPublicIps = publicIps
						healthyPublicIps6 = publicIps6
						healthyPrivateIps = privateIps
						healthyPrivateIps6 = privateIps6
						healthyCloudPublicIps = cloudPublicIps
						healthyCloudPublicIps6 = cloudPublicIps6
						healthyCloudPrivateIps = cloudPrivateIps
					} else {
						unhealthyPublicIps = publicIps
						unhealthyPublicIps6 = publicIps6
						unhealthyPrivateIps = privateIps
						unhealthyPrivateIps6 = privateIps6
						unhealthyCloudPublicIps = cloudPublicIps
						unhealthyCloudPublicIps6 = cloudPublicIps6
						unhealthyCloudPrivateIps = cloudPrivateIps
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

					unit.CloudPublicIps = append(
						unit.CloudPublicIps, cloudPublicIps...)
					unit.HealthyCloudPublicIps = append(
						unit.HealthyCloudPublicIps,
						healthyCloudPublicIps...,
					)
					unit.UnhealthyCloudPublicIps = append(
						unit.UnhealthyCloudPublicIps,
						unhealthyCloudPublicIps...,
					)
					unit.CloudPublicIps6 = append(
						unit.CloudPublicIps6, cloudPublicIps6...)
					unit.HealthyCloudPublicIps6 = append(
						unit.HealthyCloudPublicIps6,
						healthyCloudPublicIps6...,
					)
					unit.UnhealthyCloudPublicIps6 = append(
						unit.UnhealthyCloudPublicIps6,
						unhealthyCloudPublicIps6...,
					)
					unit.CloudPrivateIps = append(
						unit.CloudPrivateIps, cloudPrivateIps...)
					unit.HealthyCloudPrivateIps = append(
						unit.HealthyCloudPrivateIps,
						healthyCloudPrivateIps...,
					)
					unit.UnhealthyCloudPrivateIps = append(
						unit.UnhealthyCloudPrivateIps,
						unhealthyCloudPrivateIps...,
					)
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
