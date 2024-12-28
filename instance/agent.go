package instance

import (
	"context"
	"fmt"
	"time"

	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/settings"
)

type AgentLog struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Instance  primitive.ObjectID `bson:"i" json:"i"`
	Timestamp time.Time          `bson:"t" json:"t"`
	Message   string             `bson:"m" json:"m"`
}

func (l *AgentLog) String() string {
	return fmt.Sprintf(
		"[%s]%s\n",
		l.Timestamp.Format("Mon Jan _2 15:04:05 2006"),
		l.Message,
	)
}

func (l *AgentLog) Insert(db *database.Database) (err error) {
	coll := db.InstancesAgent()

	_, err = coll.InsertOne(db, l)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAgentLog(c context.Context, db *database.Database,
	instId primitive.ObjectID) (output []string, err error) {

	coll := db.InstancesAgent()

	limit := int64(settings.Hypervisor.ImdsLogDisplayLimit)

	cursor, err := coll.Find(
		c,
		&bson.M{
			"i": instId,
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
		doc := &AgentLog{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		outputRevrse = append(outputRevrse, doc.String())
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
