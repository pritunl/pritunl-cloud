package instance

import (
	"fmt"
	"time"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
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
