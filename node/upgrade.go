package node

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
)

type zoneUgradeDoc struct {
	Datacenter primitive.ObjectID `bson:"datacenter"`
	Zone       primitive.ObjectID `bson:"zone"`
}

func dataUpgradeZone(db *database.Database) (err error) {
	zoneColl := db.Zones()

	zoneDatacenterMap := make(map[primitive.ObjectID]primitive.ObjectID)

	getDatacenterForZone := func(zoneID primitive.ObjectID) (primitive.ObjectID, error) {
		if datacenterID, ok := zoneDatacenterMap[zoneID]; ok {
			return datacenterID, nil
		}

		zne := &zoneUgradeDoc{}
		err := zoneColl.FindOne(db, bson.M{
			"_id": zoneID,
		}).Decode(zne)
		if err != nil {
			return primitive.NilObjectID, database.ParseError(err)
		}

		zoneDatacenterMap[zoneID] = zne.Datacenter
		return zne.Datacenter, nil
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
			bson.M{"_id": cursor.Current.Lookup("_id").ObjectID()},
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
			bson.M{"_id": cursor.Current.Lookup("_id").ObjectID()},
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
			bson.M{"_id": cursor.Current.Lookup("_id").ObjectID()},
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
			bson.M{"_id": cursor.Current.Lookup("_id").ObjectID()},
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
			bson.M{"_id": cursor.Current.Lookup("_id").ObjectID()},
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

	return nil
}
