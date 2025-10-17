package pool

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Pool struct {
	Id               bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name             string        `bson:"name" json:"name"`
	Comment          string        `bson:"comment" json:"comment"`
	DeleteProtection bool          `bson:"delete_protection" json:"delete_protection"`
	Datacenter       bson.ObjectID `bson:"datacenter" json:"datacenter"`
	Zone             bson.ObjectID `bson:"zone" json:"zone"`
	Type             string        `bson:"type" json:"type"`
	VgName           string        `bson:"vg_name" json:"vg_name"`
}

type Completion struct {
	Id   bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name string        `bson:"name" json:"name"`
	Zone bson.ObjectID `bson:"zone" json:"zone"`
}

func (p *Pool) Json(nodeNames map[bson.ObjectID]string) {
}

func (p *Pool) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	p.Name = utils.FilterName(p.Name)

	if p.Datacenter.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "invalid_datacenter",
			Message: "Missing required datacenter",
		}
		return
	}

	if p.Zone.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "invalid_zone",
			Message: "Missing required zone",
		}
		return
	}

	return
}

func (p *Pool) Commit(db *database.Database) (err error) {
	coll := db.Pools()

	err = coll.Commit(p.Id, p)
	if err != nil {
		return
	}

	return
}

func (p *Pool) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Pools()

	err = coll.CommitFields(p.Id, p, fields)
	if err != nil {
		return
	}

	return
}

func (p *Pool) Insert(db *database.Database) (err error) {
	coll := db.Pools()

	_, err = coll.InsertOne(db, p)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
