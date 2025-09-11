package types

import (
	"fmt"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/vpc"
)

type Vpc struct {
	Id       bson.ObjectID `json:"id"`
	Name     string        `json:"name"`
	VpcId    int           `json:"vpc_id"`
	Network  string        `json:"network"`
	Network6 string        `json:"network6"`
	Subnets  []*Subnet     `json:"subnets"`
	Routes   []*Route      `json:"routes"`
}

type Subnet struct {
	Id      bson.ObjectID `json:"id"`
	Name    string        `json:"name"`
	Network string        `json:"network"`
}

func (s *Subnet) String() string {
	return s.Network
}

type Route struct {
	Destination string `json:"destination"`
	Target      string `json:"target"`
}

func (r *Route) String() string {
	return fmt.Sprintf("%s via %s", r.Destination, r.Target)
}

func NewSubnet(subnet *vpc.Subnet) *Subnet {
	if subnet == nil {
		return &Subnet{}
	}
	return &Subnet{
		Id:      subnet.Id,
		Name:    subnet.Name,
		Network: subnet.Network,
	}
}

func NewRoute(subnet *vpc.Route) *Route {
	if subnet == nil {
		return &Route{}
	}
	return &Route{
		Destination: subnet.Destination,
		Target:      subnet.Target,
	}
}

func NewVpc(vpc *vpc.Vpc) *Vpc {
	if vpc == nil {
		return &Vpc{}
	}

	vpc.Json()

	subnets := []*Subnet{}
	if vpc.Subnets != nil {
		for _, subnet := range vpc.Subnets {
			subnets = append(subnets, NewSubnet(subnet))
		}
	}

	routes := []*Route{}
	if vpc.Routes != nil {
		for _, route := range vpc.Routes {
			routes = append(routes, NewRoute(route))
		}
	}

	return &Vpc{
		Id:       vpc.Id,
		Name:     vpc.Name,
		VpcId:    vpc.VpcId,
		Network:  vpc.Network,
		Network6: vpc.Network6,
		Subnets:  subnets,
		Routes:   routes,
	}
}
