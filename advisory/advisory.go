package advisory

import (
	"slices"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/telemetry"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vulnerability"
	"github.com/pritunl/pritunl-cloud/vuxml"
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

func (a *Advisory) MergePackages(pkgs []string) {
	merged := slices.Concat(a.Packages, pkgs)
	slices.Sort(merged)
	a.Packages = slices.Compact(merged)
}

func (a *Advisory) MergeVuxml(pkg string, entry *vuxml.VuxmlEntry,
	vulns []*vulnerability.Vulnerability) {

	a.MergePackages([]string{pkg})

	vuxmlsSet := set.NewSet()
	for _, vid := range a.Vuxmls {
		vuxmlsSet.Add(vid)
	}

	if vuxmlsSet.Contains(entry.Vid) {
		return
	}
	a.Vuxmls = append(a.Vuxmls, entry.Vid)
	slices.Sort(a.Vuxmls)

	if len(entry.Paragraphs) > 0 {
		desc := strings.Join(entry.Paragraphs, "\n")
		if a.Description == "" {
			a.Description = desc
		} else {
			a.Description = a.Description + "\n\n" + desc
		}
		a.Description = utils.FilterStrExt(
			a.Description, settings.Telemetry.DescriptionLimit)
	}

	vulnsSet := set.NewSet()
	for _, vuln := range a.Vulnerabilities {
		vulnsSet.Add(vuln.Id)
	}

	for _, vuln := range vulns {
		if vuln == nil || vulnsSet.Contains(vuln.Id) {
			continue
		}
		vulnsSet.Add(vuln.Id)

		a.Vulnerabilities = append(a.Vulnerabilities, vuln)
	}
}

func FromUpdate(updt *telemetry.Update, orgId bson.ObjectID, now time.Time,
	vulns []*vulnerability.Vulnerability) *Advisory {

	return &Advisory{
		Organization:       orgId,
		Reference:          updt.Id,
		Type:               RedHat,
		Updated:            now,
		Severity:           updt.Severity,
		Description:        updt.Description,
		Packages:           updt.Packages,
		Vulnerabilities:    vulns,
		Instances:          []bson.ObjectID{},
		Nodes:              []bson.ObjectID{},
		DismissedResources: []bson.ObjectID{},
	}
}

func NewUpdate(ref string, typ string, orgId bson.ObjectID,
	now time.Time) *Advisory {

	return &Advisory{
		Organization:       orgId,
		Reference:          ref,
		Type:               typ,
		Updated:            now,
		Severity:           "",
		Description:        "",
		Packages:           []string{},
		Vulnerabilities:    []*vulnerability.Vulnerability{},
		Instances:          []bson.ObjectID{},
		Nodes:              []bson.ObjectID{},
		DismissedResources: []bson.ObjectID{},
	}
}
