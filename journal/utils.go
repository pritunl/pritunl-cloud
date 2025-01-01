package journal

import (
	"context"

	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/settings"
)

func GetOutput(c context.Context, db *database.Database,
	resource primitive.ObjectID, kind int) (output []string, err error) {

	coll := db.Journal()

	limit := int64(settings.Hypervisor.JournalDisplayLimit)

	cursor, err := coll.Find(
		c,
		&bson.M{
			"r": resource,
			"k": kind,
		},
		&options.FindOptions{
			Limit: &limit,
			Sort: &bson.D{
				{"t", -1},
				{"_id", -1},
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(c)

	outputRevrse := []string{}
	for cursor.Next(c) {
		jrnl := &Journal{}
		err = cursor.Decode(jrnl)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		outputRevrse = append(outputRevrse, jrnl.String())
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	for i := len(outputRevrse) - 1; i >= 0; i-- {
		output = append(output, outputRevrse[i])
	}

	return
}

func Remove(db *database.Database, resource primitive.ObjectID,
	kind int) (err error) {

	coll := db.Journal()

	_, err = coll.DeleteMany(db, &bson.M{
		"r": resource,
		"k": kind,
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
