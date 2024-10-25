package deployment

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
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

func GetAll(db *database.Database, query *bson.M) (
	deplys []*Deployment, err error) {

	coll := db.Deployments()
	deplys = []*Deployment{}

	cursor, err := coll.Find(db, query)
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

func GetAllIds(db *database.Database) (deplyIds set.Set, err error) {
	coll := db.Deployments()
	deplyIds = set.NewSet()

	cursor, err := coll.Find(
		db,
		bson.M{},
		&options.FindOptions{
			Projection: bson.M{
				"_id": 1,
			},
		},
	)
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

		deplyIds.Add(deply.Id)
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

func RemoveMulti(db *database.Database, serviceId primitive.ObjectID,
	unitId primitive.ObjectID, deplyIds []primitive.ObjectID) (err error) {

	deplys, err := GetAll(db, &bson.M{
		"_id": &bson.M{
			"$in": deplyIds,
		},
		"service": serviceId,
		"unit":    unitId,
	})
	if err != nil {
		return
	}

	for _, deply := range deplys {
		inst, e := instance.Get(db, deply.Instance)
		if e != nil {
			err = e
			if _, ok := e.(*database.NotFoundError); ok {
				err = nil
				inst = nil
			} else {
				return
			}
		}

		if inst != nil {
			if inst.DeleteProtection {
				continue
			}

			err = instance.Delete(db, inst.Id)
			if err != nil {
				if _, ok := e.(*database.NotFoundError); !ok {
					err = nil
				} else {
					return
				}
			}
		}

		err = Remove(db, deply.Id)
		if err != nil {
			if _, ok := e.(*database.NotFoundError); !ok {
				err = nil
			} else {
				return
			}
		}
	}

	return
}
