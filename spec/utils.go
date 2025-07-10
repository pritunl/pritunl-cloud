package spec

import (
	"regexp"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	yamlBlockRe = regexp.MustCompile("(?s)^```yaml\\n(.*?)```")
)

func filterSpecHash(input string) string {
	return yamlBlockRe.ReplaceAllStringFunc(input, func(block string) string {
		lines := strings.Split(block, "\n")
		result := []string{}

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "name:") ||
				strings.HasPrefix(line, "count:") {

				continue
			}
			result = append(result, line)
		}

		return strings.Join(result, "\n")
	})
}

func New(podId, unitId, orgId primitive.ObjectID, data string) (spc *Spec) {
	spc = &Spec{
		Unit:         unitId,
		Pod:          podId,
		Organization: orgId,
		Timestamp:    time.Now(),
		Data:         data,
	}

	return
}

func Get(db *database.Database, commitId primitive.ObjectID) (
	spc *Spec, err error) {

	coll := db.Specs()
	spc = &Spec{}

	err = coll.FindOneId(commitId, spc)
	if err != nil {
		return
	}

	return
}

func GetOne(db *database.Database, query *bson.M) (
	spc *Spec, err error) {

	coll := db.Specs()
	spc = &Spec{}

	err = coll.FindOne(db, query).Decode(spc)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (spcs []*Spec, count int64, err error) {

	coll := db.Specs()
	spcs = []*Spec{}

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
			Projection: &bson.M{
				"_id":       1,
				"unit":      1,
				"index":     1,
				"timestamp": 1,
			},
			Sort: &bson.D{
				{"timestamp", -1},
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
		spc := &Spec{}
		err = cursor.Decode(spc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		spcs = append(spcs, spc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	spcs []*Spec, err error) {

	coll := db.Specs()
	spcs = []*Spec{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		spc := &Spec{}
		err = cursor.Decode(spc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		spcs = append(spcs, spc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllIndexes(db *database.Database, query *bson.M) (
	spcs []*Spec, err error) {

	coll := db.Specs()
	spcs = []*Spec{}

	cursor, err := coll.Find(
		db,
		query,
		&options.FindOptions{
			Projection: &bson.M{
				"_id":       1,
				"unit":      1,
				"index":     1,
				"timestamp": 1,
			},
			Sort: &bson.D{
				{"timestamp", 1},
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		spc := &Spec{}
		err = cursor.Decode(spc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		spcs = append(spcs, spc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllProjectSorted(db *database.Database, query *bson.M) (
	spcs []*Spec, err error) {

	coll := db.Specs()
	spcs = []*Spec{}

	cursor, err := coll.Find(
		db,
		query,
		&options.FindOptions{
			Projection: &bson.M{
				"_id":       1,
				"unit":      1,
				"index":     1,
				"timestamp": 1,
				"hash":      1,
				"data":      1,
			},
			Sort: &bson.D{
				{"timestamp", -1},
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		spc := &Spec{}
		err = cursor.Decode(spc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		spcs = append(spcs, spc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllIds(db *database.Database) (specIds set.Set, err error) {
	coll := db.Specs()
	specIds = set.NewSet()

	cursor, err := coll.Find(
		db,
		bson.M{},
		&options.FindOptions{
			Projection: bson.M{
				"_id": 1,
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		spc := &Spec{}
		err = cursor.Decode(spc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		specIds.Add(spc.Id)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, commitId primitive.ObjectID) (err error) {
	coll := db.Specs()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": commitId,
	})
	if err != nil {
		err = database.ParseError(err)
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		} else {
			return
		}
	}

	return
}

func RemoveAll(db *database.Database, query *bson.M) (err error) {
	coll := db.Specs()

	_, err = coll.DeleteMany(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
