package nodeport

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type NodePort struct {
	Id           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Datacenter   primitive.ObjectID `bson:"datacenter" json:"datacenter"`
	Organization primitive.ObjectID `bson:"organization" json:"organization"`
	Protocol     string             `bson:"protocol" json:"protocol"`
	Port         int                `bson:"port" json:"port"`
}

func (n *NodePort) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	switch n.Protocol {
	case Tcp, Udp:
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "invalid_protocol",
			Message: "Invalid node port protocol",
		}
		return
	}

	if n.Port != 0 {
		portRanges, e := GetPortRanges()
		if e != nil {
			err = e
			return
		}

		matched := false
		for _, ports := range portRanges {
			if ports.Contains(n.Port) {
				matched = true
				break
			}
		}

		if !matched {
			errData = &errortypes.ErrorData{
				Error:   "invalid_port",
				Message: "Invalid node port",
			}
			return
		}
	}

	return
}

func (n *NodePort) Sync(db *database.Database) (err error) {
	coll := db.Instances()

	count, err := coll.CountDocuments(db, &bson.M{
		"node_ports.node_port": n.Id,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if count == 0 {
		err = Remove(db, n.Id)
		if err != nil {
			return
		}
	}

	return
}

func (n *NodePort) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.NodePorts()

	err = coll.CommitFields(n.Id, n, fields)
	if err != nil {
		return
	}

	return
}

func (n *NodePort) Insert(db *database.Database) (err error) {
	coll := db.NodePorts()

	resp, err := coll.InsertOne(db, n)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	n.Id = resp.InsertedID.(primitive.ObjectID)
	return
}
