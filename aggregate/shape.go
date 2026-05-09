package aggregate

import (
	"sync"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/shape"
)

type ShapePipe struct {
	shape.Shape `bson:",inline"`
	NodeDocs    []*node.Node `bson:"node_docs"`
}

type ShapesPipe struct {
	Metadata []*Metadata  `bson:"meta"`
	Shapes   []*ShapePipe `bson:"shapes"`
}

func GetShapePaged(db *database.Database, query *bson.M, page,
	pageCount int64) (shapes []*shape.Shape, count int64, err error) {

	coll := db.Shapes()
	shapes = []*shape.Shape{}

	if pageCount == 0 {
		pageCount = 20
	}
	skip := page * pageCount

	sizeQuery := &bson.M{
		"$ifNull": bson.A{
			&bson.M{
				"$setIntersection": bson.A{
					&bson.M{"$ifNull": bson.A{"$roles", bson.A{}}},
					&bson.M{"$ifNull": bson.A{"$$shape_roles", bson.A{}}},
				},
			},
			bson.A{},
		},
	}

	nodeLookup := &bson.M{
		"$lookup": &bson.M{
			"from": "nodes",
			"let": &bson.M{
				"shape_roles": "$roles",
			},
			"pipeline": []*bson.M{
				{
					"$match": &bson.M{
						"$expr": &bson.M{
							"$gt": bson.A{
								&bson.M{
									"$size": sizeQuery,
								},
								0,
							},
						},
					},
				},
			},
			"as": "node_docs",
		},
	}

	addShape := func(doc *ShapePipe) {
		shpe := &doc.Shape
		shpe.NodeCount = len(doc.NodeDocs)
		shapes = append(shapes, shpe)
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
			nodeLookup,
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}
		defer cursor.Close(db)

		for cursor.Next(db) {
			doc := &ShapePipe{}
			err = cursor.Decode(doc)
			if err != nil {
				err = database.ParseError(err)
				return
			}

			addShape(doc)
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
					"shapes": []*bson.M{
						&bson.M{
							"$skip": skip,
						},
						&bson.M{
							"$limit": pageCount,
						},
						nodeLookup,
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

		doc := &ShapesPipe{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		if len(doc.Metadata) > 0 {
			count = doc.Metadata[0].Count
		}

		for _, shapeDoc := range doc.Shapes {
			addShape(shapeDoc)
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
