package statec

import (
	"strings"

	"github.com/pritunl/mongo-go-driver/v2/bson"
)

const (
	opInsert     = "insert"
	opUpdate     = "update"
	opReplace    = "replace"
	opDelete     = "delete"
	opDrop       = "drop"
	opRename     = "rename"
	opDropDb     = "dropDatabase"
	opInvalidate = "invalidate"
)

type EventNs struct {
	Db   string `bson:"db"`
	Coll string `bson:"coll"`
}

type Event struct {
	Id              bson.Raw       `bson:"_id"`
	OperationType   string         `bson:"operationType"`
	Ns              EventNs        `bson:"ns"`
	DocumentKey     bson.Raw       `bson:"documentKey"`
	Doc             bson.Raw       `bson:"doc"`
	UpdatedKeys     []string       `bson:"updated_keys"`
	RemovedFields   []string       `bson:"removed_fields"`
	TruncatedArrays []string       `bson:"truncated_arrays"`
	ClusterTime     bson.Timestamp `bson:"clusterTime"`
}

func (e *Event) IsStructural() bool {
	switch e.OperationType {
	case opDrop, opRename, opDropDb, opInvalidate:
		return true
	}
	return false
}

func (e *Event) DocId() (id bson.ObjectID, ok bool) {
	if len(e.DocumentKey) == 0 {
		return
	}

	val, err := e.DocumentKey.LookupErr("_id")
	if err != nil {
		return
	}

	id, ok = val.ObjectIDOK()
	return
}

func (e *Event) DocObjectId(key string) (id bson.ObjectID, ok bool) {
	if len(e.Doc) == 0 {
		return
	}

	val, err := e.Doc.LookupErr(key)
	if err != nil {
		return
	}

	id, ok = val.ObjectIDOK()
	return
}

func (e *Event) DocString(key string) (str string, ok bool) {
	if len(e.Doc) == 0 {
		return
	}

	val, err := e.Doc.LookupErr(key)
	if err != nil {
		return
	}

	str, ok = val.StringValueOK()
	return
}

func (e *Event) ChangedKeys() (keys []string) {
	keys = make([]string, 0, len(e.UpdatedKeys)+
		len(e.RemovedFields)+len(e.TruncatedArrays))
	keys = append(keys, e.UpdatedKeys...)
	keys = append(keys, e.RemovedFields...)
	keys = append(keys, e.TruncatedArrays...)
	return
}

func (e *Event) Touches(keys ...string) bool {
	for _, key := range keys {
		prefix := key + "."

		for _, updated := range e.UpdatedKeys {
			if updated == key || strings.HasPrefix(updated, prefix) {
				return true
			}
		}

		for _, removed := range e.RemovedFields {
			if removed == key || strings.HasPrefix(removed, prefix) {
				return true
			}
		}

		for _, truncated := range e.TruncatedArrays {
			if truncated == key || strings.HasPrefix(truncated, prefix) {
				return true
			}
		}
	}

	return false
}
