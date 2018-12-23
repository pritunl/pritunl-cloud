package node

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
)

func Get(db *database.Database, nodeId bson.ObjectId) (
	nde *Node, err error) {

	coll := db.Nodes()
	nde = &Node{}

	err = coll.FindOneId(nodeId, nde)
	if err != nil {
		return
	}

	return
}

func GetAll(db *database.Database) (nodes []*Node, err error) {
	coll := db.Nodes()
	nodes = []*Node{}

	cursor := coll.Find(bson.M{}).Iter()

	nde := &Node{}
	for cursor.Next(nde) {
		nde.SetActive()
		nodes = append(nodes, nde)
		nde = &Node{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllHypervisors(db *database.Database, query *bson.M) (
	nodes []*Node, err error) {

	coll := db.Nodes()
	nodes = []*Node{}

	cursor := coll.Find(query).Sort("name").Select(&bson.M{
		"name":  1,
		"types": 1,
	}).Iter()

	nde := &Node{}
	for cursor.Next(nde) {
		if !nde.IsHypervisor() {
			nde = &Node{}
		} else {
			nodes = append(nodes, nde)
			nde = &Node{}
		}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M, page, pageCount int) (
	nodes []*Node, count int, err error) {

	coll := db.Nodes()
	nodes = []*Node{}

	qury := coll.Find(query)

	count, err = qury.Count()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if page*pageCount == count && page > 0 {
		page -= 1
	}

	skip := utils.Min(page*pageCount, count)

	cursor := qury.Sort("name").Skip(skip).Limit(pageCount).Iter()

	nde := &Node{}
	for cursor.Next(nde) {
		nde.SetActive()
		nodes = append(nodes, nde)
		nde = &Node{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, nodeId bson.ObjectId) (err error) {
	coll := db.Nodes()

	err = coll.Remove(&bson.M{
		"_id": nodeId,
	})
	if err != nil {
		err = database.ParseError(err)
		switch err.(type) {
		case *database.NotFoundError:
			err = nil
		default:
			return
		}
	}

	return
}
