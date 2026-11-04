package node

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/manifest"
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
		options.Find().SetProjection(&bson.D{
			{"name", 1},
		}),
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
		options.Find().
			SetSort(&bson.D{
				{"name", 1},
			}).
			SetProjection(&bson.D{
				{"name", 1},
				{"types", 1},
				{"gui", 1},
				{"available_vpcs", 1},
				{"cloud_subnets", 1},
				{"default_no_public_address", 1},
				{"default_no_public_address6", 1},
			}),
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
		options.Find().SetProjection(&bson.D{
			{"name", 1},
		}),
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

func GetAllShape(db *database.Database, zones []bson.ObjectID,
	roles []string) (nodes []*Node, err error) {

	coll := db.Nodes()
	nodes = []*Node{}

	cursor, err := coll.Aggregate(db, []*bson.M{
		&bson.M{
			"$match": &bson.M{
				"zone": &bson.M{
					"$in": zones,
				},
				"roles": &bson.M{
					"$in": roles,
				},
			},
		},
		RuntimeLookup(),
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
		doc := &RuntimePipe{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		nde := doc.Merge()
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

	cursor, err := coll.Find(db, bson.M{}, options.Find().SetProjection(&bson.D{
		{"datacenter", 1},
		{"zone", 1},
		{"private_ips", 1},
	}))
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

	err = manifest.RemoveResource(db, nodeId)
	if err != nil {
		return
	}

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
