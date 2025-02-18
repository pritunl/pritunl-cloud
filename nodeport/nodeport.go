package nodeport

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
)

type NodePort struct {
	Id         primitive.ObjectID `bson:"id" json:"id"`
	Datacenter primitive.ObjectID `bson:"datacenter" json:"datacenter"`
	Protocol   string             `bson:"protocol" json:"protocol"`
	Port       int                `bson:"port" json:"port"`
	Resource   primitive.ObjectID `bson:"resource" json:"resource"`
}

func (n *NodePort) Insert(db *database.Database) (err error) {
	coll := db.NodePorts()

	_, err = coll.InsertOne(db, n)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
