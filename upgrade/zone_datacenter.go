package upgrade

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
)

type zoneUgradeDoc struct {
	Id         bson.ObjectID `bson:"_id"`
	Node       bson.ObjectID `bson:"node"`
	Datacenter bson.ObjectID `bson:"datacenter"`
	Zone       bson.ObjectID `bson:"zone"`
}

func zoneDatacenterUpgrade(db *database.Database) (err error) {
	zoneColl := db.Zones()

	zoneDatacenterMap := make(map[bson.ObjectID]bson.ObjectID)
	nodeMap := make(map[bson.ObjectID]*zoneUgradeDoc)

	getDatacenterForZone := func(zoneID bson.ObjectID) (
		bson.ObjectID, error) {

		if datacenterID, ok := zoneDatacenterMap[zoneID]; ok {
			return datacenterID, nil
		}

		zne := &zoneUgradeDoc{}
		err := zoneColl.FindOne(db, bson.M{
			"_id": zoneID,
		}).Decode(zne)
		if err != nil {
			return bson.NilObjectID, database.ParseError(err)
		}

		zoneDatacenterMap[zoneID] = zne.Datacenter
		return zne.Datacenter, nil
	}

	getNode := func(nodeId bson.ObjectID) (
		*zoneUgradeDoc, error) {

		if nde, ok := nodeMap[nodeId]; ok {
			return nde, nil
		}

		coll := db.Nodes()

		nde := &zoneUgradeDoc{}
		err := coll.FindOne(db, bson.M{
			"_id": nodeId,
		}).Decode(nde)
		if err != nil {
			return nil, database.ParseError(err)
		}

		nodeMap[nodeId] = nde
		return nde, nil
	}

	coll := db.Nodes()
	cursor, err := coll.Find(
		db,
		bson.M{
			"zone":       bson.M{"$exists": true},
			"datacenter": bson.M{"$exists": false},
		},
	)
	if err != nil {
		return database.ParseError(err)
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		doc := &zoneUgradeDoc{}
		err = cursor.Decode(doc)
		if err != nil {
			return database.ParseError(err)
		}

		datacenterID, err := getDatacenterForZone(doc.Zone)
		if err != nil {
			return err
		}

		_, err = coll.UpdateOne(
			db,
			bson.M{"_id": doc.Id},
			bson.M{"$set": bson.M{"datacenter": datacenterID}},
		)
		if err != nil {
			return database.ParseError(err)
		}
	}
	err = cursor.Err()
	if err != nil {
		return database.ParseError(err)
	}

	coll = db.Deployments()
	cursor, err = coll.Find(
		db,
		bson.M{
			"zone":       bson.M{"$exists": true},
			"datacenter": bson.M{"$exists": false},
		},
	)
	if err != nil {
		return database.ParseError(err)
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		doc := &zoneUgradeDoc{}
		err = cursor.Decode(doc)
		if err != nil {
			return database.ParseError(err)
		}

		datacenterID, err := getDatacenterForZone(doc.Zone)
		if err != nil {
			return err
		}

		_, err = coll.UpdateOne(
			db,
			bson.M{"_id": doc.Id},
			bson.M{"$set": bson.M{"datacenter": datacenterID}},
		)
		if err != nil {
			return database.ParseError(err)
		}
	}
	err = cursor.Err()
	if err != nil {
		return database.ParseError(err)
	}

	coll = db.Instances()
	cursor, err = coll.Find(
		db,
		bson.M{
			"zone":       bson.M{"$exists": true},
			"datacenter": bson.M{"$exists": false},
		},
	)
	if err != nil {
		return database.ParseError(err)
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		doc := &zoneUgradeDoc{}
		err = cursor.Decode(doc)
		if err != nil {
			return database.ParseError(err)
		}

		datacenterID, err := getDatacenterForZone(doc.Zone)
		if err != nil {
			return err
		}

		_, err = coll.UpdateOne(
			db,
			bson.M{"_id": doc.Id},
			bson.M{"$set": bson.M{"datacenter": datacenterID}},
		)
		if err != nil {
			return database.ParseError(err)
		}
	}
	err = cursor.Err()
	if err != nil {
		return database.ParseError(err)
	}

	coll = db.Pools()
	cursor, err = coll.Find(
		db,
		bson.M{
			"zone":       bson.M{"$exists": true},
			"datacenter": bson.M{"$exists": false},
		},
	)
	if err != nil {
		return database.ParseError(err)
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		doc := &zoneUgradeDoc{}
		err = cursor.Decode(doc)
		if err != nil {
			return database.ParseError(err)
		}

		datacenterID, err := getDatacenterForZone(doc.Zone)
		if err != nil {
			return err
		}

		_, err = coll.UpdateOne(
			db,
			bson.M{"_id": doc.Id},
			bson.M{"$set": bson.M{"datacenter": datacenterID}},
		)
		if err != nil {
			return database.ParseError(err)
		}
	}
	err = cursor.Err()
	if err != nil {
		return database.ParseError(err)
	}

	coll = db.Specs()
	cursor, err = coll.Find(
		db,
		bson.M{
			"zone":       bson.M{"$exists": true},
			"datacenter": bson.M{"$exists": false},
		},
	)
	if err != nil {
		return database.ParseError(err)
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		doc := &zoneUgradeDoc{}
		err = cursor.Decode(doc)
		if err != nil {
			return database.ParseError(err)
		}

		datacenterID, err := getDatacenterForZone(doc.Zone)
		if err != nil {
			return err
		}

		_, err = coll.UpdateOne(
			db,
			bson.M{"_id": doc.Id},
			bson.M{"$set": bson.M{"datacenter": datacenterID}},
		)
		if err != nil {
			return database.ParseError(err)
		}
	}
	err = cursor.Err()
	if err != nil {
		return database.ParseError(err)
	}

	coll = db.Disks()
	cursor, err = coll.Find(
		db,
		bson.M{
			"zone":       bson.M{"$exists": false},
			"datacenter": bson.M{"$exists": false},
		},
	)
	if err != nil {
		return database.ParseError(err)
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		doc := &zoneUgradeDoc{}
		err = cursor.Decode(doc)
		if err != nil {
			return database.ParseError(err)
		}

		nde, err := getNode(doc.Node)
		if err != nil {
			return err
		}

		_, err = coll.UpdateOne(
			db,
			bson.M{"_id": doc.Id},
			bson.M{"$set": bson.M{
				"datacenter": nde.Datacenter,
				"zone":       nde.Zone,
			}},
		)
		if err != nil {
			return database.ParseError(err)
		}
	}
	err = cursor.Err()
	if err != nil {
		return database.ParseError(err)
	}

	return nil
}
