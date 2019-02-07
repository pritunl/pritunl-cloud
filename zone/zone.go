package zone

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type Zone struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Datacenter  primitive.ObjectID `bson:"datacenter,omitempty" json:"datacenter"`
	Name        string             `bson:"name" json:"name"`
	NetworkMode string             `bson:"network_mode" json:"network_mode"`
}

func (z *Zone) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if z.Datacenter.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "datacenter_required",
			Message: "Missing required datacenter",
		}
		return
	}

	switch z.NetworkMode {
	case Default:
		break
	case VxLan:
		break
	case "":
		z.NetworkMode = Default
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "invalid_network_mode",
			Message: "Network mode invalid",
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

	if !z.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("zone: Zone already exists"),
		}
		return
	}

	_, err = coll.InsertOne(db, z)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
