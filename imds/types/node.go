package types

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/node"
)

type Node struct {
	Id         bson.ObjectID `json:"id"`
	Name       string        `json:"name"`
	PublicIps  []string      `json:"public_ips"`
	PublicIps6 []string      `json:"public_ips6"`
}

func NewNode(nde *node.Node) *Node {
	if nde == nil {
		return &Node{}
	}

	return &Node{
		Id:         nde.Id,
		Name:       nde.Name,
		PublicIps:  nde.PublicIps,
		PublicIps6: nde.PublicIps6,
	}
}
