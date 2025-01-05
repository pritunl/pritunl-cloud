package policy

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
)

func Get(db *database.Database, policyId primitive.ObjectID) (
	polcy *Policy, err error) {

	coll := db.Policies()
	polcy = &Policy{}

	err = coll.FindOneId(policyId, polcy)
	if err != nil {
		return
	}

	return
}

func GetService(db *database.Database, podId primitive.ObjectID) (
	policies []*Policy, err error) {

	coll := db.Policies()
	policies = []*Policy{}

	cursor, err := coll.Find(
		db,
		&bson.M{
			"pods": podId,
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		polcy := &Policy{}
		err = cursor.Decode(polcy)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		policies = append(policies, polcy)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetRoles(db *database.Database, roles []string) (
	policies []*Policy, err error) {

	coll := db.Policies()
	policies = []*Policy{}

	if roles == nil {
		roles = []string{}
	}

	cursor, err := coll.Find(
		db,
		&bson.M{
			"roles": &bson.M{
				"$in": roles,
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		polcy := &Policy{}
		err = cursor.Decode(polcy)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		policies = append(policies, polcy)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAll(db *database.Database) (policies []*Policy, err error) {
	coll := db.Policies()
	policies = []*Policy{}

	cursor, err := coll.Find(
		db,
		&bson.M{},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		polcy := &Policy{}
		err = cursor.Decode(polcy)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		policies = append(policies, polcy)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, policyId primitive.ObjectID) (err error) {
	coll := db.Policies()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": policyId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
