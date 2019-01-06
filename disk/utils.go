package disk

import (
	"fmt"
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
)

func Get(db *database.Database, diskId bson.ObjectId) (dsk *Disk, err error) {
	coll := db.Disks()
	dsk = &Disk{}

	err = coll.FindOneId(diskId, dsk)
	if err != nil {
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, diskId bson.ObjectId) (
	dsk *Disk, err error) {

	coll := db.Disks()
	dsk = &Disk{}

	err = coll.FindOne(&bson.M{
		"_id":          diskId,
		"organization": orgId,
	}, dsk)
	if err != nil {
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	disks []*Disk, err error) {

	coll := db.Disks()
	disks = []*Disk{}

	cursor := coll.Find(query).Iter()

	nde := &Disk{}
	for cursor.Next(nde) {
		disks = append(disks, nde)
		nde = &Disk{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M, page, pageCount int) (
	disks []*Disk, count int, err error) {

	coll := db.Disks()
	disks = []*Disk{}

	qury := coll.Find(query)

	count, err = qury.Count()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	page = utils.Min(page, count/pageCount)
	skip := utils.Min(page*pageCount, count)

	cursor := qury.Sort("name").Skip(skip).Limit(pageCount).Iter()

	disk := &Disk{}
	for cursor.Next(disk) {
		disks = append(disks, disk)
		disk = &Disk{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetInstance(db *database.Database, instId bson.ObjectId) (
	disks []*Disk, err error) {

	coll := db.Disks()
	disks = []*Disk{}

	cursor := coll.Find(&bson.M{
		"instance": instId,
	}).Sort("index").Iter()

	dsk := &Disk{}
	for cursor.Next(dsk) {
		disks = append(disks, dsk)
		dsk = &Disk{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetInstanceIndex(db *database.Database, instId bson.ObjectId,
	index string) (dsk *Disk, err error) {

	coll := db.Disks()
	dsk = &Disk{}

	err = coll.FindOne(&bson.M{
		"instance": instId,
		"index":    index,
	}, dsk)
	if err != nil {
		return
	}

	return
}

func GetNode(db *database.Database, nodeId bson.ObjectId) (
	disks []*Disk, err error) {

	coll := db.Disks()
	disks = []*Disk{}

	cursor := coll.Find(&bson.M{
		"node": nodeId,
	}).Iter()

	dsk := &Disk{}
	for cursor.Next(dsk) {
		disks = append(disks, dsk)
		dsk = &Disk{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, diskId bson.ObjectId) (err error) {
	coll := db.Disks()

	err = coll.Remove(&bson.M{
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

func Detach(db *database.Database, dskIds bson.ObjectId) (err error) {
	coll := db.Disks()

	err = coll.UpdateId(dskIds, &bson.M{
		"$set": &bson.M{
			"instance": "",
			"index":    fmt.Sprintf("hold_%s", bson.NewObjectId().Hex()),
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Delete(db *database.Database, dskId bson.ObjectId) (err error) {
	coll := db.Disks()

	err = coll.UpdateId(dskId, &bson.M{
		"$set": &bson.M{
			"state": Destroy,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func DeleteOrg(db *database.Database, orgId, dskId bson.ObjectId) (err error) {
	coll := db.Disks()

	err = coll.Update(&bson.M{
		"_id":          dskId,
		"organization": orgId,
	}, &bson.M{
		"$set": &bson.M{
			"state": Destroy,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func DeleteMulti(db *database.Database, dskIds []bson.ObjectId) (err error) {
	coll := db.Disks()

	_, err = coll.UpdateAll(&bson.M{
		"_id": &bson.M{
			"$in": dskIds,
		},
	}, &bson.M{
		"$set": &bson.M{
			"state": Destroy,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func DeleteMultiOrg(db *database.Database, orgId bson.ObjectId,
	dskIds []bson.ObjectId) (err error) {

	coll := db.Disks()

	_, err = coll.UpdateAll(&bson.M{
		"_id": &bson.M{
			"$in": dskIds,
		},
		"organization": orgId,
	}, &bson.M{
		"$set": &bson.M{
			"state": Destroy,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func UpdateMulti(db *database.Database, dskIds []bson.ObjectId,
	doc *bson.M) (err error) {

	coll := db.Disks()

	query := &bson.M{
		"_id": &bson.M{
			"$in": dskIds,
		},
	}

	if (*doc)["state"] == Destroy {
		(*query)["delete_protection"] = &bson.M{
			"$ne": true,
		}
	}

	_, err = coll.UpdateAll(query, &bson.M{
		"$set": doc,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func UpdateMultiOrg(db *database.Database, orgId bson.ObjectId,
	dskIds []bson.ObjectId, doc *bson.M) (err error) {

	coll := db.Disks()

	query := &bson.M{
		"_id": &bson.M{
			"$in": dskIds,
		},
		"organization": orgId,
	}

	if (*doc)["state"] == Destroy {
		(*query)["delete_protection"] = &bson.M{
			"$ne": true,
		}
	}

	_, err = coll.UpdateAll(&bson.M{
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

func GetAllKeys(db *database.Database, ndeId bson.ObjectId) (
	keys set.Set, err error) {

	coll := db.Disks()
	keys = set.NewSet()

	cursor := coll.Find(&bson.M{
		"node": ndeId,
	}).Select(&bson.M{
		"node":          1,
		"backing_image": 1,
	}).Iter()

	dsk := &Disk{}
	for cursor.Next(dsk) {
		if dsk.BackingImage != "" {
			keys.Add(dsk.BackingImage)
		}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
