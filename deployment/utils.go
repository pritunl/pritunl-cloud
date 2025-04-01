package deployment

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/journal"
)

func Get(db *database.Database, deplyId primitive.ObjectID) (
	deply *Deployment, err error) {

	coll := db.Deployments()
	deply = &Deployment{}

	err = coll.FindOneId(deplyId, deply)
	if err != nil {
		return
	}

	return
}

func GetUnit(db *database.Database, unitId, deplyId primitive.ObjectID) (
	deply *Deployment, err error) {

	coll := db.Deployments()
	deply = &Deployment{}

	err = coll.FindOne(db, &bson.M{
		"_id":  deplyId,
		"unit": unitId,
	}).Decode(deply)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, unitId primitive.ObjectID) (
	deply *Deployment, err error) {

	coll := db.Deployments()
	deply = &Deployment{}

	err = coll.FindOne(db, &bson.M{
		"_id":          unitId,
		"organization": orgId,
	}).Decode(deply)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetUnitOrg(db *database.Database,
	orgId, unitId, deplyId primitive.ObjectID) (
	deply *Deployment, err error) {

	coll := db.Deployments()
	deply = &Deployment{}

	err = coll.FindOne(db, &bson.M{
		"_id":          deplyId,
		"unit":         unitId,
		"organization": orgId,
	}).Decode(deply)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	deplys []*Deployment, err error) {

	coll := db.Deployments()
	deplys = []*Deployment{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		deply := &Deployment{}
		err = cursor.Decode(deply)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		deplys = append(deplys, deply)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllSorted(db *database.Database, query *bson.M) (
	deplys []*Deployment, err error) {

	coll := db.Deployments()
	deplys = []*Deployment{}

	cursor, err := coll.Find(
		db,
		query,
		&options.FindOptions{
			Sort: &bson.D{
				{"timestamp", -1},
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		deply := &Deployment{}
		err = cursor.Decode(deply)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		deplys = append(deplys, deply)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllActiveIds(db *database.Database) (deplyIds set.Set, err error) {
	coll := db.Deployments()
	deplyIds = set.NewSet()

	cursor, err := coll.Find(
		db,
		bson.M{
			"state": bson.M{
				"$in": []string{
					Reserved,
					Deployed,
				},
			},
		},
		&options.FindOptions{
			Projection: bson.M{
				"_id": 1,
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		deply := &Deployment{}
		err = cursor.Decode(deply)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		deplyIds.Add(deply.Id)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllStates(db *database.Database) (
	deplysMap map[primitive.ObjectID]*Deployment, err error) {

	coll := db.Deployments()
	deplysMap = map[primitive.ObjectID]*Deployment{}

	cursor, err := coll.Find(
		db,
		bson.M{},
		&options.FindOptions{
			Projection: bson.M{
				"_id":    1,
				"state":  1,
				"action": 1,
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		deply := &Deployment{}
		err = cursor.Decode(deply)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		deplysMap[deply.Id] = deply
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveDomains(db *database.Database, deplyId primitive.ObjectID) (
	err error) {

	recs, err := domain.GetRecordAll(db, &bson.M{
		"deployment": deplyId,
	})
	if err != nil {
		return
	}

	domnIdsSet := set.NewSet()
	for _, rec := range recs {
		domnIdsSet.Add(rec.Domain)
	}

	domnIds := []primitive.ObjectID{}
	for domnIdInf := range domnIdsSet.Iter() {
		domnIds = append(domnIds, domnIdInf.(primitive.ObjectID))
	}

	if len(domnIds) > 0 {
		domns, e := domain.GetAll(db, &bson.M{
			"_id": &bson.M{
				"$in": domnIds,
			},
		})
		if e != nil {
			err = e
			return
		}

		for _, domn := range domns {
			err = domn.LoadRecords(db)
			if err != nil {
				return
			}

			domn.PreCommit()

			changed := false
			for _, rec := range domn.Records {
				if rec.Deployment == deplyId {
					changed = true
					rec.Operation = domain.DELETE
				}
			}

			if changed {
				err = domn.CommitRecords(db)
				if err != nil {
					return
				}
			}
		}
	}

	event.PublishDispatch(db, "domain.change")

	return
}

func Remove(db *database.Database, deplyId primitive.ObjectID) (err error) {
	coll := db.Deployments()

	err = journal.Remove(db, deplyId, journal.DeploymentAgent)
	if err != nil {
		return
	}

	err = RemoveDomains(db, deplyId)
	if err != nil {
		return
	}

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": deplyId,
	})
	if err != nil {
		err = database.ParseError(err)
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		} else {
			return
		}
	}

	event.PublishDispatch(db, "domain.change")
	event.PublishDispatch(db, "pod.change")

	return
}

func RemoveMulti(db *database.Database, podId primitive.ObjectID,
	unitId primitive.ObjectID, deplyIds []primitive.ObjectID) (err error) {

	coll := db.Deployments()

	_, err = coll.UpdateMany(db, &bson.M{
		"_id": &bson.M{
			"$in": deplyIds,
		},
		"pod":  podId,
		"unit": unitId,
	}, &bson.M{
		"$set": &bson.M{
			"action": Destroy,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func ArchiveMulti(db *database.Database, podId primitive.ObjectID,
	unitId primitive.ObjectID, deplyIds []primitive.ObjectID) (err error) {

	coll := db.Deployments()

	_, err = coll.UpdateMany(db, &bson.M{
		"_id": &bson.M{
			"$in": deplyIds,
		},
		"pod":   podId,
		"unit":  unitId,
		"state": Deployed,
	}, &bson.M{
		"$set": &bson.M{
			"action": Archive,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RestoreMulti(db *database.Database, podId primitive.ObjectID,
	unitId primitive.ObjectID, deplyIds []primitive.ObjectID) (err error) {

	coll := db.Deployments()

	_, err = coll.UpdateMany(db, &bson.M{
		"_id": &bson.M{
			"$in": deplyIds,
		},
		"pod":   podId,
		"unit":  unitId,
		"state": Archived,
	}, &bson.M{
		"$set": &bson.M{
			"action": Restore,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
