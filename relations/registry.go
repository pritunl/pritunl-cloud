package relations

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
)

var registry = map[string]Query{}

func Register(kind string, definition Query) {
	registry[kind] = definition
}

func Aggregate(db *database.Database, kind string, id primitive.ObjectID) (
	resp *Response, err error) {

	definition, ok := registry[kind]
	if !ok {
		return
	}

	definition.Id = id

	resp, err = definition.Aggregate(db)
	if err != nil {
		return
	}

	return
}
