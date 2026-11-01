package instance

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/metric"
)

type Runtime struct {
	Id         bson.ObjectID             `bson:"_id,omitempty" json:"-"`
	Status     string                    `bson:"status" json:"status"`
	Timestamp  time.Time                 `bson:"timestamp" json:"timestamp"`
	Heartbeat  time.Time                 `bson:"heartbeat" json:"heartbeat"`
	Memory     float64                   `bson:"memory" json:"memory"`
	Swap       float64                   `bson:"swap" json:"swap"`
	HugePages  float64                   `bson:"hugepages" json:"hugepages"`
	Load1      float64                   `bson:"load1" json:"load1"`
	Load5      float64                   `bson:"load5" json:"load5"`
	Load15     float64                   `bson:"load15" json:"load15"`
	Processes  []string                  `bson:"processes" json:"processes"`
	Modules    []string                  `bson:"modules" json:"modules"`
	Ports      []string                  `bson:"ports" json:"ports"`
	Disks      []*metric.DiskStatic      `bson:"disks" json:"disks"`
	Mounts     []*metric.MountStatic     `bson:"mounts" json:"mounts"`
	Interfaces []*metric.InterfaceStatic `bson:"interfaces" json:"interfaces"`
}

type Pipe struct {
	Instance    `bson:",inline"`
	RuntimeDocs []*Runtime `bson:"runtime_docs"`
	InfoDocs    []*Info    `bson:"info_docs"`
}

func (p *Pipe) Merge() (inst *Instance) {
	inst = &p.Instance
	if len(p.RuntimeDocs) > 0 {
		inst.Runtime = p.RuntimeDocs[0]
	}
	if len(p.InfoDocs) > 0 {
		inst.Info = p.InfoDocs[0]
	}
	return
}

func RuntimeLookup() *bson.M {
	return &bson.M{
		"$lookup": &bson.M{
			"from":         "instances_runtime",
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "runtime_docs",
		},
	}
}

func GetRuntime(db *database.Database, instId bson.ObjectID) (
	rntm *Runtime, err error) {

	coll := db.InstancesRuntime()
	rntm = &Runtime{}

	err = coll.FindOneId(instId, rntm)
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

func UpdateRuntime(db *database.Database, instId bson.ObjectID,
	fields bson.M) (err error) {

	coll := db.InstancesRuntime()

	_, err = coll.UpdateOne(
		db,
		&bson.M{
			"_id": instId,
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

func RemoveRuntime(db *database.Database, instId bson.ObjectID) (
	err error) {

	coll := db.InstancesRuntime()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": instId,
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
