package instance

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/journal"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/sirupsen/logrus"
)

func Get(db *database.Database, instId bson.ObjectID) (
	inst *Instance, err error) {

	coll := db.Instances()
	inst = &Instance{}

	err = coll.FindOneId(instId, inst)
	if err != nil {
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, instId bson.ObjectID) (
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

func ExistsIp(db *database.Database, addr string) (exists bool, err error) {
	coll := db.Instances()

	n, err := coll.CountDocuments(db, &bson.M{
		"$or": []bson.M{
			{"public_ips": addr},
			{"public_ips6": addr},
			{"cloud_private_ips": addr},
			{"cloud_public_ips": addr},
			{"cloud_public_ips6": addr},
			{"host_ips": addr},
			{"node_port_ips": addr},
		},
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

func ExistsOrg(db *database.Database, orgId, instId bson.ObjectID) (
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

func GetAllRoles(db *database.Database, query *bson.M) (
	insts []*Instance, rolesSet set.Set, err error) {

	coll := db.Instances()
	insts = []*Instance{}
	rolesSet = set.NewSet()

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

		for _, role := range inst.Roles {
			rolesSet.Add(role)
		}
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

	poolsMap := map[bson.ObjectID]*pool.Pool{}
	for _, pl := range pools {
		poolsMap[pl.Id] = pl
	}

	instanceDisks := map[bson.ObjectID][]*disk.Disk{}
	if disks != nil {
		for _, dsk := range disks {
			if dsk.Action == disk.Destroy {
				if dsk.DeleteProtection {
					logrus.WithFields(logrus.Fields{
						"disk_id": dsk.Id.Hex(),
					}).Info("instance: Delete protection ignore disk detach")
				} else {
					continue
				}
			} else if dsk.State != disk.Available &&
				dsk.State != disk.Attached {

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
	pools []*pool.Pool, instanceDisks map[bson.ObjectID][]*disk.Disk) (
	insts []*Instance, err error) {

	coll := db.Instances()
	insts = []*Instance{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	poolsMap := map[bson.ObjectID]*pool.Pool{}
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
				if dsk.Action == disk.Destroy {
					if dsk.DeleteProtection {
						logrus.WithFields(logrus.Fields{
							"disk_id": dsk.Id.Hex(),
						}).Info("instance: Delete protection ignore disk detach")
					} else {
						continue
					}
				} else if dsk.State != disk.Available &&
					dsk.State != disk.Attached {

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

func LoadAllVirt(insts []*Instance, pools []*pool.Pool,
	instanceDisks map[bson.ObjectID][]*disk.Disk) []*Instance {

	poolsMap := map[bson.ObjectID]*pool.Pool{}
	for _, pl := range pools {
		poolsMap[pl.Id] = pl
	}

	for _, inst := range insts {
		virtDsks := []*disk.Disk{}

		dsks := instanceDisks[inst.Id]
		for _, dsk := range dsks {
			if dsk.Action == disk.Destroy {
				if dsk.DeleteProtection {
					logrus.WithFields(logrus.Fields{
						"disk_id": dsk.Id.Hex(),
					}).Info("instance: Delete protection ignore disk detach")
				} else {
					continue
				}
			} else if dsk.State != disk.Available &&
				dsk.State != disk.Attached {

				continue
			}

			virtDsks = append(virtDsks, dsk)
		}

		inst.LoadVirt(poolsMap, virtDsks)
	}

	return insts
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

func Remove(db *database.Database, instId bson.ObjectID) (err error) {
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

	err = journal.Remove(db, instId, journal.InstanceAgent)
	if err != nil {
		return
	}

	coll := db.Instances()

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

	_ = inst.Cleanup(db)

	return
}

func Delete(db *database.Database, instId bson.ObjectID) (err error) {
	coll := db.Instances()

	err = coll.UpdateId(instId, &bson.M{
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

func DeleteOrg(db *database.Database, orgId, instId bson.ObjectID) (
	err error) {

	coll := db.Instances()

	err = coll.UpdateId(instId, &bson.M{
		"$set": &bson.M{
			"action": Destroy,
		},
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func DeleteMulti(db *database.Database, instIds []bson.ObjectID) (
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
	instIds []bson.ObjectID) (err error) {

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
			"action": Destroy,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func UpdateMulti(db *database.Database, instIds []bson.ObjectID,
	doc *bson.M) (err error) {

	coll := db.Instances()

	query := &bson.M{
		"_id": &bson.M{
			"$in": instIds,
		},
		"action": &bson.M{
			"$ne": Destroy,
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
	instIds []bson.ObjectID, doc *bson.M) (err error) {

	coll := db.Instances()

	query := &bson.M{
		"_id": &bson.M{
			"$in": instIds,
		},
		"organization": orgId,
		"action": &bson.M{
			"$ne": Destroy,
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

func SetAction(db *database.Database, instId bson.ObjectID,
	action string) (err error) {

	coll := db.Instances()

	_, err = coll.UpdateOne(db, &bson.M{
		"_id": instId,
		"action": &bson.M{
			"$ne": Destroy,
		},
	}, &bson.M{
		"$set": &bson.M{
			"action": action,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func SetDownloadProgress(db *database.Database,
	instId bson.ObjectID, progress int, speedMb float64) (err error) {

	coll := db.Instances()

	_, err = coll.UpdateOne(db, &bson.M{
		"_id": instId,
	}, &bson.M{
		"$set": &bson.M{
			"status_info.download_progress": progress,
			"status_info.download_speed":    speedMb,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
