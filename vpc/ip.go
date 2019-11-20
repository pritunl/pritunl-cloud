package vpc

import (
	"net"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/utils"
)

type VpcIp struct {
	Id       primitive.ObjectID `bson:"_id,omitempty"`
	Vpc      primitive.ObjectID `bson:"vpc"`
	Subnet   primitive.ObjectID `bson:"subnet"`
	Ip       int64              `bson:"ip"`
	Instance primitive.ObjectID `bson:"instance"`
}
