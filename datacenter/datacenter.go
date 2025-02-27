package datacenter

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Datacenter struct {
	Id                  primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name                string               `bson:"name" json:"name"`
	Comment             string               `bson:"comment" json:"comment"`
	MatchOrganizations  bool                 `bson:"match_organizations" json:"match_organizations"`
	Organizations       []primitive.ObjectID `bson:"organizations" json:"organizations"`
	NetworkMode         string               `bson:"network_mode" json:"network_mode"`
	PublicStorages      []primitive.ObjectID `bson:"public_storages" json:"public_storages"`
	PrivateStorage      primitive.ObjectID   `bson:"private_storage,omitempty" json:"private_storage"`
	PrivateStorageClass string               `bson:"private_storage_class" json:"private_storage_class"`
	BackupStorage       primitive.ObjectID   `bson:"backup_storage,omitempty" json:"backup_storage"`
	BackupStorageClass  string               `bson:"backup_storage_class" json:"backup_storage_class"`
}

func (d *Datacenter) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	d.Name = utils.FilterName(d.Name)

	if d.Organizations == nil || !d.MatchOrganizations {
		d.Organizations = []primitive.ObjectID{}
	}

	if d.PublicStorages == nil {
		d.PublicStorages = []primitive.ObjectID{}
	}

	return
}

func (d *Datacenter) Commit(db *database.Database) (err error) {
	coll := db.Datacenters()

	err = coll.Commit(d.Id, d)
	if err != nil {
		return
	}

	return
}

func (d *Datacenter) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Datacenters()

	err = coll.CommitFields(d.Id, d, fields)
	if err != nil {
		return
	}

	return
}

func (d *Datacenter) Insert(db *database.Database) (err error) {
	coll := db.Datacenters()

	if !d.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("datacenter: Datacenter already exists"),
		}
		return
	}

	resp, err := coll.InsertOne(db, d)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	d.Id = resp.InsertedID.(primitive.ObjectID)

	return
}
