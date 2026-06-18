package advisory

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/telemetry"
	"github.com/pritunl/pritunl-cloud/vulnerability"
)

type Advisory struct {
	Id                 bson.ObjectID                  `bson:"_id" json:"id"`
	Organization       bson.ObjectID                  `bson:"organization" json:"organization"`
	Reference          string                         `bson:"reference" json:"reference"`
	Dismissed          bool                           `bson:"dismissed" json:"dismissed"`
	Type               string                         `bson:"type" json:"type"`
	Updated            time.Time                      `bson:"updated" json:"updated"`
	Severity           string                         `bson:"severity" json:"severity"`
	Description        string                         `bson:"description" json:"description"`
	Score              int                            `bson:"score" json:"score"`
	Packages           []string                       `bson:"packages" json:"packages"`
	Vuxmls             []string                       `bson:"vuxmls" json:"vuxmls"`
	Vulnerabilities    []*vulnerability.Vulnerability `bson:"vulnerabilities" json:"vulnerabilities"`
	Instances          []bson.ObjectID                `bson:"instances" json:"instances"`
	Nodes              []bson.ObjectID                `bson:"nodes" json:"nodes"`
	DismissedResources []bson.ObjectID                `bson:"dismissed_resources" json:"dismissed_resources"`
}

func (a *Advisory) scoreAdvisory(vuln *vulnerability.Vulnerability) int {
	if vuln == nil {
		return Low
	}

	isNetwork := vuln.Vector == vulnerability.Network
	isAdjacent := vuln.Vector == vulnerability.Adjacent
	isUnauth := vuln.Privileges == vulnerability.None
	isNoInteraction := vuln.Interaction == vulnerability.None
	isCritical := vuln.Severity == vulnerability.Critical
	isHigh := vuln.Severity == vulnerability.High

	if isNetwork && isUnauth && isNoInteraction &&
		(isCritical || vuln.Score >= 9.0) {

		if a.Severity == moderate {
			return High
		}
		return Critical
	}

	if isNetwork && isUnauth {
		if a.Severity == moderate {
			return Medium
		}
		return High
	}
	if isNetwork && isCritical {
		if a.Severity == moderate {
			return Medium
		}
		return High
	}
	if (isNetwork || isAdjacent) && vuln.Score >= 9.5 {
		if a.Severity == moderate {
			return Medium
		}
		return High
	}

	if isNetwork && (isHigh || vuln.Score >= 7.0) {
		if a.Severity == moderate {
			return Low
		}
		return Medium
	}
	if isAdjacent && isUnauth && (isCritical || isHigh) {
		if a.Severity == moderate {
			return Low
		}
		return Medium
	}
	if isCritical {
		if a.Severity == moderate {
			return Low
		}
		return Medium
	}

	return Low
}

func (a *Advisory) UpdateScore() {
	top := Low
	for _, vuln := range a.Vulnerabilities {
		score := a.scoreAdvisory(vuln)
		if score > top {
			top = score
		}
	}
	a.Score = top
}

func FromUpdate(updt *telemetry.Update, orgId bson.ObjectID, now time.Time,
	vulns []*vulnerability.Vulnerability) *Advisory {

	return &Advisory{
		Organization:    orgId,
		Reference:       updt.Id,
		Type:            RedHat,
		Updated:         now,
		Severity:        updt.Severity,
		Description:     updt.Description,
		Packages:        updt.Packages,
		Vulnerabilities: vulns,
		Instances:       []bson.ObjectID{},
		Nodes:           []bson.ObjectID{},
	}
}
