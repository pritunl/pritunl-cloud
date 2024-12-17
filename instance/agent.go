package instance

import (
	"time"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
)

type AgentLog struct {
	Id        primitive.ObjectID `bson:"_id" json:"id"`
	Instance  primitive.ObjectID `bson:"i" json:"i"`
	Timestamp time.Time          `bson:"t" json:"t"`
	Message   string             `bson:"m" json:"m"`
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
