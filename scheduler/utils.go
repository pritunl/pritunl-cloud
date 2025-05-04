package scheduler

import (
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/unit"
)

func Exists(db *database.Database, schdId primitive.ObjectID) (
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

func Get(db *database.Database, schdId primitive.ObjectID) (
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

func Remove(db *database.Database, schdId primitive.ObjectID) (
	deleted bool, err error) {

	coll := db.Schedulers()

	resp, err := coll.DeleteOne(db, &bson.M{
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

	if resp.DeletedCount > 0 {
		deleted = true
	}

	return
}

func Schedule(db *database.Database, unt *unit.Unit) (err error) {
	exists, e := Exists(db, unt.Id)
	if e != nil {
		err = e
		return
	}

	if exists {
		return
	}

	spc, err := spec.Get(db, unt.DeploySpec)
	if err != nil {
		return
	}

	errData, err := spc.Refresh(db)
	if err != nil {
		return
	}

	if errData != nil {
		err = errData.GetError()
		return
	}

	switch unt.Kind {
	case deployment.Instance, deployment.Image:
		schd := NewInstanceUnit(unt, spc)
		err = schd.Schedule(db, 0)
		if err != nil {
			return
		}
	}

	return
}

func ManualSchedule(db *database.Database, unt *unit.Unit,
	specId primitive.ObjectID, count int) (
	errData *errortypes.ErrorData, err error) {

	exists, e := Exists(db, unt.Id)
	if e != nil {
		err = e
		return
	}

	if exists {
		errData = &errortypes.ErrorData{
			Error:   "scheduler_active",
			Message: "Cannot schedule deployments while scheduler is active",
		}
		return
	}

	if specId.IsZero() {
		specId = unt.DeploySpec
	}

	spc, err := spec.Get(db, specId)
	if err != nil {
		return
	}

	errData, err = spc.Refresh(db)
	if err != nil {
		return
	}

	if errData != nil {
		return
	}

	if spc.Unit != unt.Id {
		errData = &errortypes.ErrorData{
			Error:   "unit_deploy_spec_invalid",
			Message: "Invalid unit deployment commit",
		}
		return
	}

	switch unt.Kind {
	case deployment.Instance, deployment.Image:
		if unt.Kind == deployment.Image {
			count = 1
		}

		schd := NewInstanceUnit(unt, spc)
		err = schd.Schedule(db, count)
		if err != nil {
			return
		}
	default:
		err = &errortypes.ParseError{
			errors.Newf("scheduler: Unknown unit kind %s", unt.Kind),
		}
		return
	}

	return
}
