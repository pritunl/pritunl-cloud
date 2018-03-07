package zone

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"gopkg.in/mgo.v2/bson"
)

type Zone struct {
	Id            bson.ObjectId   `bson:"_id,omitempty" json:"id"`
	Organizations []bson.ObjectId `bson:"organizations" json:"organizations"`
	Datacenter    bson.ObjectId   `bson:"datacenter,omitempty" json:"datacenter"`
	Name          string          `bson:"name" json:"name"`
}

func (z *Zone) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if z.Organizations == nil {
		z.Organizations = []bson.ObjectId{}
	}

	if z.Datacenter == "" {
		errData = &errortypes.ErrorData{
			Error:   "datacenter_required",
			Message: "Missing required datacenter",
		}
		return
	}

	return
}

func (z *Zone) Commit(db *database.Database) (err error) {
	coll := db.Zones()

	err = coll.Commit(z.Id, z)
	if err != nil {
		return
	}

	return
}

func (z *Zone) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Zones()

	err = coll.CommitFields(z.Id, z, fields)
	if err != nil {
		return
	}

	return
}

func (z *Zone) Insert(db *database.Database) (err error) {
	coll := db.Zones()

	if z.Id != "" {
		err = &errortypes.DatabaseError{
			errors.New("zone: Zone already exists"),
		}
		return
	}

	err = coll.Insert(z)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
