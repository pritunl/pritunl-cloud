package node

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Get(db *database.Database, nodeId bson.ObjectID) (
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

	cursor, err := coll.Find(db, bson.M{})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		nde := &Node{}
		err = cursor.Decode(nde)
		if err != nil {
			err = database.ParseError(err)
			return
		}

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

func GetOne(db *database.Database, query *bson.M) (nde *Node, err error) {
	coll := db.Nodes()
	nde = &Node{}

	err = coll.FindOne(db, query).Decode(nde)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllNamesMap(db *database.Database, query *bson.M) (
	nodeNames map[bson.ObjectID]string, err error) {

	coll := db.Nodes()
	nodeNames = map[bson.ObjectID]string{}

	cursor, err := coll.Find(
		db,
		query,
		&options.FindOptions{
			Projection: &bson.D{
				{"name", 1},
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		nde := &Node{}
		err = cursor.Decode(nde)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		nodeNames[nde.Id] = nde.Name
	}

	err = cursor.Err()
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

	cursor, err := coll.Find(
		db,
		query,
		&options.FindOptions{
			Sort: &bson.D{
				{"name", 1},
			},
			Projection: &bson.D{
				{"name", 1},
				{"types", 1},
				{"gui", 1},
				{"pools", 1},
				{"available_vpcs", 1},
				{"cloud_subnets", 1},
				{"default_no_public_address", 1},
				{"default_no_public_address6", 1},
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		nde := &Node{}
		err = cursor.Decode(nde)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		if !nde.IsHypervisor() {
			continue
		}
		nde.JsonHypervisor()

		nodes = append(nodes, nde)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPool(db *database.Database, poolId bson.ObjectID) (
	nodes []*Node, err error) {

	coll := db.Nodes()
	nodes = []*Node{}

	cursor, err := coll.Find(
		db,
		&bson.M{
			"pools": poolId,
		},
		&options.FindOptions{
			Projection: &bson.D{
				{"name", 1},
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		nde := &Node{}
		err = cursor.Decode(nde)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		nodes = append(nodes, nde)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (nodes []*Node, count int64, err error) {

	coll := db.Nodes()
	nodes = []*Node{}

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

	cursor, err := coll.Find(
		db,
		query,
		&options.FindOptions{
			Sort: &bson.D{
				{"name", 1},
			},
			Skip:  &skip,
			Limit: &pageCount,
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		nde := &Node{}
		err = cursor.Decode(nde)
		if err != nil {
			err = database.ParseError(err)
			return
		}

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

func GetAllShape(db *database.Database, zones []bson.ObjectID,
	roles []string) (nodes []*Node, err error) {

	coll := db.Nodes()
	nodes = []*Node{}

	query := &bson.M{
		"zone": &bson.M{
			"$in": zones,
		},
		"roles": &bson.M{
			"$in": roles,
		},
	}

	cursor, err := coll.Find(
		db,
		query,
		&options.FindOptions{
			//Sort: &bson.D{
			//	{"name", 1},
			//},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		nde := &Node{}
		err = cursor.Decode(nde)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		if !nde.IsHypervisor() || !nde.IsOnline() {
			continue
		}

		nodes = append(nodes, nde)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllNet(db *database.Database) (nodes []*Node, err error) {
	coll := db.Nodes()
	nodes = []*Node{}

	opts := &options.FindOptions{
		Projection: &bson.D{
			{"datacenter", 1},
			{"zone", 1},
			{"private_ips", 1},
		},
	}

	cursor, err := coll.Find(db, bson.M{}, opts)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		nde := &Node{}
		err = cursor.Decode(nde)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		nodes = append(nodes, nde)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, nodeId bson.ObjectID) (err error) {
	coll := db.Nodes()

	_, err = coll.DeleteOne(db, &bson.M{
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
