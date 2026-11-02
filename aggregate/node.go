package aggregate

import (
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
)

func GetNode(db *database.Database, nodeId bson.ObjectID) (
	nde *node.Node, err error) {

	coll := db.Nodes()

	cursor, err := coll.Aggregate(db, []*bson.M{
		&bson.M{
			"$match": &bson.M{
				"_id": nodeId,
			},
		},
		node.RuntimeLookup(),
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	if !cursor.Next(db) {
		err = cursor.Err()
		if err != nil {
			err = database.ParseError(err)
			return
		}

		err = &database.NotFoundError{
			errors.New("aggregate: Node not found"),
		}
		return
	}

	doc := &node.RuntimePipe{}
	err = cursor.Decode(doc)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	nde = doc.Merge()
	nde.SetActive()

	return
}

func GetNodesPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (nodes []*node.Node, count int64, err error) {

	coll := db.Nodes()
	nodes = []*node.Node{}

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

	if pageCount == 0 {
		pageCount = 20
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
			"$sort": &bson.D{
				{"name", 1},
			},
		},
		&bson.M{
			"$skip": skip,
		},
		&bson.M{
			"$limit": pageCount,
		},
		node.RuntimeLookup(),
		&bson.M{
			"$project": &bson.D{
				{"runtime_docs.processes", 0},
				{"runtime_docs.modules", 0},
				{"runtime_docs.ports", 0},
			},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		doc := &node.RuntimePipe{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		nde := doc.Merge()
		nde.SetActive()
		nodes = append(nodes, nde)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
