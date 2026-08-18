package types

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/deployment"
)

type Deployment struct {
	Id           bson.ObjectID `json:"id"`
	Pod          bson.ObjectID `json:"pod"`
	Unit         bson.ObjectID `json:"unit"`
	Organization bson.ObjectID `json:"organization"`
	Timestamp    time.Time     `json:"timestamp"`
	Tags         []string      `json:"tags"`
	Spec         bson.ObjectID `json:"spec"`
	Kind         string        `json:"kind"`
	State        string        `json:"state"`
	Action       string        `json:"action"`
	Status       string        `json:"status"`
	Failover     string        `json:"failover"`
}

func NewDeployment(deply *deployment.Deployment) *Deployment {
	if deply == nil {
		return &Deployment{}
	}

	return &Deployment{
		Id:           deply.Id,
		Pod:          deply.Pod,
		Unit:         deply.Unit,
		Organization: deply.Organization,
		Timestamp:    deply.Timestamp,
		Tags:         deply.Tags,
		Spec:         deply.Spec,
		Kind:         deply.Kind,
		State:        deply.State,
		Action:       deply.Action,
		Status:       deply.Status,
		Failover:     deply.Failover,
	}
}
