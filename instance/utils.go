package instance

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vpc"
	"gopkg.in/mgo.v2/bson"
)

func Get(db *database.Database, instId bson.ObjectId) (
	inst *Instance, err error) {

	coll := db.Instances()
	inst = &Instance{}

	err = coll.FindOneId(instId, inst)
	if err != nil {
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, instId bson.ObjectId) (
	inst *Instance, err error) {

	coll := db.Instances()
	inst = &Instance{}

	err = coll.FindOne(&bson.M{
		"_id":          instId,
		"organization": orgId,
	}, inst)
	if err != nil {
		return
	}

	return
}

func ExistsOrg(db *database.Database, orgId, instId bson.ObjectId) (
	exists bool, err error) {

	coll := db.Instances()

	n, err := coll.Find(&bson.M{
		"_id":          instId,
		"organization": orgId,
	}).Count()
	if err != nil {
		return
	}

	if n > 0 {
		exists = true
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	insts []*Instance, err error) {

	coll := db.Instances()
	insts = []*Instance{}

	cursor := coll.Find(query).Iter()

	inst := &Instance{}
	for cursor.Next(inst) {
		insts = append(insts, inst)
		inst = &Instance{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllVirt(db *database.Database, query *bson.M, disks []*disk.Disk) (
	insts []*Instance, err error) {

	instanceDisks := map[bson.ObjectId][]*disk.Disk{}
	for _, dsk := range disks {
		if dsk.State != disk.Available && dsk.State != disk.Snapshot {
			continue
		}

		dsks := instanceDisks[dsk.Instance]
		if dsks == nil {
			dsks = []*disk.Disk{}
		}
		instanceDisks[dsk.Instance] = append(dsks, dsk)
	}

	coll := db.Instances()
	insts = []*Instance{}

	cursor := coll.Find(query).Iter()

	inst := &Instance{}
	for cursor.Next(inst) {
		inst.LoadVirt(instanceDisks[inst.Id])
		insts = append(insts, inst)
		inst = &Instance{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllName(db *database.Database, query *bson.M) (
	instances []*Instance, err error) {

	coll := db.Instances()
	instances = []*Instance{}

	cursor := coll.Find(query).Select(&bson.M{
		"name": 1,
	}).Iter()

	inst := &Instance{}
	for cursor.Next(inst) {
		instances = append(instances, inst)
		inst = &Instance{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M, page, pageCount int) (
	insts []*Instance, count int, err error) {

	coll := db.Instances()
	insts = []*Instance{}

	qury := coll.Find(query)

	count, err = qury.Count()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	skip := utils.Min(page*pageCount, utils.Max(0, count-pageCount))

	cursor := qury.Sort("name").Skip(skip).Limit(pageCount).Iter()

	inst := &Instance{}
	for cursor.Next(inst) {
		insts = append(insts, inst)
		inst = &Instance{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, instId bson.ObjectId) (err error) {
	coll := db.Instances()

	err = vpc.RemoveInstanceIps(db, instId)
	if err != nil {
		return
	}

	err = coll.Remove(&bson.M{
		"_id": instId,
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

func Delete(db *database.Database, instId bson.ObjectId) (err error) {
	coll := db.Instances()

	err = coll.UpdateId(instId, &bson.M{
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

func DeleteOrg(db *database.Database, orgId, instId bson.ObjectId) (
	err error) {

	coll := db.Instances()

	err = coll.UpdateId(instId, &bson.M{
		"$set": &bson.M{
			"state": Destroy,
		},
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func DeleteMulti(db *database.Database, instIds []bson.ObjectId) (err error) {
	coll := db.Instances()

	_, err = coll.UpdateAll(&bson.M{
		"_id": &bson.M{
			"$in": instIds,
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
	instIds []bson.ObjectId) (err error) {

	coll := db.Instances()

	_, err = coll.UpdateAll(&bson.M{
		"_id": &bson.M{
			"$in": instIds,
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

func UpdateMulti(db *database.Database, instIds []bson.ObjectId,
	doc *bson.M) (err error) {

	coll := db.Instances()

	_, err = coll.UpdateAll(&bson.M{
		"_id": &bson.M{
			"$in": instIds,
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

func UpdateMultiOrg(db *database.Database, orgId bson.ObjectId,
	instIds []bson.ObjectId, doc *bson.M) (err error) {

	coll := db.Instances()

	_, err = coll.UpdateAll(&bson.M{
		"_id": &bson.M{
			"$in": instIds,
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
