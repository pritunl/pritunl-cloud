package advisory

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Get(db *database.Database, advId bson.ObjectID) (
	adv *Advisory, err error) {

	coll := db.Advisories()
	adv = &Advisory{}

	err = coll.FindOneId(advId, adv)
	if err != nil {
		return
	}

	return
}

func GetOne(db *database.Database, query *bson.M) (adv *Advisory, err error) {
	coll := db.Advisories()
	adv = &Advisory{}

	err = coll.FindOne(db, query).Decode(adv)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, advId bson.ObjectID) (
	adv *Advisory, err error) {

	coll := db.Advisories()
	adv = &Advisory{}

	err = coll.FindOne(db, &bson.M{
		"_id":          advId,
		"organization": orgId,
	}).Decode(adv)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	advisories []*Advisory, err error) {

	coll := db.Advisories()
	advisories = []*Advisory{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		adv := &Advisory{}
		err = cursor.Decode(adv)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		advisories = append(advisories, adv)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetInstance(db *database.Database, instId bson.ObjectID) (
	advisories []*Advisory, err error) {

	coll := db.Advisories()
	advisories = []*Advisory{}

	cursor, err := coll.Find(db, &bson.M{
		"instances": instId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		adv := &Advisory{}
		err = cursor.Decode(adv)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		advisories = append(advisories, adv)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetNode(db *database.Database, nodeId bson.ObjectID) (
	advisories []*Advisory, err error) {

	coll := db.Advisories()
	advisories = []*Advisory{}

	cursor, err := coll.Find(db, &bson.M{
		"nodes": nodeId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		adv := &Advisory{}
		err = cursor.Decode(adv)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		advisories = append(advisories, adv)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (advisories []*Advisory, count int64, err error) {

	coll := db.Advisories()
	advisories = []*Advisory{}

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

	if pageCount == 0 {
		pageCount = 20
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
		options.Find().
			SetSort(bson.D{{"reference", 1}}).
			SetSkip(skip).
			SetLimit(pageCount),
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		adv := &Advisory{}
		err = cursor.Decode(adv)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		advisories = append(advisories, adv)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, advId bson.ObjectID) (err error) {
	coll := db.Advisories()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": advId,
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

func RemoveOrg(db *database.Database, orgId, advId bson.ObjectID) (
	err error) {

	coll := db.Advisories()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id":          advId,
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

func RemoveMulti(db *database.Database, advIds []bson.ObjectID) (
	err error) {

	coll := db.Advisories()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": advIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveMultiOrg(db *database.Database, orgId bson.ObjectID,
	advIds []bson.ObjectID) (err error) {

	coll := db.Advisories()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": advIds,
		},
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func buildDismissUpdate(adv *Advisory, dismiss, restore bool,
	dismissals, restores []bson.ObjectID) (update bson.M) {

	setDoc := bson.M{}

	if dismiss {
		setDoc["dismissed"] = true
	} else if restore {
		setDoc["dismissed"] = false
	}

	if len(dismissals) > 0 || len(restores) > 0 {
		known := set.NewSet()
		for _, instId := range adv.Instances {
			known.Add(instId)
		}
		for _, nodeId := range adv.Nodes {
			known.Add(nodeId)
		}

		dismissed := set.NewSet()
		for _, resourceId := range adv.DismissedResources {
			dismissed.Add(resourceId)
		}

		for _, resourceId := range dismissals {
			if known.Contains(resourceId) {
				dismissed.Add(resourceId)
			}
		}
		for _, resourceId := range restores {
			dismissed.Remove(resourceId)
		}

		newDismissals := []bson.ObjectID{}
		for resourceIdInf := range dismissed.Iter() {
			newDismissals = append(
				newDismissals, resourceIdInf.(bson.ObjectID))
		}

		setDoc["dismissed_resources"] = newDismissals
	}

	if len(setDoc) == 0 {
		return
	}

	update = bson.M{
		"$set": setDoc,
	}

	return
}

func UpdateDismiss(db *database.Database, advId bson.ObjectID,
	dismiss, restore bool, dismissals, restores []bson.ObjectID) (err error) {

	adv, err := Get(db, advId)
	if err != nil {
		return
	}

	update := buildDismissUpdate(adv, dismiss, restore, dismissals, restores)
	if update == nil {
		return
	}

	coll := db.Advisories()

	_, err = coll.UpdateOne(db, &bson.M{
		"_id": advId,
	}, update)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func UpdateDismissOrg(db *database.Database, orgId, advId bson.ObjectID,
	dismiss, restore bool, dismissals, restores []bson.ObjectID) (err error) {

	adv, err := GetOrg(db, orgId, advId)
	if err != nil {
		return
	}

	update := buildDismissUpdate(adv, dismiss, restore, dismissals, restores)
	if update == nil {
		return
	}

	coll := db.Advisories()

	_, err = coll.UpdateOne(db, &bson.M{
		"_id":          advId,
		"organization": orgId,
	}, update)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
