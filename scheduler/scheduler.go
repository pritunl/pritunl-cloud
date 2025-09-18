package scheduler

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
)

type Scheduler struct {
	Id            bson.ObjectID         `bson:"_id" json:"id"`
	Organization  bson.ObjectID         `bson:"organization" json:"organization"`
	Pod           bson.ObjectID         `bson:"pod" json:"pod"`
	Kind          string                `bson:"kind" json:"kind"`
	Created       time.Time             `bson:"created" json:"created"`
	Modified      time.Time             `bson:"modified" json:"modified"`
	Count         int                   `bson:"count" json:"count"`
	Spec          bson.ObjectID         `bson:"spec" json:"spec"`
	OverrideCount int                   `bson:"override_count" json:"override_count"`
	Consumed      int                   `bson:"consumed" json:"consumed"`
	Tickets       TicketsStore          `bson:"tickets" json:"tickets"`
	Failures      map[bson.ObjectID]int `bson:"failures" json:"failures"`
}

type Ticket struct {
	Node   bson.ObjectID `bson:"n" json:"n"`
	Offset int           `bson:"t" json:"t"`
}

type TicketsStore map[bson.ObjectID][]*Ticket

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

func (s *Scheduler) ClearTickets(db *database.Database) (err error) {
	coll := db.Schedulers()
	schd := &Scheduler{}

	err = coll.FindOneAndUpdate(db, bson.M{
		"_id": s.Id,
	}, bson.M{
		"$unset": bson.M{
			"tickets." + node.Self.Id.Hex(): "",
		},
	}, options.FindOneAndUpdate().SetReturnDocument(
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

func (s *Scheduler) Failure(db *database.Database) (limit bool, err error) {
	coll := db.Schedulers()
	schd := &Scheduler{}

	if s.Failures == nil {
		s.Failures = map[bson.ObjectID]int{}
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

func (s *Scheduler) Ready() bool {
	if s.Failures == nil {
		return true
	}
	return s.Failures[node.Self.Id] < settings.Hypervisor.MaxDeploymentFailures
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
