package service

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
)

type Unit struct {
	Id       primitive.ObjectID `bson:"_id" json:"id"`
	Name     string             `bson:"name" json:"name"`
	Kind     string             `bson:"kind" json:"kind"`
	Data     interface{}        `bson:"data" json:"data"`
	Deployed interface{}        `bson:"deployed" json:"deployed"`
	Spec     string             `bson:"spec" json:"spec"`
}
