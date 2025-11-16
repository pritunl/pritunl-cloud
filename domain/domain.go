package domain

import (
	"context"
	"sync"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/dns"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

type Domain struct {
	Id            bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string        `bson:"name" json:"name"`
	Comment       string        `bson:"comment" json:"comment"`
	Organization  bson.ObjectID `bson:"organization" json:"organization"`
	Type          string        `bson:"type" json:"type"`
	Secret        bson.ObjectID `bson:"secret" json:"secret"`
	RootDomain    string        `bson:"root_domain" json:"root_domain"`
	LockId        bson.ObjectID `bson:"lock_id" json:"lock_id"`
	LockTimestamp time.Time     `bson:"lock_timestamp" json:"lock_timestamp"`
	LastUpdate    time.Time     `bson:"last_update" json:"last_update"`
	Records       []*Record     `bson:"-" json:"records"`
	OrigRecords   []*Record     `bson:"-" json:"-"`
}

type Completion struct {
	Id           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string        `bson:"name" json:"name"`
	Organization bson.ObjectID `bson:"organization" json:"organization"`
}

func (d *Domain) Locked() bool {
	return !d.LockId.IsZero() && time.Since(d.LockTimestamp) < time.Duration(
		settings.System.DomainLockTtl)*time.Second
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

func (d *Domain) Json() {
	newRecords := make([]*Record, 0, len(d.Records))

	for _, rec := range d.Records {
		if !rec.IsDeleted() {
			newRecords = append(newRecords, rec)
		}
	}

	d.Records = newRecords
}

func (d *Domain) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	d.Name = utils.FilterName(d.Name)

	d.RootDomain = utils.FilterDomain(d.RootDomain)

	if d.Organization.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "organization_required",
			Message: "Missing required organization",
		}
		return
	}

	switch d.Type {
	case Local, "":
		d.Type = Local
		d.Secret = bson.NilObjectID
		break
	case AWS:
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

	if d.Type != Local && d.Secret.IsZero() {
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
	err = d.commitRecords(db, true)
	if err != nil {
		return
	}

	return
}

func (d *Domain) CommitRecordsSilent(db *database.Database) (err error) {
	err = d.commitRecords(db, false)
	if err != nil {
		return
	}

	return
}

func (d *Domain) commitRecords(db *database.Database,
	setTtl bool) (err error) {

	acquired := false
	var lockId bson.ObjectID
	for i := 0; i < 100; i++ {
		lockId, acquired, err = Lock(db, d.Id)
		if err != nil {
			return
		}

		if acquired {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !acquired {
		err = &errortypes.RequestError{
			errors.New("domain: Failed to acquire domain lock"),
		}
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		cancel()

		time.Sleep(100 * time.Millisecond)

		e := Unlock(db, d.Id, lockId)
		if e != nil {
			logrus.WithFields(logrus.Fields{
				"domain": d.Id.Hex(),
				"error":  e,
			}).Error("domain: Failed to unlock domain")
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				e := Relock(db, d.Id, lockId)
				if e != nil {
					logrus.WithFields(logrus.Fields{
						"domain": d.Id.Hex(),
						"error":  e,
					}).Error("domain: Failed to relock domain")
				}
			}
		}
	}()

	var secr *secret.Secret
	if d.Type != Local {
		secr, err = secret.GetOrg(db, d.Organization, d.Secret)
		if err != nil {
			return
		}
	}

	newRecords := []*Record{}
	for _, record := range d.Records {
		if record.Operation == DELETE || record.IsDeleted() {
			record.Operation = DELETE
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

	batches := map[string]map[string]*Record{}

	for _, record := range d.Records {
		batchKey := record.SubDomain + ":" + record.Type
		if batches[batchKey] == nil {
			batches[batchKey] = map[string]*Record{}
		}
		curRecord := batches[batchKey][record.Value]
		if curRecord == nil || record.Priority() > curRecord.Priority() {
			batches[batchKey][record.Value] = record
		}
	}

	if setTtl {
		d.LastUpdate = time.Now()
		err = d.CommitFields(db, set.NewSet("last_update"))
		if err != nil {
			return
		}
	}

	if d.Type == OracleCloud {
		err = d.asyncBatches(db, secr, batches)
		if err != nil {
			return
		}
	} else {
		err = d.syncBatches(db, secr, batches)
		if err != nil {
			return
		}
	}

	return
}

func (d *Domain) syncBatches(db *database.Database, secr *secret.Secret,
	batches map[string]map[string]*Record) (err error) {

	for _, recordMap := range batches {
		records := make([]*Record, 0, len(recordMap))
		for _, record := range recordMap {
			records = append(records, record)
		}

		err = d.UpdateRecords(db, secr, records)
		if err != nil {
			return
		}
	}

	return
}

func (d *Domain) asyncBatches(db *database.Database, secr *secret.Secret,
	batches map[string]map[string]*Record) (err error) {

	waiters := &sync.WaitGroup{}
	waiters.Add(len(batches))

	semaphore := make(
		chan struct{},
		settings.Acme.DnsMaxConcurrent,
	)
	errs := make(chan error, len(batches))

	for _, recordMap := range batches {
		records := make([]*Record, 0, len(recordMap))
		for _, record := range recordMap {
			records = append(records, record)
		}

		go func() {
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
				waiters.Done()
			}()

			e := d.UpdateRecords(db, secr, records)
			if e != nil {
				errs <- e
			}
		}()
	}

	waiters.Wait()
	close(errs)

	select {
	case err = <-errs:
		return err
	default:
		return nil
	}
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

	if d.Type != Local {
		svc, e := d.GetDnsService(db)
		if e != nil {
			err = e
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

func (d *Domain) MergeRecords(deployId bson.ObjectID,
	newRecs []*Record) (newDomn *Domain) {

	recMap := map[string]map[string]map[string]*Record{}

	for _, rec := range d.Records {
		if rec.Deployment != deployId || rec.IsDeleted() {
			continue
		}

		if recMap[rec.SubDomain] == nil {
			recMap[rec.SubDomain] = map[string]map[string]*Record{}
		}
		if recMap[rec.SubDomain][rec.Type] == nil {
			recMap[rec.SubDomain][rec.Type] = map[string]*Record{}
		}
		recMap[rec.SubDomain][rec.Type][rec.Value] = rec
	}

	for _, newRec := range newRecs {
		if recMap[newRec.SubDomain] == nil {
			recMap[newRec.SubDomain] = map[string]map[string]*Record{}
		}
		if recMap[newRec.SubDomain][newRec.Type] == nil {
			recMap[newRec.SubDomain][newRec.Type] = map[string]*Record{}
		}

		rec := recMap[newRec.SubDomain][newRec.Type][newRec.Value]
		if rec == nil {
			if newDomn == nil {
				newDomn = d.Copy()
				newDomn.PreCommit()
			}
			newRec.Operation = INSERT
			newDomn.Records = append(newDomn.Records, newRec)
		} else {
			delete(recMap[newRec.SubDomain][newRec.Type], newRec.Value)
		}
	}

	for subDomain, typeMap := range recMap {
		for typeName, valueMap := range typeMap {
			for value := range valueMap {
				if newDomn == nil {
					newDomn = d.Copy()
					newDomn.PreCommit()
				}

				for i, domainRec := range newDomn.Records {
					if domainRec.SubDomain == subDomain &&
						domainRec.Type == typeName &&
						domainRec.Value == value {

						newDomn.Records[i].Operation = DELETE
						break
					}
				}
			}
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

func (d *Domain) preloadRecords(recs []*Record) {
	if recs == nil {
		d.Records = []*Record{}
	} else {
		d.Records = recs
	}
}

func (d *Domain) LoadRecords(db *database.Database,
	skipDeleted bool) (err error) {

	coll := db.DomainsRecords()
	recs := []*Record{}

	cursor, err := coll.Find(db, &bson.M{
		"domain": d.Id,
	}, options.Find().
		SetSort(bson.D{{"sub_domain", 1}}),
	)
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

		if skipDeleted && (rec.Operation == DELETE ||
			!rec.DeleteTimestamp.IsZero()) {

			continue
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
