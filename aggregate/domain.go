package aggregate

import (
	"sync"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/domain"
)

type Domain struct {
	domain.Domain `bson:",inline"`
	Records       []*domain.Record `bson:"records" json:"records"`
}

type DomainsPipe struct {
	Metadata []*Metadata `bson:"meta"`
	Domains  []*Domain   `bson:"domains"`
}

func GetDomainPaged(db *database.Database, query *bson.M, page,
	pageCount int64) (domains []*domain.Domain, count int64, err error) {

	coll := db.Domains()
	domains = []*domain.Domain{}

	if pageCount == 0 {
		pageCount = 20
	}
	skip := page * pageCount

	addDomain := func(domn *Domain) {
		domn.Domain.Records = domn.Records
		domn.Json()
		domains = append(domains, &domn.Domain)
	}

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
					"from":         "domains_records",
					"localField":   "_id",
					"foreignField": "domain",
					"as":           "records",
				},
			},
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}
		defer cursor.Close(db)

		for cursor.Next(db) {
			domn := &Domain{}
			err = cursor.Decode(domn)
			if err != nil {
				err = database.ParseError(err)
				return
			}

			addDomain(domn)
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
					"domains": []*bson.M{
						&bson.M{
							"$skip": skip,
						},
						&bson.M{
							"$limit": pageCount,
						},
						&bson.M{
							"$lookup": &bson.M{
								"from":         "domains_records",
								"localField":   "_id",
								"foreignField": "domain",
								"as":           "records",
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

		doc := &DomainsPipe{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		if len(doc.Metadata) > 0 {
			count = doc.Metadata[0].Count
		}

		for _, domn := range doc.Domains {
			addDomain(domn)
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
