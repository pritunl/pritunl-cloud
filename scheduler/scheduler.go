package scheduler

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
)

type Scheduler struct {
	Id            Resource                   `bson:"_id" json:"id"`
	Kind          string                     `bson:"kind" json:"kind"`
	Created       time.Time                  `bson:"created" json:"created"`
	Modified      time.Time                  `bson:"modified" json:"modified"`
	Count         int                        `bson:"count" json:"count"`
	Spec          primitive.ObjectID         `bson:"spec" json:"spec"`
	OverrideCount int                        `bson:"override_count" json:"override_count"`
	Consumed      int                        `bson:"consumed" json:"consumed"`
	Tickets       TicketsStore               `bson:"tickets" json:"tickets"`
	Failures      map[primitive.ObjectID]int `bson:"failures" json:"failures"`
}

type Resource struct {
	Pod  primitive.ObjectID `bson:"pod,omitempty" json:"pod"`
	Unit primitive.ObjectID `bson:"unit,omitempty" json:"unit"`
}

type Ticket struct {
	Node   primitive.ObjectID `bson:"n" json:"n"`
	Offset int                `bson:"t" json:"t"`
}

type TicketsStore map[primitive.ObjectID][]*Ticket

func (s *Scheduler) Refresh(db *database.Database) (exists bool, err error) {
	coll := db.Schedulers()
	schd := &Scheduler{}

	err = coll.FindOne(db, bson.M{
		"_id": s.Id,
	}, database.FindOneProject(
		"count",
		"consumed",
		"tickets",
		"failures",
	)).Decode(schd)
	if err != nil {
		err = database.ParseError(err)
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		} else {
			return
		}
		return
	}

	exists = true
	s.Count = schd.Count
	s.Consumed = schd.Consumed
	s.Tickets = schd.Tickets
	s.Failures = schd.Failures

	return
}

func (s *Scheduler) Failure(db *database.Database) (limit bool, err error) {
	coll := db.Schedulers()
	schd := &Scheduler{}

	if s.Failures == nil {
		s.Failures = map[primitive.ObjectID]int{}
	}
	s.Failures[node.Self.Id] += 1

	update := bson.M{
		"$inc": bson.M{
			"failures." + node.Self.Id.Hex(): 1,
		},
	}

	if s.Failures[node.Self.Id] >= settings.Hypervisor.MaxDeploymentFailures {
		limit = true
		update["$unset"] = bson.M{
			"tickets." + node.Self.Id.Hex(): "",
		}
	}

	err = coll.FindOneAndUpdate(db, bson.M{
		"_id": s.Id,
	}, update, options.FindOneAndUpdate().SetReturnDocument(
		options.After)).Decode(schd)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	s.Count = schd.Count
	s.Consumed = schd.Consumed
	s.Tickets = schd.Tickets
	s.Failures = schd.Failures

	return
}

func (s *Scheduler) Consume(db *database.Database) (err error) {
	coll := db.Schedulers()
	schd := &Scheduler{}

	err = coll.FindOneAndUpdate(db, bson.M{
		"_id": s.Id,
		"$expr": bson.M{
			"$lt": []interface{}{"$consumed", "$count"},
		},
	}, bson.M{
		"$set": bson.M{
			"modified": time.Now(),
		},
		"$inc": bson.M{
			"consumed": 1,
		},
	}, options.FindOneAndUpdate().SetReturnDocument(
		options.After)).Decode(schd)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	s.Count = schd.Count
	s.Consumed = schd.Consumed
	s.Failures = schd.Failures

	return
}

func (s *Scheduler) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	return
}

func (s *Scheduler) Commit(db *database.Database) (err error) {
	coll := db.Schedulers()

	err = coll.Commit(s.Id, s)
	if err != nil {
		return
	}

	return
}

func (s *Scheduler) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Schedulers()

	err = coll.CommitFields(s.Id, s, fields)
	if err != nil {
		return
	}

	return
}

func (s *Scheduler) Insert(db *database.Database) (err error) {
	coll := db.Schedulers()

	_, err = coll.InsertOne(db, s)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
