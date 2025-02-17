package domain

import (
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/dns"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/utils"
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
	OrigRecords  []*Record          `bson:"-" json:"-"`
}

func (d *Domain) Copy() *Domain {
	domn := *d

	recs := make([]*Record, len(domn.Records))
	for i, rec := range domn.Records {
		recs[i] = rec.Copy()
	}
	domn.Records = recs

	origRecs := make([]*Record, len(domn.OrigRecords))
	for i, rec := range domn.OrigRecords {
		origRecs[i] = rec.Copy()
	}
	domn.OrigRecords = origRecs

	return &domn
}

func (d *Domain) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	d.Name = utils.FilterName(d.Name)

	d.RootDomain = strings.ToLower(d.RootDomain)

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

	newRecords := []*Record{}
	for _, record := range d.Records {
		record.Domain = d.Id

		if record.Operation == DELETE && record.Id.IsZero() {
			continue
		}

		errData, err = record.Validate(db)
		if err != nil {
			return
		}

		if errData != nil {
			return
		}

		newRecords = append(newRecords, record)
	}
	d.Records = newRecords

	return
}

func (d *Domain) PreCommit() {
	d.OrigRecords = d.Records
}

func (d *Domain) CommitRecords(db *database.Database) (err error) {
	secr, err := secret.GetOrg(db, d.Organization, d.Secret)
	if err != nil {
		return
	}

	newRecords := []*Record{}
	for _, record := range d.Records {
		if record.Operation == DELETE {
			for _, origRecord := range d.OrigRecords {
				if record.Id == origRecord.Id {
					record = origRecord
					record.Operation = DELETE
					break
				}
			}
		}

		newRecords = append(newRecords, record)
	}
	d.Records = newRecords

	batches := map[string][]*Record{}

	for _, record := range d.Records {
		batchKey := record.SubDomain + ":" + record.Type
		batches[batchKey] = append(batches[batchKey], record)
	}

	for _, records := range batches {
		err = d.UpdateRecords(db, secr, records)
		if err != nil {
			return
		}
	}

	return
}

func (d *Domain) UpdateRecords(db *database.Database, secr *secret.Secret,
	records []*Record) (err error) {

	ops := []*dns.Operation{}
	subDomain := ""
	dnsType := ""

	for _, rec := range records {
		if subDomain == "" {
			subDomain = rec.SubDomain
		} else if rec.SubDomain != subDomain {
			err = &errortypes.ParseError{
				errors.Newf("domain: Update subdomain inconsistent"),
			}
			return
		}

		if dnsType == "" {
			dnsType = rec.Type
		} else if rec.Type != dnsType {
			err = &errortypes.ParseError{
				errors.Newf("domain: Update type inconsistent"),
			}
			return
		}

		switch rec.Operation {
		case INSERT, UPDATE:
			ops = append(ops, &dns.Operation{
				Operation: dns.UPSERT,
				Value:     rec.Value,
			})
			break
		case DELETE:
			ops = append(ops, &dns.Operation{
				Operation: dns.DELETE,
				Value:     rec.Value,
			})
			break
		default:
			ops = append(ops, &dns.Operation{
				Operation: dns.RETAIN,
				Value:     rec.Value,
			})
		}
	}

	domain := subDomain + "." + d.RootDomain

	svc, err := d.GetDnsService(db)
	if err != nil {
		return
	}

	err = svc.Connect(db, secr)
	if err != nil {
		return
	}

	err = svc.DnsCommit(db, domain, dnsType, ops)
	if err != nil {
		return
	}

	for _, rec := range records {
		rec.Timestamp = time.Now()

		switch rec.Operation {
		case INSERT:
			err = rec.Insert(db)
			if err != nil {
				return
			}
			break
		case DELETE:
			err = rec.Remove(db)
			if err != nil {
				return
			}
			break
		default:
			err = rec.Commit(db)
			if err != nil {
				return
			}
		}
	}

	return
}

func (d *Domain) MergeRecords(deplyId primitive.ObjectID,
	newRecs []*Record) (newDomn *Domain) {
	recMap := map[string]map[string]*Record{}

	for _, rec := range d.Records {
		if rec.Deployment != deplyId {
			continue
		}

		if recMap[rec.SubDomain] == nil {
			recMap[rec.SubDomain] = map[string]*Record{}
		}
		recMap[rec.SubDomain][rec.Value] = rec
	}

	for _, newRec := range newRecs {
		subRecs := recMap[newRec.SubDomain]

		rec := subRecs[newRec.Value]
		if rec == nil {
			if newDomn == nil {
				newDomn = d.Copy()
				d = newDomn
				d.PreCommit()
			}
			newRec.Operation = INSERT
			d.Records = append(d.Records, newRec)
		} else {
			delete(subRecs, newRec.Value)
		}
	}

	for _, subRecs := range recMap {
		for _, rec := range subRecs {
			if newDomn == nil {
				newDomn = d.Copy()
				d = newDomn
				d.PreCommit()
			}
			rec.Operation = DELETE
		}
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
	}, &options.FindOptions{
		Sort: &bson.D{
			{"sub_domain", 1},
		},
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
