package advisory

import (
	"slices"

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

func CountResource(resourceId bson.ObjectID, advisories []*Advisory) (
	count, maxScore int) {

	for _, adv := range advisories {
		if adv.Dismissed {
			continue
		}

		if slices.Contains(adv.DismissedResources, resourceId) {
			continue
		}

		if adv.Score >= High {
			count += 1
		}
		if adv.Score > maxScore {
			maxScore = adv.Score
		}
	}

	return
}

func UpdateResource(db *database.Database, resId bson.ObjectID) (err error) {
	coll := db.Instances()

	count, err := coll.CountDocuments(db, &bson.M{
		"_id": resId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	if count > 0 {
		err = UpdateInstance(db, resId)
		if err != nil {
			return
		}
	}

	coll = db.Nodes()

	count, err = coll.CountDocuments(db, &bson.M{
		"_id": resId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	if count > 0 {
		err = UpdateNode(db, resId)
		if err != nil {
			return
		}
	}

	return
}

func UpdateResourceOrg(db *database.Database,
	resId, orgId bson.ObjectID) (err error) {

	coll := db.Instances()

	count, err := coll.CountDocuments(db, &bson.M{
		"_id":          resId,
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	if count > 0 {
		err = UpdateInstance(db, resId)
		if err != nil {
			return
		}
	}

	return
}

func UpdateInstance(db *database.Database, instId bson.ObjectID) (err error) {
	coll := db.Advisories()

	cursor, err := coll.Find(db, &bson.M{
		"instances": instId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	count := 0
	maxScore := 0
	for cursor.Next(db) {
		adv := &Advisory{}
		err = cursor.Decode(adv)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		if adv.Dismissed {
			continue
		}

		if slices.Contains(adv.DismissedResources, instId) {
			continue
		}

		if adv.Score >= High {
			count += 1
		}
		if adv.Score > maxScore {
			maxScore = adv.Score
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Instances()

	_, err = coll.UpdateOne(db, &bson.M{
		"_id": instId,
	}, &bson.M{
		"$set": &bson.M{
			"advisory_count": count,
			"advisory_max":   maxScore,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		} else {
			return
		}
	}

	return
}

func UpdateNode(db *database.Database, nodeId bson.ObjectID) (err error) {
	coll := db.Advisories()

	cursor, err := coll.Find(db, &bson.M{
		"nodes": nodeId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	count := 0
	maxScore := 0
	for cursor.Next(db) {
		adv := &Advisory{}
		err = cursor.Decode(adv)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		if adv.Dismissed {
			continue
		}

		if slices.Contains(adv.DismissedResources, nodeId) {
			continue
		}

		if adv.Score >= High {
			count += 1
		}
		if adv.Score > maxScore {
			maxScore = adv.Score
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Nodes()

	_, err = coll.UpdateOne(db, &bson.M{
		"_id": nodeId,
	}, &bson.M{
		"$set": &bson.M{
			"advisory_count": count,
			"advisory_max":   maxScore,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		} else {
			return
		}
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

func UpdateDismiss(db *database.Database, advId bson.ObjectID,
	dismiss, restore bool, dismissals, restores []bson.ObjectID) (err error) {

	adv, err := Get(db, advId)
	if err != nil {
		return
	}

	update := adv.buildDismissUpdate(dismiss, restore, dismissals, restores)
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

	if dismiss || restore {
		for _, ndeId := range adv.Nodes {
			err = UpdateNode(db, ndeId)
			if err != nil {
				return
			}
		}
		for _, instId := range adv.Instances {
			err = UpdateInstance(db, instId)
			if err != nil {
				return
			}
		}
	}

	if len(dismissals) > 0 {
		for _, resourceId := range dismissals {
			err = UpdateResource(db, resourceId)
			if err != nil {
				return
			}
		}
	}

	if len(restores) > 0 {
		for _, resourceId := range restores {
			err = UpdateResource(db, resourceId)
			if err != nil {
				return
			}
		}
	}

	return
}

func UpdateDismissOrg(db *database.Database, orgId, advId bson.ObjectID,
	dismiss, restore bool, dismissals, restores []bson.ObjectID) (err error) {

	adv, err := GetOrg(db, orgId, advId)
	if err != nil {
		return
	}

	update := adv.buildDismissUpdate(dismiss, restore, dismissals, restores)
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

	if dismiss || restore {
		for _, ndeId := range adv.Nodes {
			err = UpdateNode(db, ndeId)
			if err != nil {
				return
			}
		}
		for _, instId := range adv.Instances {
			err = UpdateInstance(db, instId)
			if err != nil {
				return
			}
		}
	}

	if len(dismissals) > 0 {
		for _, resourceId := range dismissals {
			err = UpdateResourceOrg(db, resourceId, orgId)
			if err != nil {
				return
			}
		}
	}

	if len(restores) > 0 {
		for _, resourceId := range restores {
			err = UpdateResourceOrg(db, resourceId, orgId)
			if err != nil {
				return
			}
		}
	}

	return
}

func UpdateMulti(db *database.Database, advIds []bson.ObjectID,
	dismiss, restore bool) (err error) {

	if !dismiss && !restore {
		return
	}

	coll := db.Advisories()

	_, err = coll.UpdateMany(db, &bson.M{
		"_id": &bson.M{
			"$in": advIds,
		},
	}, &bson.M{
		"$set": &bson.M{
			"dismissed": dismiss,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func UpdateMultiOrg(db *database.Database, orgId bson.ObjectID,
	advIds []bson.ObjectID, dismiss, restore bool) (err error) {

	if !dismiss && !restore {
		return
	}

	coll := db.Advisories()

	_, err = coll.UpdateMany(db, &bson.M{
		"_id": &bson.M{
			"$in": advIds,
		},
		"organization": orgId,
	}, &bson.M{
		"$set": &bson.M{
			"dismissed": dismiss,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
