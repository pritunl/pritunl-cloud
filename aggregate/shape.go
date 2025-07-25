package aggregate

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/utils"
)

type ShapePipe struct {
	shape.Shape `bson:",inline"`
	NodeDocs    []*node.Node `bson:"node_docs"`
}

func GetShapePaged(db *database.Database, query *bson.M, page,
	pageCount int64) (shapes []*shape.Shape, count int64, err error) {

	coll := db.Shapes()
	shapes = []*shape.Shape{}

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
										"$size": &bson.M{
											"$setIntersection": []interface{}{
												"$roles",
												"$$shape_roles",
											},
										},
									},
									0,
								},
							},
						},
					},
				},
				"as": "node_docs",
			},
		},
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

		shpe := &doc.Shape
		shpe.NodeCount = len(doc.NodeDocs)

		shapes = append(shapes, shpe)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
