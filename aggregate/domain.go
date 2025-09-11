package aggregate

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Domain struct {
	domain.Domain `bson:",inline"`
	Records       []*domain.Record `bson:"records" json:"records"`
}

func GetDomainPaged(db *database.Database, query *bson.M, page,
	pageCount int64) (domains []*Domain, count int64, err error) {

	coll := db.Domains()
	domains = []*Domain{}

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

		domains = append(domains, domn)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
