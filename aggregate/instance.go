package aggregate

import (
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/utils"
)

type instanceRuntimePipe struct {
	RuntimeDocs []*instance.Runtime `bson:"runtime_docs"`
}

func getInstance(db *database.Database, query *bson.M) (
	inst *instance.Instance, err error) {

	coll := db.Instances()

	cursor, err := coll.Aggregate(db, []*bson.M{
		&bson.M{
			"$match": query,
		},
		instance.RuntimeLookup(),
		instance.InfoLookup(),
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
			errors.New("aggregate: Instance not found"),
		}
		return
	}

	doc := &instance.Pipe{}
	err = cursor.Decode(doc)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	inst = doc.Merge()

	return
}

func GetInstance(db *database.Database, instId bson.ObjectID) (
	inst *instance.Instance, err error) {

	inst, err = getInstance(db, &bson.M{
		"_id": instId,
	})
	if err != nil {
		return
	}

	return
}

func GetInstanceOrg(db *database.Database, orgId, instId bson.ObjectID) (
	inst *instance.Instance, err error) {

	inst, err = getInstance(db, &bson.M{
		"_id":          instId,
		"organization": orgId,
	})
	if err != nil {
		return
	}

	return
}

func getInstanceRuntime(db *database.Database, query *bson.M) (
	rntm *instance.Runtime, err error) {

	coll := db.Instances()

	cursor, err := coll.Aggregate(db, []*bson.M{
		&bson.M{
			"$match": query,
		},
		&bson.M{
			"$project": &bson.D{
				{"_id", 1},
			},
		},
		instance.RuntimeLookup(),
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
			errors.New("aggregate: Instance not found"),
		}
		return
	}

	doc := &instanceRuntimePipe{}
	err = cursor.Decode(doc)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if len(doc.RuntimeDocs) > 0 {
		rntm = doc.RuntimeDocs[0]
	} else {
		rntm = &instance.Runtime{}
	}

	return
}

func GetInstanceRuntime(db *database.Database, instId bson.ObjectID) (
	rntm *instance.Runtime, err error) {

	rntm, err = getInstanceRuntime(db, &bson.M{
		"_id": instId,
	})
	if err != nil {
		return
	}

	return
}

func GetInstanceRuntimeOrg(db *database.Database,
	orgId, instId bson.ObjectID) (rntm *instance.Runtime, err error) {

	rntm, err = getInstanceRuntime(db, &bson.M{
		"_id":          instId,
		"organization": orgId,
	})
	if err != nil {
		return
	}

	return
}

func GetInstancesPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (insts []*instance.Instance, count int64,
	err error) {

	coll := db.Instances()
	insts = []*instance.Instance{}

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
		instance.RuntimeLookup(),
		instance.InfoLookup(),
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
		doc := &instance.Pipe{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		insts = append(insts, doc.Merge())
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
