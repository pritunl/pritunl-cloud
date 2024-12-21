package instance

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/sirupsen/logrus"
)

func GetInstanceMtu(jumboFrames bool, vxlan bool) int {
	mtuSize := 0
	if jumboFrames {
		mtuSize = settings.Hypervisor.JumboMtu
	} else {
		mtuSize = settings.Hypervisor.NormalMtu
	}

	if vxlan {
		mtuSize -= 54
	}

	return mtuSize
}

func Get(db *database.Database, instId primitive.ObjectID) (
	inst *Instance, err error) {

	coll := db.Instances()
	inst = &Instance{}

	err = coll.FindOneId(instId, inst)
	if err != nil {
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, instId primitive.ObjectID) (
	inst *Instance, err error) {

	coll := db.Instances()
	inst = &Instance{}

	err = coll.FindOne(db, &bson.M{
		"_id":          instId,
		"organization": orgId,
	}).Decode(inst)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOne(db *database.Database, query *bson.M) (inst *Instance, err error) {
	coll := db.Instances()
	inst = &Instance{}

	err = coll.FindOne(db, query).Decode(inst)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func ExistsOrg(db *database.Database, orgId, instId primitive.ObjectID) (
	exists bool, err error) {

	coll := db.Instances()

	n, err := coll.CountDocuments(db, &bson.M{
		"_id":          instId,
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
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

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		inst := &Instance{}
		err = cursor.Decode(inst)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		insts = append(insts, inst)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllVirt(db *database.Database, query *bson.M,
	pools []*pool.Pool, disks []*disk.Disk) (
	insts []*Instance, err error) {

	poolsMap := map[primitive.ObjectID]*pool.Pool{}
	for _, pl := range pools {
		poolsMap[pl.Id] = pl
	}

	instanceDisks := map[primitive.ObjectID][]*disk.Disk{}
	if disks != nil {
		for _, dsk := range disks {
			if dsk.State == disk.Destroy && dsk.DeleteProtection {
				logrus.WithFields(logrus.Fields{
					"disk_id": dsk.Id.Hex(),
				}).Info("instance: Delete protection ignore disk detach")
			} else if dsk.State != disk.Available &&
				dsk.State != disk.Snapshot &&
				dsk.State != disk.Backup &&
				dsk.State != disk.Restore &&
				dsk.State != disk.Expand {

				continue
			}

			dsks := instanceDisks[dsk.Instance]
			if dsks == nil {
				dsks = []*disk.Disk{}
			}
			instanceDisks[dsk.Instance] = append(dsks, dsk)
		}
	}

	coll := db.Instances()
	insts = []*Instance{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		inst := &Instance{}
		err = cursor.Decode(inst)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		inst.LoadVirt(poolsMap, instanceDisks[inst.Id])
		insts = append(insts, inst)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllVirtMapped(db *database.Database, query *bson.M,
	pools []*pool.Pool, instanceDisks map[primitive.ObjectID][]*disk.Disk) (
	insts []*Instance, err error) {

	coll := db.Instances()
	insts = []*Instance{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	poolsMap := map[primitive.ObjectID]*pool.Pool{}
	for _, pl := range pools {
		poolsMap[pl.Id] = pl
	}

	for cursor.Next(db) {
		inst := &Instance{}
		err = cursor.Decode(inst)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		virtDsks := []*disk.Disk{}

		dsks := instanceDisks[inst.Id]
		if dsks != nil {
			for _, dsk := range dsks {
				if dsk.State == disk.Destroy && dsk.DeleteProtection {
					logrus.WithFields(logrus.Fields{
						"disk_id": dsk.Id.Hex(),
					}).Info("instance: Delete protection ignore disk detach")
				} else if dsk.State != disk.Available &&
					dsk.State != disk.Snapshot &&
					dsk.State != disk.Backup &&
					dsk.State != disk.Restore &&
					dsk.State != disk.Expand {

					continue
				}

				virtDsks = append(virtDsks, dsk)
			}
		}

		inst.LoadVirt(poolsMap, virtDsks)
		insts = append(insts, inst)
	}

	err = cursor.Err()
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

	cursor, err := coll.Find(
		db,
		query,
		&options.FindOptions{
			Projection: &bson.D{
				{"name", 1},
			},
		},
	)
	defer cursor.Close(db)

	for cursor.Next(db) {
		inst := &Instance{}
		err = cursor.Decode(inst)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		instances = append(instances, inst)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (insts []*Instance, count int64, err error) {

	coll := db.Instances()
	insts = []*Instance{}

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
		inst := &Instance{}
		err = cursor.Decode(inst)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		insts = append(insts, inst)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, instId primitive.ObjectID) (err error) {
	inst, err := Get(db, instId)
	if err != nil {
		return
	}

	if inst.DeleteProtection {
		logrus.WithFields(logrus.Fields{
			"instance_id": instId.Hex(),
		}).Info("instance: Delete protection ignore instance remove")
		return
	}

	err = block.RemoveInstanceIps(db, instId)
	if err != nil {
		return
	}

	err = vpc.RemoveInstanceIps(db, instId)
	if err != nil {
		return
	}

	coll := db.InstancesAgent()

	_, err = coll.DeleteMany(db, &bson.M{
		"i": instId,
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

	coll = db.InstancesKmsg()

	_, err = coll.DeleteMany(db, &bson.M{
		"i": instId,
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

	coll = db.Instances()

	_, err = coll.DeleteOne(db, &bson.M{
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

func Delete(db *database.Database, instId primitive.ObjectID) (err error) {
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

func DeleteOrg(db *database.Database, orgId, instId primitive.ObjectID) (
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

func DeleteMulti(db *database.Database, instIds []primitive.ObjectID) (
	err error) {

	coll := db.Instances()

	_, err = coll.UpdateMany(db, &bson.M{
		"_id": &bson.M{
			"$in": instIds,
		},
		"delete_protection": &bson.M{
			"$ne": true,
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

func DeleteMultiOrg(db *database.Database, orgId primitive.ObjectID,
	instIds []primitive.ObjectID) (err error) {

	coll := db.Instances()

	_, err = coll.UpdateMany(db, &bson.M{
		"_id": &bson.M{
			"$in": instIds,
		},
		"organization": orgId,
		"delete_protection": &bson.M{
			"$ne": true,
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

func UpdateMulti(db *database.Database, instIds []primitive.ObjectID,
	doc *bson.M) (err error) {

	coll := db.Instances()

	query := &bson.M{
		"_id": &bson.M{
			"$in": instIds,
		},
	}

	if (*doc)["state"] == Destroy {
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

func UpdateMultiOrg(db *database.Database, orgId primitive.ObjectID,
	instIds []primitive.ObjectID, doc *bson.M) (err error) {

	coll := db.Instances()

	query := &bson.M{
		"_id": &bson.M{
			"$in": instIds,
		},
		"organization": orgId,
	}

	if (*doc)["state"] == Destroy {
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

func SetState(db *database.Database, instId primitive.ObjectID,
	state string) (err error) {

	coll := db.Instances()

	_, err = coll.UpdateOne(db, &bson.M{
		"_id": instId,
	}, &bson.M{
		"$set": &bson.M{
			"state": state,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
