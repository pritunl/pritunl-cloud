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

	_, err = coll.InsertOne(db, n)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

type Mapping struct {
	NodePort     primitive.ObjectID `bson:"node_port" json:"node_port"`
	Protocol     string             `bson:"protocol" json:"protocol"`
	ExternalPort int                `bson:"external_port" json:"external_port"`
	InternalPort int                `bson:"internal_port" json:"internal_port"`
	Delete       bool               `bson:"-" json:"delete"`
}

func (m *Mapping) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	switch m.Protocol {
	case Tcp, Udp:
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "invalid_protocol",
			Message: "Invalid node port protocol",
		}
		return
	}

	portRanges, err := GetPortRanges()
	if err != nil {
		return
	}

	matched := false
	for _, ports := range portRanges {
		if ports.Contains(m.ExternalPort) {
			matched = true
			break
		}
	}

	if !matched {
		errData = &errortypes.ErrorData{
			Error:   "invalid_external_port",
			Message: "Invalid external node port",
		}
		return
	}

	if m.InternalPort <= 0 || m.InternalPort > 65535 {
		errData = &errortypes.ErrorData{
			Error:   "invalid_internal_port",
			Message: "Invalid internal node port",
		}
		return
	}

	return
}

func (m *Mapping) Diff(mapping *Mapping) bool {
	if m.Protocol != mapping.Protocol {
		return true
	}

	if m.ExternalPort != mapping.ExternalPort {
		return true
	}

	if m.InternalPort != mapping.InternalPort {
		return true
	}

	return false
}
