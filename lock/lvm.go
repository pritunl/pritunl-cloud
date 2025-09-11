package lock

import (
	"fmt"
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
)

type LvmLocker struct {
	Id        string        `bson:"_id" json:"_id"`
	Node      bson.ObjectID `bson:"node" json:"node"`
	Timestamp time.Time     `bson:"timestamp" json:"timestamp"`
}

func LvmLock(db *database.Database, vgName, lvName string) (
	acquired bool, err error) {

	coll := db.LvmLock()

	doc := &LvmLocker{
		Id:        fmt.Sprintf("%s/%s", vgName, lvName),
		Node:      node.Self.Id,
		Timestamp: time.Now(),
	}

	_, err = coll.InsertOne(db, doc)
	if err != nil {
		err = database.ParseError(err)
		if _, ok := err.(*database.DuplicateKeyError); ok {
			err = nil
			return
		}
		return
	}

	acquired = true

	return
}

func LvmRelock(db *database.Database, vgName, lvName string) (err error) {
	coll := db.LvmLock()

	_, err = coll.UpdateOne(db, &bson.M{
		"id":   fmt.Sprintf("%s/%s", vgName, lvName),
		"node": node.Self.Id,
	}, &bson.M{
		"timestamp": time.Now(),
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func LvmUnlock(db *database.Database, vgName, lvName string) (err error) {
	coll := db.LvmLock()

	_, err = coll.DeleteOne(db, &bson.M{
		"id":   fmt.Sprintf("%s/%s", vgName, lvName),
		"node": node.Self.Id,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
