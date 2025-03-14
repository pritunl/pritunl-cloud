package domain

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Record struct {
	Id              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Domain          primitive.ObjectID `bson:"domain" json:"domain"`
	Resource        primitive.ObjectID `bson:"resource,omitempty" json:"resource"`
	Deployment      primitive.ObjectID `bson:"deployment,omitempty" json:"deployment"`
	Timestamp       time.Time          `bson:"timestamp" json:"timestamp"`
	DeleteTimestamp time.Time          `bson:"delete_timestamp,omitempty" json:"delete_timestamp"`
	SubDomain       string             `bson:"sub_domain" json:"sub_domain"`
	Type            string             `bson:"type" json:"type"`
	Value           string             `bson:"value" json:"value"`
	Operation       string             `bson:"-" json:"operation"`
}

func (r *Record) IsDeleted() bool {
	return !r.DeleteTimestamp.IsZero()
}

func (r *Record) Copy() *Record {
	rec := *r
	return &rec
}

func (r *Record) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	r.SubDomain = utils.FilterDomain(r.SubDomain)

	if r.Domain.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "domain_required",
			Message: "Missing required domain",
		}
		return
	}

	switch r.Type {
	case A:
		break
	case AAAA:
		break
	default:
		err = &errortypes.UnknownError{
			errors.New("domain: Unknown record type"),
		}
		return
	}

	if r.Value == "" {
		errData = &errortypes.ErrorData{
			Error:   "value_required",
			Message: "Missing required value",
		}
		return
	}

	return
}

func (r *Record) Commit(db *database.Database) (err error) {
	coll := db.DomainsRecords()

	err = coll.Commit(r.Id, r)
	if err != nil {
		return
	}

	return
}

func (r *Record) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.DomainsRecords()

	err = coll.CommitFields(r.Id, r, fields)
	if err != nil {
		return
	}

	return
}

func (r *Record) Remove(db *database.Database) (err error) {
	coll := db.DomainsRecords()

	if r.DeleteTimestamp.IsZero() {
		r.DeleteTimestamp = time.Now()

		err = coll.CommitFields(r.Id, r, set.NewSet("delete_timestamp"))
		if err != nil {
			return
		}

		return
	}

	deleteTtl := time.Duration(settings.System.DomainDeleteTtl) * time.Second
	if time.Since(r.DeleteTimestamp) < deleteTtl {
		return
	}

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": r.Id,
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

func (r *Record) Insert(db *database.Database) (err error) {
	coll := db.DomainsRecords()

	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	newRec := &Record{}
	err = coll.FindOneAndUpdate(db, &bson.M{
		"domain":     r.Domain,
		"sub_domain": r.SubDomain,
		"type":       r.Type,
		"value":      r.Value,
	}, &bson.M{
		"$set": r,
	}, opts).Decode(&newRec)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	r.Id = newRec.Id

	return
}
