package advisory

import (
	"net/http"
	"time"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

type Advisory struct {
	Id              string  `bson:"id" json:"id"`
	Status          string  `bson:"status" json:"status"`
	Description     string  `bson:"description" json:"description"`
	Score           float64 `bson:"score" json:"score"`
	Severity        string  `bson:"severity" json:"severity"`
	Vector          string  `bson:"vector" json:"vector"`
	Complexity      string  `bson:"complexity" json:"complexity"`
	Privileges      string  `bson:"privileges" json:"privileges"`
	Interaction     string  `bson:"interaction" json:"interaction"`
	Scope           string  `bson:"scope" json:"scope"`
	Confidentiality string  `bson:"confidentiality" json:"confidentiality"`
	Integrity       string  `bson:"integrity" json:"integrity"`
	Availability    string  `bson:"availability" json:"availability"`
}
