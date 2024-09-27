package database

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo"
	"github.com/pritunl/mongo-go-driver/mongo/options"
)

type Collection struct {
	db *Database
	*mongo.Collection
}

func (c *Collection) FindOneId(id interface{}, data interface{}) (err error) {
	err = c.FindOne(c.db, &bson.M{
		"_id": id,
	}).Decode(data)
	if err != nil {
		err = ParseError(err)
		return
	}

	return
}

func (c *Collection) UpdateId(id interface{}, data interface{}) (err error) {
	_, err = c.UpdateOne(c.db, &bson.M{
		"_id": id,
	}, data)
	if err != nil {
		err = ParseError(err)
		return
	}

	return
}

func (c *Collection) Commit(id interface{}, data interface{}) (err error) {
	_, err = c.UpdateOne(c.db, &bson.M{
		"_id": id,
	}, &bson.M{
		"$set": data,
	})
	if err != nil {
		err = ParseError(err)
		return
	}

	return
}

func (c *Collection) CommitFields(id interface{}, data interface{},
	fields set.Set) (err error) {

	_, err = c.UpdateOne(c.db, &bson.M{
		"_id": id,
	}, SelectFieldsAll(data, fields))
	if err != nil {
		err = ParseError(err)
		return
	}

	return
}

func (c *Collection) Upsert(query *bson.M, data interface{}) (err error) {
	opts := &options.UpdateOptions{}
	opts.SetUpsert(true)

	_, err = c.UpdateOne(c.db, query, &bson.M{
		"$set": data,
	}, opts)
	if err != nil {
		err = ParseError(err)
		return
	}

	return
}

func SelectFields(obj interface{}, fields set.Set) (data bson.M) {
	val := reflect.ValueOf(obj).Elem()
	data = bson.M{}

	n := val.NumField()
	for i := 0; i < n; i++ {
		typ := val.Type().Field(i)

		if typ.PkgPath != "" {
			continue
		}

		tag := typ.Tag.Get("bson")
		if tag == "" || tag == "-" {
			continue
		}
		tag = strings.Split(tag, ",")[0]

		if !fields.Contains(tag) {
			continue
		}

		val := val.Field(i).Interface()

		switch valTyp := val.(type) {
		case primitive.ObjectID:
			if valTyp.IsZero() {
				data[tag] = nil
			} else {
				data[tag] = val
			}
			break
		default:
			data[tag] = val
		}
	}

	return
}

func SelectFieldsAll(obj interface{}, fields set.Set) (data bson.M) {
	val := reflect.ValueOf(obj).Elem()

	dataSet := bson.M{}
	dataUnset := bson.M{}
	dataUnseted := false

	n := val.NumField()
	for i := 0; i < n; i++ {
		typ := val.Type().Field(i)

		if typ.PkgPath != "" {
			continue
		}

		tag := typ.Tag.Get("bson")
		if tag == "" || tag == "-" {
			continue
		}

		omitempty := strings.Contains(tag, "omitempty")

		tag = strings.Split(tag, ",")[0]
		if !fields.Contains(tag) {
			continue
		}

		val := val.Field(i).Interface()

		switch valTyp := val.(type) {
		case primitive.ObjectID:
			if valTyp.IsZero() {
				if omitempty {
					dataUnset[tag] = 1
					dataUnseted = true
				} else {
					dataSet[tag] = nil
				}
			} else {
				dataSet[tag] = val
			}
			break
		default:
			dataSet[tag] = val
		}
	}

	data = bson.M{
		"$set": dataSet,
	}
	if dataUnseted {
		data["$unset"] = dataUnset
	}

	return
}

type ArraySelectFields struct {
	count       int
	setFields   bson.M
	unsetFields bson.M
	filters     []interface{}
	push        []interface{}
	pull        []primitive.ObjectID
	rootKey     string
	idKey       string
}

func (a *ArraySelectFields) Update(docId primitive.ObjectID,
	update bson.M) {

	matchKey := fmt.Sprintf("elem%d", a.count)
	a.count += 1

	setStr := fmt.Sprintf("%s.$[%s].", a.rootKey, matchKey)

	for key, val := range update {
		a.setFields[setStr+key] = val
	}

	a.filters = append(a.filters, bson.M{
		fmt.Sprintf("%s.%s", matchKey, a.idKey): docId,
	})
}

func (a *ArraySelectFields) Push(doc interface{}) {
	a.push = append(a.push, doc)
}

func (a *ArraySelectFields) Delete(docId primitive.ObjectID) {
	a.pull = append(a.pull, docId)
}

func (a *ArraySelectFields) GetQuery() (query bson.M, filters []interface{}) {
	query = bson.M{
		"$set": a.setFields,
	}
	if len(a.unsetFields) > 0 {
		query["$unset"] = a.unsetFields
	}

	filters = a.filters

	if len(a.push) > 0 {
		query["$push"] = bson.M{
			a.rootKey: bson.M{
				"$each": &a.push,
			},
		}
	}

	if len(a.pull) > 0 {
		query["$pull"] = bson.M{
			a.rootKey: bson.M{
				a.idKey: bson.M{
					"$in": a.pull,
				},
			},
		}
	}

	return
}

func NewArraySelectFields(obj interface{}, rootKey string, fields set.Set) (
	arraySel *ArraySelectFields) {

	selectFields := SelectFieldsAll(obj, fields)
	setFields := selectFields["$set"].(bson.M)

	var unsetFields bson.M
	if _, exists := selectFields["$unset"]; exists {
		unsetFields = selectFields["$unset"].(bson.M)
	}

	arraySel = &ArraySelectFields{
		count:       1,
		setFields:   setFields,
		unsetFields: unsetFields,
		filters:     []interface{}{},
		push:        []interface{}{},
		pull:        []primitive.ObjectID{},
		rootKey:     rootKey,
		idKey:       "id",
	}

	return
}
