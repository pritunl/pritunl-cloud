package scheduler

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/service"
	"github.com/pritunl/pritunl-cloud/spec"
)

func Exists(db *database.Database, schdId Resource) (
	exists bool, err error) {

	coll := db.Schedulers()
	schd := &Scheduler{}

	err = coll.FindOneId(schdId, schd)
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		} else {
			return
		}
		return
	}

	exists = true
	return
}

func Get(db *database.Database, schdId Resource) (
	schd *Scheduler, err error) {

	coll := db.Schedulers()
	schd = &Scheduler{}

	err = coll.FindOneId(schdId, schd)
	if err != nil {
		return
	}

	return
}

func GetAll(db *database.Database) (schds []*Scheduler, err error) {
	coll := db.Schedulers()
	schds = []*Scheduler{}

	cursor, err := coll.Find(db, bson.M{})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		schd := &Scheduler{}
		err = cursor.Decode(schd)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		schds = append(schds, schd)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllActive(db *database.Database) (schds []*Scheduler, err error) {
	coll := db.Schedulers()
	schds = []*Scheduler{}

	cursor, err := coll.Find(db, bson.M{})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		schd := &Scheduler{}
		err = cursor.Decode(schd)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		if schd.Consumed < schd.Count {
			schds = append(schds, schd)
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, schdId Resource) (err error) {
	coll := db.Schedulers()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": schdId,
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

func Schedule(db *database.Database, unit *service.Unit) (err error) {
	exists, e := Exists(db, Resource{
		Service: unit.Service.Id,
		Unit:    unit.Id,
	})
	if e != nil {
		err = e
		return
	}

	if exists {
		return
	}

	spc, err := spec.Get(db, unit.DeployCommit)
	if err != nil {
		return
	}

	switch unit.Kind {
	case spec.InstanceKind:
		schd := NewInstanceUnit(unit, spc)
		err = schd.Schedule(db)
		if err != nil {
			return
		}
	}

	return
}
