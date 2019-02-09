package vpc

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Get(db *database.Database, vcId primitive.ObjectID) (
	vc *Vpc, err error) {

	coll := db.Vpcs()
	vc = &Vpc{}

	err = coll.FindOneId(vcId, vc)
	if err != nil {
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, vcId primitive.ObjectID) (
	vc *Vpc, err error) {

	coll := db.Vpcs()
	vc = &Vpc{}

	err = coll.FindOne(db, &bson.M{
		"_id":          vcId,
		"organization": orgId,
	}).Decode(vc)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func ExistsOrg(db *database.Database, orgId, vcId primitive.ObjectID) (
	exists bool, err error) {

	coll := db.Vpcs()
	n, err := coll.Count(
		db,
		&bson.M{
			"_id":          vcId,
			"organization": orgId,
		},
	)
	if err != nil {
		return
	}

	if n > 0 {
		exists = true
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	vcs []*Vpc, err error) {

	coll := db.Vpcs()
	vcs = []*Vpc{}

	cursor, err := coll.Find(
		db,
		query,
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		vc := &Vpc{}
		err = cursor.Decode(vc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		vcs = append(vcs, vc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllNames(db *database.Database, query *bson.M) (
	vpcs []*Vpc, err error) {

	coll := db.Vpcs()
	vpcs = []*Vpc{}

	cursor, err := coll.Find(
		db,
		query,
		&options.FindOptions{
			Sort: &bson.D{
				{"name", 1},
			},
			Projection: &bson.D{
				{"name", 1},
				{"organization", 1},
				{"type", 1},
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		vc := &Vpc{}
		err = cursor.Decode(vc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		vpcs = append(vpcs, vc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (vcs []*Vpc, count int64, err error) {

	coll := db.Vpcs()
	vcs = []*Vpc{}

	count, err = coll.Count(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	page = utils.Min64(page, count/pageCount)
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
		vc := &Vpc{}
		err = cursor.Decode(vc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		vcs = append(vcs, vc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetIds(db *database.Database, ids []primitive.ObjectID) (
	vcs []*Vpc, err error) {

	coll := db.Vpcs()
	vcs = []*Vpc{}

	cursor, err := coll.Find(
		db,
		&bson.M{
			"_id": &bson.M{
				"$in": ids,
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		vc := &Vpc{}
		err = cursor.Decode(vc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		vcs = append(vcs, vc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetDatacenter(db *database.Database, dcId primitive.ObjectID) (
	vcs []*Vpc, err error) {

	coll := db.Vpcs()
	vcs = []*Vpc{}

	cursor, err := coll.Find(
		db,
		&bson.M{
			"datacenter": dcId,
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		vc := &Vpc{}
		err = cursor.Decode(vc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		vcs = append(vcs, vc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func DistinctIds(db *database.Database, matchIds []primitive.ObjectID) (
	idsSet set.Set, err error) {

	coll := db.Images()
	idsSet = set.NewSet()

	idsInf, err := coll.Distinct(
		db,
		"_id",
		&bson.M{
			"_id": &bson.M{
				"$in": matchIds,
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	for _, idInf := range idsInf {
		if id, ok := idInf.(primitive.ObjectID); ok {
			idsSet.Add(id)
		}
	}

	return
}

func Remove(db *database.Database, vcId primitive.ObjectID) (err error) {
	coll := db.VpcsIp()

	_, err = coll.DeleteMany(db, &bson.M{
		"vpc": vcId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Vpcs()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": vcId,
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

func RemoveOrg(db *database.Database, orgId, vcId primitive.ObjectID) (err error) {
	coll := db.VpcsIp()

	_, err = coll.DeleteMany(db, &bson.M{
		"vpc": vcId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Vpcs()

	_, err = coll.DeleteOne(db, &bson.M{
		"organization": orgId,
		"_id":          vcId,
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

func RemoveMulti(db *database.Database, vcIds []primitive.ObjectID) (err error) {
	coll := db.VpcsIp()

	_, err = coll.DeleteMany(db, &bson.M{
		"vpc": &bson.M{
			"$in": vcIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Vpcs()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": vcIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveInstanceIps(db *database.Database, instId primitive.ObjectID) (
	err error) {

	coll := db.VpcsIp()

	_, err = coll.UpdateMany(db, &bson.M{
		"instance": instId,
	}, &bson.M{
		"$set": &bson.M{
			"instance": nil,
		},
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

func RemoveInstanceIp(db *database.Database, instId,
	vpcId primitive.ObjectID) (err error) {

	coll := db.VpcsIp()

	_, err = coll.UpdateOne(
		db,
		&bson.M{
			"vpc":      vpcId,
			"instance": instId,
		},
		&bson.M{
			"$set": &bson.M{
				"instance": nil,
			},
		},
	)
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
