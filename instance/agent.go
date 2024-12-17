package instance

import (
	"time"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
)

type AgentLog struct {
	Id        primitive.ObjectID `bson:"_id" json:"id"`
	Instance  primitive.ObjectID `bson:"i" json:"i"`
	Timestamp time.Time          `bson:"t" json:"t"`
	Message   []string           `bson:"m" json:"m"`
}
