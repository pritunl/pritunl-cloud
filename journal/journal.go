package journal

import (
	"fmt"
	"time"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
)

type Journal struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Resource  primitive.ObjectID `bson:"r" json:"r"`
	Kind      int32              `bson:"k" json:"k"`
	Level     int32              `bson:"l" json:"l"`
	Timestamp time.Time          `bson:"t" json:"t"`
	Message   string             `bson:"m" json:"m"`
	Fields    map[string]string  `bson:"f,omitempty" json:"f"`
}

func (j *Journal) String() string {
	return fmt.Sprintf(
		"[%s] %s\n",
		j.Timestamp.Format("2006-01-02 15:04:05"),
		j.Message,
	)
}

func (j *Journal) Insert(db *database.Database) (err error) {
	coll := db.Journal()

	if j.Level == 0 {
		j.Level = Info
	}

	_, err = coll.InsertOne(db, j)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
