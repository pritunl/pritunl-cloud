package aggregate

import (
	"sync"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/database"
)

type BlockPipe struct {
	block.Block `bson:",inline"`
	IpCount     int64 `bson:"ip_count"`
}

type BlocksPipe struct {
	Metadata []*Metadata  `bson:"meta"`
	Blocks   []*BlockPipe `bson:"blocks"`
}

type BlockAggregate struct {
	block.Block
	Available int64 `json:"available"`
	Capacity  int64 `json:"capacity"`
}

func GetBlockPaged(db *database.Database, query *bson.M, page,
	pageCount int64) (blocks []*BlockAggregate, count int64, err error) {

	coll := db.Blocks()
	blocks = []*BlockAggregate{}

	if pageCount == 0 {
		pageCount = 20
	}
	skip := page * pageCount

	addBlock := func(doc *BlockPipe) error {
		total, e := doc.GetIpCount()
		if e != nil {
			return e
		}

		blck := &BlockAggregate{
			Block:     doc.Block,
			Available: total - doc.IpCount,
			Capacity:  total,
		}

		blocks = append(blocks, blck)
		return nil
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
					"from":         "blocks_ip",
					"localField":   "_id",
					"foreignField": "block",
					"as":           "ips",
				},
			},
			&bson.M{
				"$addFields": &bson.M{
					"ip_count": &bson.M{
						"$size": "$ips",
					},
				},
			},
			&bson.M{
				"$project": &bson.M{
					"ips": 0,
				},
			},
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}
		defer cursor.Close(db)

		for cursor.Next(db) {
			doc := &BlockPipe{}
			err = cursor.Decode(doc)
			if err != nil {
				err = database.ParseError(err)
				return
			}

			err = addBlock(doc)
			if err != nil {
				return
			}
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
					"blocks": []*bson.M{
						&bson.M{
							"$skip": skip,
						},
						&bson.M{
							"$limit": pageCount,
						},
						&bson.M{
							"$lookup": &bson.M{
								"from":         "blocks_ip",
								"localField":   "_id",
								"foreignField": "block",
								"as":           "ips",
							},
						},
						&bson.M{
							"$addFields": &bson.M{
								"ip_count": &bson.M{
									"$size": "$ips",
								},
							},
						},
						&bson.M{
							"$project": &bson.M{
								"ips": 0,
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

		doc := &BlocksPipe{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		if len(doc.Metadata) > 0 {
			count = doc.Metadata[0].Count
		}

		for _, blockDoc := range doc.Blocks {
			err = addBlock(blockDoc)
			if err != nil {
				return
			}
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
