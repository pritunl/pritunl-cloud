package shape

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type Shape struct {
	Id               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name             string             `bson:"name" json:"name"`
	Comment          string             `bson:"comment" json:"comment"`
	DeleteProtection bool               `bson:"delete_protection" json:"delete_protection"`
	Zone             primitive.ObjectID `bson:"zone" json:"zone"`
	Roles            []string           `bson:"roles" json:"roles"`
	Flexible         bool               `bson:"flexible" json:"flexible"`
	DiskType         string             `bson:"disk_type" json:"disk_type"`
	DiskPool         primitive.ObjectID `bson:"disk_pool" json:"disk_pool"`
	Memory           int                `bson:"memory" json:"memory"`
	Processors       int                `bson:"processors" json:"processors"`
}

func (p *Shape) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	return
}

func (p *Shape) Commit(db *database.Database) (err error) {
	coll := db.Shapes()

	err = coll.Commit(p.Id, p)
	if err != nil {
		return
	}

	return
}

func (p *Shape) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Shapes()

	err = coll.CommitFields(p.Id, p, fields)
	if err != nil {
		return
	}

	return
}

func (p *Shape) Insert(db *database.Database) (err error) {
	coll := db.Shapes()

	_, err = coll.InsertOne(db, p)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
