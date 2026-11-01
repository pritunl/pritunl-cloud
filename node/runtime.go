package node

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/metric"
	"github.com/pritunl/pritunl-cloud/settings"
)

type Runtime struct {
	Id         bson.ObjectID             `bson:"_id,omitempty" json:"-"`
	Timestamp  time.Time                 `bson:"timestamp" json:"timestamp"`
	Memory     float64                   `bson:"memory" json:"memory"`
	Swap       float64                   `bson:"swap" json:"swap"`
	HugePages  float64                   `bson:"hugepages" json:"hugepages"`
	Load1      float64                   `bson:"load1" json:"load1"`
	Load5      float64                   `bson:"load5" json:"load5"`
	Load15     float64                   `bson:"load15" json:"load15"`
	Disks      []*metric.DiskStatic      `bson:"disks" json:"disks"`
	Mounts     []*metric.MountStatic     `bson:"mounts" json:"mounts"`
	Interfaces []*metric.InterfaceStatic `bson:"interfaces" json:"interfaces"`
	Processes  []string                  `bson:"processes" json:"processes"`
	Modules    []string                  `bson:"modules" json:"modules"`
	Ports      []string                  `bson:"ports" json:"ports"`
}

func (r *Runtime) Online() bool {
	if r == nil || r.Timestamp.IsZero() {
		return false
	}

	return time.Since(r.Timestamp) <= time.Duration(
		settings.System.NodeTimestampTtl)*time.Second
}

type RuntimePipe struct {
	Node        `bson:",inline"`
	RuntimeDocs []*Runtime `bson:"runtime_docs"`
}

func (p *RuntimePipe) Merge() (nde *Node) {
	nde = &p.Node
	if len(p.RuntimeDocs) > 0 {
		nde.Runtime = p.RuntimeDocs[0]
	}
	return
}

func RuntimeLookup() *bson.M {
	return &bson.M{
		"$lookup": &bson.M{
			"from":         "nodes_runtime",
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "runtime_docs",
		},
	}
}

func GetRuntime(db *database.Database, nodeId bson.ObjectID) (
	rntm *Runtime, err error) {

	coll := db.NodesRuntime()
	rntm = &Runtime{}

	err = coll.FindOneId(nodeId, rntm)
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
			rntm = &Runtime{}
			return
		}
		return
	}

	return
}

func UpdateRuntime(db *database.Database, nodeId bson.ObjectID,
	fields bson.M) (err error) {

	coll := db.NodesRuntime()

	_, err = coll.UpdateOne(
		db,
		&bson.M{
			"_id": nodeId,
		},
		&bson.M{
			"$set": fields,
		},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveRuntime(db *database.Database, nodeId bson.ObjectID) (
	err error) {

	coll := db.NodesRuntime()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": nodeId,
	})
	if err != nil {
		err = database.ParseError(err)
		switch err.(type) {
		case *database.NotFoundError:
			err = nil
		default:
			return
		}
	}

	return
}
