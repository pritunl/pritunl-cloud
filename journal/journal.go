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
	Kind      int                `bson:"k" json:"k"`
	Timestamp time.Time          `bson:"t" json:"t"`
	Message   string             `bson:"m" json:"m"`
}

func (j *Journal) String() string {
	return fmt.Sprintf(
		"[%s] %s\n",
		j.Timestamp.Format("Mon Jan _2 15:04:05 2006"),
		j.Message,
	)
}

func (j *Journal) Insert(db *database.Database) (err error) {
	coll := db.Journal()

	_, err = coll.InsertOne(db, j)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
