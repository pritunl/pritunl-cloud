package domain

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type Record struct {
	Id           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Organization primitive.ObjectID `bson:"organization" json:"organization"`
	Domain       primitive.ObjectID `bson:"domain" json:"domain"`
	Node         primitive.ObjectID `bson:"node" json:"node"`
	Instance     primitive.ObjectID `bson:"instance" json:"instance"`
	Timestamp    time.Time          `bson:"timestamp" json:"timestamp"`
	Name         string             `bson:"name" json:"name"`
	Address      string             `bson:"address" json:"address"`
	Address6     string             `bson:"address6" json:"address6"`
}

func (r *Record) Remove(db *database.Database) (err error) {
	domn, err := GetOrg(db, r.Organization, r.Domain)
	if err != nil {
		return
	}

	if domn.Type == Route53 {
		err = AwsUpsertDomain(domn, r.Name, "", "")
		if err != nil {
			return
		}
	} else {
		err = &errortypes.UnknownError{
			errors.New("domain: Unknown domain type"),
		}
		return
	}

	return
}

func (r *Record) Upsert(db *database.Database, addr, addr6 string) (
	err error) {

	domn, err := GetOrg(db, r.Organization, r.Domain)
	if err != nil {
		return
	}

	r.Timestamp = time.Now()

	if r.Id.IsZero() {
		err = r.Insert(db)
		if err != nil {
			return
		}
	}

	if domn.Type == Route53 {
		err = AwsUpsertDomain(domn, r.Name, addr, addr6)
		if err != nil {
			return
		}
	} else {
		err = &errortypes.UnknownError{
			errors.New("domain: Unknown domain type"),
		}
		return
	}

	r.Address = addr
	r.Address6 = addr6

	err = r.CommitFields(
		db, set.NewSet("timestamp", "address", "address6"))
	if err != nil {
		return
	}

	return
}

func (r *Record) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if r.Node.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "node_required",
			Message: "Missing required node",
		}
		return
	}

	if r.Instance.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "instance_required",
			Message: "Missing required instance",
		}
		return
	}

	return
}

func (r *Record) Commit(db *database.Database) (err error) {
	coll := db.DomainsRecord()

	err = coll.Commit(r.Id, r)
	if err != nil {
		return
	}

	return
}

func (r *Record) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.DomainsRecord()

	err = coll.CommitFields(r.Id, r, fields)
	if err != nil {
		return
	}

	return
}

func (r *Record) Insert(db *database.Database) (err error) {
	coll := db.DomainsRecord()

	if !r.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("domain: Record already exists"),
		}
		return
	}

	r.Id = primitive.NewObjectID()

	_, err = coll.InsertOne(db, r)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
