package aggregate

import (
	"sort"
	"sync"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/unit"
)

func sortUnits(units []*unit.Unit) {
	sort.SliceStable(units, func(i, j int) bool {
		return units[i].Name < units[j].Name
	})
}

type PodPipe struct {
	pod.Pod  `bson:",inline"`
	UnitDocs []*unit.Unit `bson:"unit_docs"`
}

type Metadata struct {
	Count int64 `bson:"count"`
}

type PodsPipe struct {
	Metadata []*Metadata `bson:"meta"`
	Pods     []*PodPipe  `bson:"pods"`
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

	sortUnits(doc.UnitDocs)

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

	if pageCount == 0 {
		pageCount = 20
	}
	skip := page * pageCount

	var cursor *mongo.Cursor
	if len(*query) == 0 {
		waiter := &sync.WaitGroup{}
		var countErr error

		waiter.Add(1)
		go func() {
			defer waiter.Done()

			count, countErr = coll.EstimatedDocumentCount(db)
			if countErr != nil {
				countErr = database.ParseError(countErr)
				return
			}
		}()

		cursor, err = coll.Aggregate(db, []*bson.M{
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

			sortUnits(doc.UnitDocs)

			pd := &PodAggregate{
				Pod:   doc.Pod,
				Units: doc.UnitDocs,
			}

			pd.Json(usrId)

			pods = append(pods, pd)
		}

		waiter.Wait()
		if countErr != nil {
			err = countErr
			return
		}
	} else {
		cursor, err = coll.Aggregate(db, []*bson.M{
			&bson.M{
				"$match": query,
			},
			&bson.M{
				"$sort": &bson.M{
					"name": 1,
				},
			},
			&bson.M{
				"$facet": &bson.M{
					"meta": []*bson.M{
						&bson.M{
							"$count": "count",
						},
					},
					"pods": []*bson.M{
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
					},
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
				errors.New("aggregate: Not found"),
			}
			return
		}

		doc := &PodsPipe{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		if len(doc.Metadata) > 0 {
			count = doc.Metadata[0].Count
		}

		for _, podDoc := range doc.Pods {
			sortUnits(podDoc.UnitDocs)

			pd := &PodAggregate{
				Pod:   podDoc.Pod,
				Units: podDoc.UnitDocs,
			}

			pd.Json(usrId)

			pods = append(pods, pd)
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
