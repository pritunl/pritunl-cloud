package types

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/unit"
)

type Unit struct {
	Id               bson.ObjectID `json:"id"`
	Pod              bson.ObjectID `json:"pod"`
	Organization     bson.ObjectID `json:"organization"`
	Name             string        `json:"name"`
	Kind             string        `json:"kind"`
	Count            int           `json:"count"`
	Primary          bson.ObjectID `json:"primary"`
	PrimaryTimestamp time.Time     `json:"primary_timestamp"`
}

func NewUnit(unt *unit.Unit) *Unit {
	if unt == nil {
		return &Unit{}
	}

	return &Unit{
		Id:               unt.Id,
		Pod:              unt.Pod,
		Organization:     unt.Organization,
		Name:             unt.Name,
		Kind:             unt.Kind,
		Count:            unt.Count,
		Primary:          unt.Primary,
		PrimaryTimestamp: unt.PrimaryTimestamp,
	}
}
