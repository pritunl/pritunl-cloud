package disk

import (
	"fmt"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Get(db *database.Database, diskId bson.ObjectID) (
	dsk *Disk, err error) {

	coll := db.Disks()
	dsk = &Disk{}

	err = coll.FindOneId(diskId, dsk)
	if err != nil {
		return
	}

	return
}

func GetOne(db *database.Database, query *bson.M) (dsk *Disk, err error) {
	coll := db.Disks()
	dsk = &Disk{}

	err = coll.FindOne(db, query).Decode(dsk)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, diskId bson.ObjectID) (
	dsk *Disk, err error) {

	coll := db.Disks()
	dsk = &Disk{}

	err = coll.FindOne(db, &bson.M{
		"_id":          diskId,
		"organization": orgId,
	}).Decode(dsk)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	disks []*Disk, err error) {

	coll := db.Disks()
	disks = []*Disk{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		nde := &Disk{}
		err = cursor.Decode(nde)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		disks = append(disks, nde)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllMap(db *database.Database, query *bson.M) (
	disks map[bson.ObjectID]*Disk, err error) {

	coll := db.Disks()
	disks = map[bson.ObjectID]*Disk{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		dsk := &Disk{}
		err = cursor.Decode(dsk)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		disks[dsk.Id] = dsk
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (disks []*Disk, count int64, err error) {

	coll := db.Disks()
	disks = []*Disk{}

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
		dsk := &Disk{}
		err = cursor.Decode(dsk)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		disks = append(disks, dsk)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetInstance(db *database.Database, instId bson.ObjectID) (
	disks []*Disk, err error) {

	coll := db.Disks()
	disks = []*Disk{}

	cursor, err := coll.Find(
		db,
		&bson.M{
			"instance": instId,
		},
		&options.FindOptions{
			Sort: &bson.D{
				{"index", 1},
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		dsk := &Disk{}
		err = cursor.Decode(dsk)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		disks = append(disks, dsk)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetInstanceIndex(db *database.Database, instId bson.ObjectID,
	index string) (dsk *Disk, err error) {

	coll := db.Disks()
	dsk = &Disk{}

	err = coll.FindOne(db, &bson.M{
		"instance": instId,
		"index":    index,
	}).Decode(dsk)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetNode(db *database.Database, nodeId bson.ObjectID,
	nodePools []bson.ObjectID) (disks []*Disk, err error) {

	coll := db.Disks()
	disks = []*Disk{}

	cursor, err := coll.Find(db, &bson.M{
		"$or": []bson.M{
			{"node": nodeId},
			{"pool": &bson.M{
				"$in": nodePools,
			}},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		dsk := &Disk{}
		err = cursor.Decode(dsk)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		disks = append(disks, dsk)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, diskId bson.ObjectID) (err error) {
	coll := db.Disks()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": diskId,
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

func Detach(db *database.Database, dskIds bson.ObjectID) (err error) {
	coll := db.Disks()

	err = coll.UpdateId(dskIds, &bson.M{
		"$set": &bson.M{
			"index":      fmt.Sprintf("hold_%s", bson.NewObjectID().Hex()),
			"instance":   Vacant,
			"deployment": Vacant,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Delete(db *database.Database, dskId bson.ObjectID) (err error) {
	coll := db.Disks()

	err = coll.UpdateId(dskId, &bson.M{
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

func DeleteOrg(db *database.Database, orgId, dskId bson.ObjectID) (
	err error) {

	coll := db.Disks()

	_, err = coll.UpdateOne(db, &bson.M{
		"_id":          dskId,
		"organization": orgId,
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

func DeleteMulti(db *database.Database, dskIds []bson.ObjectID) (
	err error) {

	coll := db.Disks()

	_, err = coll.UpdateMany(db, &bson.M{
		"_id": &bson.M{
			"$in": dskIds,
		},
		"delete_protection": &bson.M{
			"$ne": true,
		},
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

func DeleteMultiOrg(db *database.Database, orgId bson.ObjectID,
	dskIds []bson.ObjectID) (err error) {

	coll := db.Disks()

	_, err = coll.UpdateMany(db, &bson.M{
		"_id": &bson.M{
			"$in": dskIds,
		},
		"organization": orgId,
		"delete_protection": &bson.M{
			"$ne": true,
		},
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

func UpdateMulti(db *database.Database, dskIds []bson.ObjectID,
	doc *bson.M) (err error) {

	coll := db.Disks()

	query := &bson.M{
		"_id": &bson.M{
			"$in": dskIds,
		},
	}

	if (*doc)["action"] == Destroy {
		(*query)["delete_protection"] = &bson.M{
			"$ne": true,
		}
	}

	_, err = coll.UpdateMany(db, query, &bson.M{
		"$set": doc,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func UpdateMultiOrg(db *database.Database, orgId bson.ObjectID,
	dskIds []bson.ObjectID, doc *bson.M) (err error) {

	coll := db.Disks()

	query := &bson.M{
		"_id": &bson.M{
			"$in": dskIds,
		},
		"organization": orgId,
	}

	if (*doc)["action"] == Destroy {
		(*query)["delete_protection"] = &bson.M{
			"$ne": true,
		}
	}

	_, err = coll.UpdateMany(db, &bson.M{
		"_id": &bson.M{
			"$in": dskIds,
		},
		"organization": orgId,
	}, &bson.M{
		"$set": doc,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllKeys(db *database.Database, ndeId bson.ObjectID) (
	keys set.Set, err error) {

	coll := db.Disks()
	keys = set.NewSet()

	cursor, err := coll.Find(db, &bson.M{
		"node": ndeId,
	}, &options.FindOptions{
		Projection: &bson.D{
			{"node", 1},
			{"backing_image", 1},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		dsk := &Disk{}
		err = cursor.Decode(dsk)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		if dsk.BackingImage != "" {
			keys.Add(dsk.BackingImage)
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func SetDeleteProtection(db *database.Database, instId bson.ObjectID,
	protection bool) (err error) {

	coll := db.Disks()

	_, err = coll.UpdateMany(db, &bson.M{
		"instance": instId,
	}, &bson.M{
		"$set": &bson.M{
			"delete_protection": protection,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
