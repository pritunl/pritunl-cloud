package balancer

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
)

type Domain struct {
	Domain string `bson:"domain" json:"domain"`
	Host   string `bson:"host" json:"host"`
}

type Backend struct {
	Protocol string `bson:"protocol" json:"protocol"`
	Hostname string `bson:"hostname" json:"hostname"`
	Port     int    `bson:"port" json:"port"`
}

type Balancer struct {
	Id           primitive.ObjectID   `bson:"_id" json:"id"`
	Name         string               `bson:"name" json:"name"`
	Organization primitive.ObjectID   `bson:"organization,omitempty" json:"organization"`
	Zone         primitive.ObjectID   `bson:"zone,omitempty"`
	Certificates []primitive.ObjectID `bson:"certificates" json:"certificates"`
	WebSockets   bool                 `bson:"websockets" json:"websockets"`
	Domains      []*Domain            `bson:"domains" json:"domains"`
	Backends     []*Backend           `bson:"backends" json:"backends"`
}
