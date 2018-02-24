package disk

import (
	"fmt"
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

	skip := utils.Min(page*pageCount, utils.Max(0, count-pageCount))

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

func UpdateMulti(db *database.Database, dskIds []bson.ObjectId,
	doc *bson.M) (err error) {

	coll := db.Disks()

	_, err = coll.UpdateAll(&bson.M{
		"_id": &bson.M{
			"$in": dskIds,
		},
	}, &bson.M{
		"$set": doc,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
