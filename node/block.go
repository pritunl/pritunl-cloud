package node

import (
	"gopkg.in/mgo.v2/bson"
)

type BlockAttachment struct {
	Interface string        `bson:"interface" json:"interface"`
	Block     bson.ObjectId `bson:"block" json:"block"`
}

