package domain

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/sirupsen/logrus"
)

func Refresh(db *database.Database, domnId bson.ObjectID) {
	coll := db.Domains()
	domn := &Domain{}

	err := coll.FindOne(db, &bson.M{
		"_id": domnId,
	}).Decode(domn)
	if err != nil {
		err = database.ParseError(err)
		logrus.WithFields(logrus.Fields{
			"domain": domn.Id.Hex(),
			"error":  err,
		}).Error("domain: Domain refresh failed to find domain")
		return
	}

	if domn.Locked() {
		logrus.WithFields(logrus.Fields{
			"domain": domn.Id.Hex(),
		}).Info("domain: Skipping refresh on locked domain")
		return
	}

	err = domn.LoadRecords(db, false)
	if err != nil {
		return
	}

	err = domn.CommitRecordsSilent(db)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": domn.Id.Hex(),
			"error":  err,
		}).Error("domain: Domain refresh failed")
		return
	}

	deleteTtl := time.Duration(settings.System.DomainDeleteTtl) * time.Second
	now := time.Now()
	for _, rec := range domn.Records {
		if rec.IsDeleted() && now.Sub(rec.DeleteTimestamp) > deleteTtl {
			_, err = coll.DeleteOne(db, &bson.M{
				"_id":              rec.Id,
				"delete_timestamp": rec.DeleteTimestamp,
			})
			if err != nil {
				err = database.ParseError(err)
				if _, ok := err.(*database.NotFoundError); ok {
					err = nil
				} else {
					return
				}
			}
		}
	}

	return
}

func Get(db *database.Database, domnId bson.ObjectID) (
	domn *Domain, err error) {

	coll := db.Domains()
	domn = &Domain{}

	err = coll.FindOneId(domnId, domn)
	if err != nil {
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, domnId bson.ObjectID) (
	domn *Domain, err error) {

	coll := db.Domains()
	domn = &Domain{}

	err = coll.FindOne(db, &bson.M{
		"_id":          domnId,
		"organization": orgId,
	}).Decode(domn)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOne(db *database.Database, query *bson.M) (domn *Domain, err error) {
	coll := db.Domains()
	domn = &Domain{}

	err = coll.FindOne(db, query).Decode(domn)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func ExistsOrg(db *database.Database, orgId, domnId bson.ObjectID) (
	exists bool, err error) {

	coll := db.Domains()

	n, err := coll.CountDocuments(db, &bson.M{
		"_id":          domnId,
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if n > 0 {
		exists = true
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	domns []*Domain, err error) {

	coll := db.Domains()
	domns = []*Domain{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		dmn := &Domain{}
		err = cursor.Decode(dmn)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		domns = append(domns, dmn)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetLoadedAllIds(db *database.Database, domnIds []bson.ObjectID) (
	domns []*Domain, err error) {

	coll := db.DomainsRecords()
	domainRecs := map[bson.ObjectID][]*Record{}

	cursor, err := coll.Find(db, &bson.M{
		"domain": &bson.M{
			"$in": domnIds,
		},
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

		domainRecs[rec.Domain] = append(domainRecs[rec.Domain], rec)
	}

	coll = db.Domains()
	domns = []*Domain{}

	cursor, err = coll.Find(db, &bson.M{
		"_id": &bson.M{
			"$in": domnIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		dmn := &Domain{}
		err = cursor.Decode(dmn)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		dmn.preloadRecords(domainRecs[dmn.Id])

		domns = append(domns, dmn)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func PreloadedRecords(domns []*Domain, recs []*Record) []*Domain {
	domainRecs := map[bson.ObjectID][]*Record{}
	for _, rec := range recs {
		domainRecs[rec.Domain] = append(domainRecs[rec.Domain], rec)
	}

	for _, domn := range domns {
		domn.preloadRecords(domainRecs[domn.Id])
	}

	return domns
}

func GetAllName(db *database.Database, query *bson.M) (
	domns []*Domain, err error) {

	coll := db.Domains()
	domns = []*Domain{}

	cursor, err := coll.Find(
		db,
		query,
		options.Find().
			SetProjection(bson.D{
				{"name", 1},
				{"organization", 1},
			}),
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		dmn := &Domain{}
		err = cursor.Decode(dmn)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		domns = append(domns, dmn)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetRecordAll(db *database.Database, query *bson.M) (
	recs []*Record, err error) {

	coll := db.DomainsRecords()
	recs = []*Record{}

	cursor, err := coll.Find(db, query)
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

	return
}

func Lock(db *database.Database, domnId bson.ObjectID) (
	lockId bson.ObjectID, acquired bool, err error) {

	coll := db.Domains()

	newLockId := bson.NewObjectID()
	now := time.Now()
	ttl := now.Add(-time.Duration(
		settings.System.DomainLockTtl) * time.Second)

	resp, err := coll.UpdateOne(db, &bson.M{
		"_id": domnId,
		"$or": []bson.M{
			{"lock_id": Vacant},
			{"lock_timestamp": bson.M{"$lt": ttl}},
		},
	}, &bson.M{
		"$set": &bson.M{
			"lock_id":        newLockId,
			"lock_timestamp": now,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
			return
		}
		return
	}

	if resp.ModifiedCount > 0 {
		lockId = newLockId
		acquired = true
	}

	return
}

func Relock(db *database.Database, domnId,
	lockId bson.ObjectID) (err error) {

	coll := db.Domains()

	_, err = coll.UpdateOne(db, &bson.M{
		"_id":     domnId,
		"lock_id": lockId,
	}, &bson.M{
		"$set": &bson.M{
			"lock_timestamp": time.Now(),
		},
	})
	if err != nil {
		err = database.ParseError(err)
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
			return
		}
		return
	}

	return
}

func Unlock(db *database.Database, domnId,
	lockId bson.ObjectID) (err error) {

	coll := db.Domains()

	_, err = coll.UpdateOne(db, &bson.M{
		"_id":     domnId,
		"lock_id": lockId,
	}, &bson.M{
		"$set": &bson.M{
			"lock_id": Vacant,
		},
		"$unset": &bson.M{
			"lock_timestamp": 1,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
			return
		}
		return
	}

	return
}

func Remove(db *database.Database, domnId bson.ObjectID) (err error) {
	coll := db.Domains()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": domnId,
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

func RemoveOrg(db *database.Database, orgId, domnId bson.ObjectID) (
	err error) {

	coll := db.Domains()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id":          domnId,
		"organization": orgId,
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

func RemoveMulti(db *database.Database, domnIds []bson.ObjectID) (err error) {
	coll := db.Domains()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": domnIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveMultiOrg(db *database.Database, orgId bson.ObjectID,
	domnIds []bson.ObjectID) (err error) {

	coll := db.Domains()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": domnIds,
		},
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
