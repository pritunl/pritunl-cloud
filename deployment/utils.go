package deployment

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
)

func Get(db *database.Database, deplyId primitive.ObjectID) (
	deply *Deployment, err error) {

	coll := db.Deployments()
	deply = &Deployment{}

	err = coll.FindOneId(deplyId, deply)
	if err != nil {
		return
	}

	return
}

func GetAll(db *database.Database) (deplys []*Deployment, err error) {
	coll := db.Deployments()
	deplys = []*Deployment{}

	cursor, err := coll.Find(db, bson.M{})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		deply := &Deployment{}
		err = cursor.Decode(deply)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		deplys = append(deplys, deply)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, deplyId primitive.ObjectID) (err error) {
	coll := db.Deployments()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": deplyId,
	})
	if err != nil {
		err = database.ParseError(err)
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		} else {
			return
		}
	}

	return
}
