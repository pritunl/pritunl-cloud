package node

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
)

type BlockAttachment struct {
	Interface string             `bson:"interface" json:"interface"`
	Block     primitive.ObjectID `bson:"block" json:"block"`
}
