package audit

import (
	"net/http"
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/useragent"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Get(db *database.Database, adtId string) (
	adt *Audit, err error) {

	coll := db.Audits()
	adt = &Audit{}

	err = coll.FindOneId(adtId, adt)
	if err != nil {
		return
	}

	return
}

func GetAll(db *database.Database, userId bson.ObjectID,
	page, pageCount int64) (audits []*Audit, count int64, err error) {

	coll := db.Audits()
	audits = []*Audit{}

	count, err = coll.CountDocuments(db, &bson.M{
		"u": userId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	maxPage := count / pageCount
	if count == pageCount {
		maxPage = 0
	}
	page = utils.Min64(page, maxPage)
	skip := utils.Min64(page*pageCount, count)

	cursor, err := coll.Find(db, &bson.M{
		"u": userId,
	}, &options.FindOptions{
		Sort: &bson.D{
			{"$natural", -1},
		},
		Skip:  &skip,
		Limit: &pageCount,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		adt := &Audit{}
		err = cursor.Decode(adt)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		audits = append(audits, adt)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func New(db *database.Database, r *http.Request,
	userId bson.ObjectID, typ string, fields Fields) (
	err error) {

	if settings.System.Demo {
		return
	}

	agnt, err := useragent.Parse(db, r)
	if err != nil {
		return
	}

	adt := &Audit{
		User:      userId,
		Timestamp: time.Now(),
		Type:      typ,
		Fields:    fields,
		Agent:     agnt,
	}

	err = adt.Insert(db)
	if err != nil {
		return
	}

	return
}
