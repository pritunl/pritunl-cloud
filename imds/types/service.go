package types

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/service"
)

type Service struct {
	Id    primitive.ObjectID `json:"id"`
	Name  string             `json:"name"`
	Units []*Unit            `json:"units"`
}

type Unit struct {
	Id               primitive.ObjectID `json:"id"`
	Name             string             `json:"name"`
	Kind             string             `json:"kind"`
	Count            int                `json:"count"`
	PublicIps        []string           `json:"public_ips"`
	PublicIps6       []string           `json:"public_ips6"`
	PrivateIps       []string           `json:"private_ips"`
	PrivateIps6      []string           `json:"private_ips6"`
	OraclePrivateIps []string           `json:"oracle_private_ips"`
	OraclePublicIps  []string           `json:"oracle_public_ips"`
	Deployments      []*Deployment      `json:"deployments"`
}

type Deployment struct {
	Id               primitive.ObjectID `json:"id"`
	Service          primitive.ObjectID `json:"service"`
	Unit             primitive.ObjectID `json:"unit"`
	Spec             primitive.ObjectID `json:"spec"`
	Kind             string             `json:"kind"`
	State            string             `json:"state"`
	Node             primitive.ObjectID `json:"node"`
	Instance         primitive.ObjectID `json:"instance"`
	PublicIps        []string           `json:"public_ips"`
	PublicIps6       []string           `json:"public_ips6"`
	PrivateIps       []string           `json:"private_ips"`
	PrivateIps6      []string           `json:"private_ips6"`
	OraclePrivateIps []string           `json:"oracle_private_ips"`
	OraclePublicIps  []string           `json:"oracle_public_ips"`
}

func NewServices(services []*service.Service,
	deployments map[primitive.ObjectID]*deployment.Deployment) []*Service {

	datas := []*Service{}

	for _, servc := range services {
		if servc == nil {
			continue
		}

		units := []*Unit{}
		for _, srvcUnit := range servc.Units {
			unit := &Unit{
				Id:               srvcUnit.Id,
				Name:             srvcUnit.Name,
				Kind:             srvcUnit.Kind,
				Count:            srvcUnit.Count,
				PublicIps:        []string{},
				PublicIps6:       []string{},
				PrivateIps:       []string{},
				PrivateIps6:      []string{},
				OraclePrivateIps: []string{},
				OraclePublicIps:  []string{},
				Deployments:      []*Deployment{},
			}

			for _, unitDeply := range srvcUnit.Deployments {
				deply := deployments[unitDeply.Id]
				if deply != nil {
					if deply.PublicIps != nil {
						unit.PublicIps = append(
							unit.PublicIps, deply.PublicIps...)
					}
					if deply.PublicIps6 != nil {
						unit.PublicIps6 = append(
							unit.PublicIps6, deply.PublicIps6...)
					}
					if deply.PrivateIps != nil {
						unit.PrivateIps = append(
							unit.PrivateIps, deply.PrivateIps...)
					}
					if deply.PrivateIps6 != nil {
						unit.PrivateIps6 = append(
							unit.PrivateIps6, deply.PrivateIps6...)
					}
					if deply.OraclePrivateIps != nil {
						unit.OraclePrivateIps = append(
							unit.OraclePrivateIps, deply.OraclePrivateIps...)
					}
					if deply.OraclePublicIps != nil {
						unit.OraclePublicIps = append(
							unit.OraclePublicIps, deply.OraclePublicIps...)
					}

					unit.Deployments = append(unit.Deployments, &Deployment{
						Id:               deply.Id,
						Service:          deply.Service,
						Unit:             deply.Unit,
						Spec:             deply.Spec,
						Kind:             deply.Kind,
						State:            deply.State,
						Node:             deply.Node,
						Instance:         deply.Instance,
						PublicIps:        deply.PublicIps,
						PublicIps6:       deply.PublicIps6,
						PrivateIps:       deply.PrivateIps,
						PrivateIps6:      deply.PrivateIps6,
						OraclePrivateIps: deply.OraclePrivateIps,
						OraclePublicIps:  deply.OraclePublicIps,
					})
				}
			}

			units = append(units, unit)
		}

		data := &Service{
			Id:    servc.Id,
			Name:  servc.Name,
			Units: units,
		}

		datas = append(datas, data)
	}

	return datas
}
