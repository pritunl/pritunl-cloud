package balancer

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
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
	Type         string               `bson:"type" json:"type"`
	Organization primitive.ObjectID   `bson:"organization,omitempty" json:"organization"`
	Zone         primitive.ObjectID   `bson:"zone,omitempty"`
	Certificates []primitive.ObjectID `bson:"certificates" json:"certificates"`
	WebSockets   bool                 `bson:"websockets" json:"websockets"`
	Domains      []*Domain            `bson:"domains" json:"domains"`
	Backends     []*Backend           `bson:"backends" json:"backends"`
}

func (b *Balancer) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if b.Type == "" {
		b.Type = Http
	}

	if b.Domains == nil {
		b.Domains = []*Domain{}
	}

	for _, backend := range b.Backends {
		if backend.Protocol != "http" && backend.Protocol != "https" {
			errData = &errortypes.ErrorData{
				Error:   "balancer_protocol_invalid",
				Message: "Invalid balancer backend protocol",
			}
			return
		}

		if backend.Hostname == "" {
			errData = &errortypes.ErrorData{
				Error:   "balancer_hostname_invalid",
				Message: "Invalid balancer backend hostname",
			}
			return
		}

		if backend.Port == 0 {
			errData = &errortypes.ErrorData{
				Error:   "balancer_port_invalid",
				Message: "Invalid balancer backend port",
			}
			return
		}
	}

	return
}
