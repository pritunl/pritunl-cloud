package advisory

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/telemetry"
	"github.com/pritunl/pritunl-cloud/vulnerability"
)

type Advisory struct {
	Id              string                         `bson:"_id" json:"id"`
	Type            string                         `bson:"type" json:"type"`
	Updated         time.Time                      `bson:"updated" json:"updated"`
	Severity        string                         `bson:"severity" json:"severity"`
	Description     string                         `bson:"description" json:"description"`
	Score           int                            `bson:"score" json:"score"`
	Packages        []string                       `bson:"packages" json:"packages"`
	Vulnerabilities []*vulnerability.Vulnerability `bson:"vulnerabilities" json:"vulnerabilities"`
	Instances       []bson.ObjectID                `bson:"instances" json:"instances"`
	Nodes           []bson.ObjectID                `bson:"nodes" json:"nodes"`
}

func FromUpdate(updt *telemetry.Update, score int,
	vulns []*vulnerability.Vulnerability) *Advisory {

	return &Advisory{
		Id:              updt.Id,
		Type:            RedHat,
		Description:     updt.Description,
		Score:           score,
		Packages:        updt.Packages,
		Vulnerabilities: vulns,
		Instances:       []bson.ObjectID{},
		Nodes:           []bson.ObjectID{},
	}
}
