package node

import "github.com/pritunl/mongo-go-driver/v2/bson"

type BlockAttachment struct {
	Interface string        `bson:"interface" json:"interface"`
	Block     bson.ObjectID `bson:"block" json:"block"`
}
