package aggregate

import (
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/unit"
	"github.com/pritunl/pritunl-cloud/utils"
)

type PodPipe struct {
	pod.Pod  `bson:",inline"`
	UnitDocs []*unit.Unit `bson:"unit_docs"`
}

type PodAggregate struct {
	pod.Pod
	Units []*unit.Unit `json:"units"`
}

func GetPod(db *database.Database, usrId bson.ObjectID, query *bson.M) (
	pd *PodAggregate, err error) {

	coll := db.Pods()

	cursor, err := coll.Aggregate(db, []*bson.M{
		&bson.M{
			"$match": query,
		},
		&bson.M{
			"$lookup": &bson.M{
				"from":         "units",
				"localField":   "_id",
				"foreignField": "pod",
				"as":           "unit_docs",
			},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	if !cursor.Next(db) {
		err = &database.NotFoundError{
			errors.New("aggregate: Pod not found"),
		}
		return
	}

	doc := &PodPipe{}
	err = cursor.Decode(doc)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	pd = &PodAggregate{
		Pod:   doc.Pod,
		Units: doc.UnitDocs,
	}

	pd.Json(usrId)

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetPodsPaged(db *database.Database, usrId bson.ObjectID,
	query *bson.M, page, pageCount int64) (pods []*PodAggregate,
	count int64, err error) {

	coll := db.Pods()
	pods = []*PodAggregate{}

	if len(*query) == 0 {
		count, err = coll.EstimatedDocumentCount(db)
		if err != nil {
			err = database.ParseError(err)
			return
		}
	} else {
		count, err = coll.CountDocuments(db, query)
		if err != nil {
			err = database.ParseError(err)
			return
		}
	}

	maxPage := count / pageCount
	if count == pageCount {
		maxPage = 0
	}
	page = utils.Min64(page, maxPage)
	skip := utils.Min64(page*pageCount, count)

	cursor, err := coll.Aggregate(db, []*bson.M{
		&bson.M{
			"$match": query,
		},
		&bson.M{
			"$sort": &bson.M{
				"name": 1,
			},
		},
		&bson.M{
			"$skip": skip,
		},
		&bson.M{
			"$limit": pageCount,
		},
		&bson.M{
			"$lookup": &bson.M{
				"from":         "units",
				"localField":   "_id",
				"foreignField": "pod",
				"as":           "unit_docs",
			},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		doc := &PodPipe{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		pd := &PodAggregate{
			Pod:   doc.Pod,
			Units: doc.UnitDocs,
		}

		pd.Json(usrId)

		pods = append(pods, pd)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
