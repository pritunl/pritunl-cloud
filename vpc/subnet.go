package vpc

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
)

type Subnet struct {
	Id      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name    string             `bson:"name" json:"name"`
	Network string             `bson:"network" json:"network"`
}
