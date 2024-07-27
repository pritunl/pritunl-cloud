package domain

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/dns"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type Domain struct {
	Id           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Comment      string             `bson:"comment" json:"comment"`
	Organization primitive.ObjectID `bson:"organization,omitempty" json:"organization"`
	Type         string             `bson:"type" json:"type"`
	Secret       primitive.ObjectID `bson:"secret" json:"secret"`
	RootDomain   string             `bson:"root_domain" json:"root_domain"`
	Records      []*Record          `bson:"-" json:"records"`
}

func (d *Domain) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if d.Organization.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "organization_required",
			Message: "Missing required organization",
		}
		return
	}

	switch d.Type {
	case AWS, "":
		d.Type = AWS
		break
	case Cloudflare:
		break
	case OracleCloud:
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "type_invalid",
			Message: "Type invalid",
		}
		return
	}

	if d.Secret.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "secret_invalid",
			Message: "Secret invalid",
		}
		return
	}

	return
}

func (d *Domain) Commit(db *database.Database) (err error) {
	coll := db.Domains()

	err = coll.Commit(d.Id, d)
	if err != nil {
		return
	}

	return
}

func (d *Domain) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Domains()

	err = coll.CommitFields(d.Id, d, fields)
	if err != nil {
		return
	}

	return
}

func (d *Domain) Insert(db *database.Database) (err error) {
	coll := db.Domains()

	if !d.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("domain: Domain already exists"),
		}
		return
	}

	_, err = coll.InsertOne(db, d)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (d *Domain) GetDnsService(db *database.Database) (
	svc dns.Service, err error) {

	switch d.Type {
	case AWS:
		svc = &dns.Aws{}
		break
	case Cloudflare:
		svc = &dns.Cloudflare{}
		break
	case OracleCloud:
		svc = &dns.Oracle{}
		break
	default:
		err = &errortypes.UnknownError{
			errors.Newf("domain: Unknown domain type"),
		}
		return
	}

	return
}

func (d *Domain) LoadRecords(db *database.Database) (err error) {
	coll := db.DomainsRecords()
	recs := []*Record{}

	cursor, err := coll.Find(db, &bson.M{
		"domain": d.Id,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		rec := &Record{}
		err = cursor.Decode(rec)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		recs = append(recs, rec)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	d.Records = recs

	return
}
