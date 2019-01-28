package vpc

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
)

type VpcIp struct {
	Id       primitive.ObjectID `bson:"_id,omitempty"`
	Vpc      primitive.ObjectID `bson:"vpc"`
	Ip       int64              `bson:"ip"`
	Type     string             `bson:"type"`
	Instance primitive.ObjectID `bson:"instance"`
}
