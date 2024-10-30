package service

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Get(db *database.Database, serviceId primitive.ObjectID) (
	srvc *Service, err error) {

	coll := db.Services()
	srvc = &Service{}

	err = coll.FindOneId(serviceId, srvc)
	if err != nil {
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, srvcId primitive.ObjectID) (
	srvc *Service, err error) {

	coll := db.Services()
	srvc = &Service{}

	err = coll.FindOne(db, &bson.M{
		"_id":          srvcId,
		"organization": orgId,
	}).Decode(srvc)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOne(db *database.Database, query *bson.M) (srvc *Service, err error) {
	coll := db.Services()
	srvc = &Service{}

	err = coll.FindOne(db, query).Decode(srvc)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	services []*Service, err error) {

	coll := db.Services()
	services = []*Service{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		srvc := &Service{}
		err = cursor.Decode(srvc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		services = append(services, srvc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (services []*Service, count int64, err error) {

	coll := db.Services()
	services = []*Service{}

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
		srvc := &Service{}
		err = cursor.Decode(srvc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		services = append(services, srvc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, serviceId primitive.ObjectID) (err error) {
	coll := db.Services()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id":               serviceId,
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

func RemoveOrg(db *database.Database, orgId, serviceId primitive.ObjectID) (
	err error) {

	coll := db.Services()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id":          serviceId,
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

func RemoveMulti(db *database.Database, serviceIds []primitive.ObjectID) (
	err error) {

	coll := db.Services()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": serviceIds,
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
	serviceIds []primitive.ObjectID) (err error) {

	coll := db.Services()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": serviceIds,
		},
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
