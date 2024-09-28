package scheduler

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type Scheduler struct {
	Id       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Kind     string             `bson:"kind" json:"kind"`
	Created  time.Time          `bson:"created" json:"created"`
	Modified time.Time          `bson:"modified" json:"modified"`
	Count    int                `bson:"int" json:"int"`
	Consumed int                `bson:"consumed" json:"consumed"`
	Tickets  TicketsStore       `bson:"tickets" json:"tickets"`
}

type Ticket struct {
	Node   primitive.ObjectID `bson:"n" json:"n"`
	Offset int                `bson:"t" json:"t"`
}

type TicketsStore map[primitive.ObjectID][]*Ticket

func (s *Scheduler) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	return
}

func (s *Scheduler) Commit(db *database.Database) (err error) {
	coll := db.Deployments()

	err = coll.Commit(s.Id, s)
	if err != nil {
		return
	}

	return
}

func (s *Scheduler) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Deployments()

	err = coll.CommitFields(s.Id, s, fields)
	if err != nil {
		return
	}

	return
}

func (s *Scheduler) Insert(db *database.Database) (err error) {
	coll := db.Deployments()

	resp, err := coll.InsertOne(db, s)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	s.Id = resp.InsertedID.(primitive.ObjectID)

	return
}
