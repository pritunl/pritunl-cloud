package vpc

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"net"
)

type Vpc struct {
	Id           bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Name         string        `bson:"name" json:"name"`
	VpcId        int           `bson:"vpc_id" json:"vpc_id"`
	Network      string        `bson:"network" json:"network"`
	Organization bson.ObjectId `bson:"organization" json:"organization"`
	Datacenter   bson.ObjectId `bson:"datacenter" json:"datacenter"`
}

func (v *Vpc) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if v.VpcId == 0 {
		errData = &errortypes.ErrorData{
			Error:   "vpc_id_invalid",
			Message: "Vpc ID invalid",
		}
		return
	}

	if v.Organization == "" {
		errData = &errortypes.ErrorData{
			Error:   "organization_required",
			Message: "Missing required organization",
		}
		return
	}

	if v.Datacenter == "" {
		errData = &errortypes.ErrorData{
			Error:   "datacenter_required",
			Message: "Missing required datacenter",
		}
		return
	}

	if v.Network != "" {
		network, e := v.GetNetwork()
		if e != nil {
			errData = &errortypes.ErrorData{
				Error:   "network_invalid",
				Message: "Network address invalid",
			}
			return
		}
		v.Network = network.String()
	}

	return
}

func (v *Vpc) GetNetwork() (network *net.IPNet, err error) {
	_, network, err = net.ParseCIDR(v.Network)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "vpc: Failed to parse network"),
		}
		return
	}
	return
}

func (v *Vpc) GenerateVpcId() {
	v.VpcId = rand.Intn(16777100) + 110
}

func (v *Vpc) Commit(db *database.Database) (err error) {
	coll := db.Vpcs()

	err = coll.Commit(v.Id, v)
	if err != nil {
		return
	}

	return
}

func (v *Vpc) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Vpcs()

	err = coll.CommitFields(v.Id, v, fields)
	if err != nil {
		return
	}

	return
}

func (v *Vpc) Insert(db *database.Database) (err error) {
	coll := db.Vpcs()

	if v.Id != "" {
		err = &errortypes.DatabaseError{
			errors.New("vpc: Vpc already exists"),
		}
		return
	}

	err = coll.Insert(v)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
