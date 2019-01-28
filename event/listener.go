package event

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/database"
	"time"
)

type Listener struct {
	state    bool
	err      error
	channels []string
	stream   chan *Event
}

func (l *Listener) Listen() chan *Event {
	return l.stream
}

func (l *Listener) Close() {
	l.state = false
	close(l.stream)
}

func (l *Listener) sub(db *database.Database, cursorId primitive.ObjectID) {
	defer db.Close()
	coll := db.Events()

	var channelBson interface{}
	if len(l.channels) == 1 {
		channelBson = l.channels[0]
	} else {
		channelBson = &bson.M{
			"$in": l.channels,
		}
	}

	queryOpts := &options.FindOptions{
		Sort: &bson.D{
			{"$natural", 1},
		},
	}
	queryOpts.SetMaxAwaitTime(10 * time.Second)
	queryOpts.SetCursorType(options.TailableAwait)

	query := &bson.M{
		"_id": &bson.M{
			"$gt": cursorId,
		},
		"channel": channelBson,
	}

	var cursor mongo.Cursor
	var err error
	for {
		cursor, err = coll.Find(
			context.Background(),
			query,
			queryOpts,
		)
		if err != nil {
			err = database.ParseError(err)

			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("event: Listener find error")
		} else {
			break
		}

		if !l.state {
			return
		}

		time.Sleep(constants.RetryDelay)

		if !l.state {
			return
		}
	}

	defer func() {
		defer func() {
			recover()
		}()
		cursor.Close(context.Background())
	}()

	for {
		for cursor.Next(context.Background()) {
			msg := &Event{}
			err = cursor.Decode(msg)
			if err != nil {
				err = database.ParseError(err)

				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("event: Listener decode error")

				time.Sleep(constants.RetryDelay)
				break
			}

			cursorId = msg.Id

			if msg.Data == nil {
				// Blank msg for cursor
				continue
			}

			if !l.state {
				return
			}

			l.stream <- msg
		}

		if !l.state {
			return
		}

		err = cursor.Err()
		if err != nil {
			err = database.ParseError(err)

			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("event: Listener cursor error")

			time.Sleep(constants.RetryDelay)
		}

		if !l.state {
			return
		}

		cursor.Close(context.Background())
		db.Close()
		db = database.GetDatabase()
		coll = db.Events()

		query := &bson.M{
			"_id": &bson.M{
				"$gt": cursorId,
			},
			"channel": channelBson,
		}

		for {
			cursor, err = coll.Find(
				context.Background(),
				query,
				queryOpts,
			)
			if err != nil {
				err = database.ParseError(err)

				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("event: Listener find error")
			} else {
				break
			}

			if !l.state {
				return
			}

			time.Sleep(constants.RetryDelay)

			if !l.state {
				return
			}
		}
	}
}

func (l *Listener) init() (err error) {
	db := database.GetDatabase()

	coll := db.Events()
	cursorId, err := getCursorId(db, coll, l.channels)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	l.state = true

	go func() {
		if r := recover(); r != nil {
			logrus.WithFields(logrus.Fields{
				"error": errors.New(fmt.Sprintf("%s", r)),
			}).Error("event: Listener panic")
		}
		l.sub(db, cursorId)
	}()

	return
}

func SubscribeListener(channels []string) (lst *Listener, err error) {
	lst = &Listener{
		channels: channels,
		stream:   make(chan *Event, 10),
	}

	err = lst.init()
	if err != nil {
		return
	}

	return
}
