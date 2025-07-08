package pod

import (
	"time"

	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/unit"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Get(db *database.Database, podId primitive.ObjectID) (
	pd *Pod, err error) {

	coll := db.Pods()
	pd = &Pod{}

	err = coll.FindOneId(podId, pd)
	if err != nil {
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, pdId primitive.ObjectID) (
	pd *Pod, err error) {

	coll := db.Pods()
	pd = &Pod{}

	err = coll.FindOne(db, &bson.M{
		"_id":          pdId,
		"organization": orgId,
	}).Decode(pd)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOne(db *database.Database, query *bson.M) (pd *Pod, err error) {
	coll := db.Pods()
	pd = &Pod{}

	err = coll.FindOne(db, query).Decode(pd)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	pods []*Pod, err error) {

	coll := db.Pods()
	pods = []*Pod{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		pd := &Pod{}
		err = cursor.Decode(pd)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		pods = append(pods, pd)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (pods []*Pod, count int64, err error) {

	coll := db.Pods()
	pods = []*Pod{}

	if len(*query) == 0 {
		count, err = coll.EstimatedDocumentCount(db)
		if err != nil {
			err = database.ParseError(err)
			return
		}
	} else {
		count, err = coll.CountDocuments(db, query)
		if err != nil {
			err = database.ParseError(err)
			return
		}
	}

	maxPage := count / pageCount
	if count == pageCount {
		maxPage = 0
	}
	page = utils.Min64(page, maxPage)
	skip := utils.Min64(page*pageCount, count)

	cursor, err := coll.Find(
		db,
		query,
		&options.FindOptions{
			Sort: &bson.D{
				{"name", 1},
			},
			Skip:  &skip,
			Limit: &pageCount,
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		pd := &Pod{}
		err = cursor.Decode(pd)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		pods = append(pods, pd)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func UpdateDrafts(db *database.Database, podId, usrId primitive.ObjectID,
	drafts []*UnitDraft) (err error) {

	for _, draft := range drafts {
		draft.Timestamp = time.Now()
	}

	coll := db.Pods()
	_, err = coll.UpdateOne(db, &bson.M{
		"_id": podId,
	}, &bson.M{
		"$set": &bson.M{
			"drafts." + usrId.Hex(): drafts,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return nil
}

func UpdateDraftsOrg(db *database.Database, orgId, podId, usrId primitive.ObjectID,
	drafts []*UnitDraft) (err error) {

	for _, draft := range drafts {
		draft.Timestamp = time.Now()
	}

	coll := db.Pods()
	_, err = coll.UpdateOne(db, &bson.M{
		"_id":          podId,
		"organization": orgId,
	}, &bson.M{
		"$set": &bson.M{
			"drafts." + usrId.Hex(): drafts,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return nil
}

func Remove(db *database.Database, podId primitive.ObjectID) (err error) {
	coll := db.Pods()

	err = spec.RemoveAll(db, &bson.M{
		"pod": podId,
	})
	if err != nil {
		return
	}

	err = unit.RemoveAll(db, &bson.M{
		"pod": podId,
	})
	if err != nil {
		return
	}

	_, err = coll.DeleteOne(db, &bson.M{
		"_id":               podId,
		"delete_protection": false,
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

func RemoveOrg(db *database.Database, orgId, podId primitive.ObjectID) (
	err error) {

	coll := db.Pods()

	err = spec.RemoveAll(db, &bson.M{
		"pod":          podId,
		"organization": orgId,
	})
	if err != nil {
		return
	}

	err = unit.RemoveAll(db, &bson.M{
		"pod":          podId,
		"organization": orgId,
	})
	if err != nil {
		return
	}

	_, err = coll.DeleteOne(db, &bson.M{
		"_id":               podId,
		"organization":      orgId,
		"delete_protection": false,
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

func RemoveMulti(db *database.Database, podIds []primitive.ObjectID) (
	err error) {

	coll := db.Pods()

	err = spec.RemoveAll(db, &bson.M{
		"pod": &bson.M{
			"$in": podIds,
		},
	})
	if err != nil {
		return
	}

	err = unit.RemoveAll(db, &bson.M{
		"pod": &bson.M{
			"$in": podIds,
		},
	})
	if err != nil {
		return
	}

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": podIds,
		},
		"delete_protection": false,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveMultiOrg(db *database.Database, orgId primitive.ObjectID,
	podIds []primitive.ObjectID) (err error) {

	coll := db.Pods()

	err = spec.RemoveAll(db, &bson.M{
		"pod": &bson.M{
			"$in": podIds,
		},
		"organization": orgId,
	})
	if err != nil {
		return
	}

	err = unit.RemoveAll(db, &bson.M{
		"pod": &bson.M{
			"$in": podIds,
		},
		"organization": orgId,
	})
	if err != nil {
		return
	}

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": podIds,
		},
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
