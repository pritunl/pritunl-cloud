package deployment

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type Deployment struct {
	Id         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Service    primitive.ObjectID `bson:"service" json:"service"`
	Unit       primitive.ObjectID `bson:"unit" json:"unit"`
	Deployment primitive.ObjectID `bson:"deployment" json:"deployment"`
	Kind       string             `bson:"kind" json:"kind"`
	State      string             `bson:"state" json:"state"`
	Node       primitive.ObjectID `bson:"node" json:"node"`
	Instance   primitive.ObjectID `bson:"instance" json:"instance"`
}

func (d *Deployment) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	return
}

func (d *Deployment) Commit(db *database.Database) (err error) {
	coll := db.Deployments()

	err = coll.Commit(d.Id, d)
	if err != nil {
		return
	}

	return
}

func (d *Deployment) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Deployments()

	err = coll.CommitFields(d.Id, d, fields)
	if err != nil {
		return
	}

	return
}

func (d *Deployment) Insert(db *database.Database) (err error) {
	coll := db.Deployments()

	_, err = coll.InsertOne(db, d)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
