package advisory

import (
	"slices"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/telemetry"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vulnerability"
	"github.com/pritunl/pritunl-cloud/vuxml"
)

type Advisory struct {
	Id              bson.ObjectID `bson:"_id" json:"id"`
	Organization    bson.ObjectID `bson:"organization" json:"organization"`
	Reference       string        `bson:"reference" json:"reference"`
	Dismissed       bool          `bson:"dismissed" json:"dismissed"`
	Type            string        `bson:"type" json:"type"`
	Updated         time.Time     `bson:"updated" json:"updated"`
	Severity        string        `bson:"severity" json:"severity"`
	Description     string        `bson:"description" json:"description"`
	Score           int           `bson:"score" json:"score"`
	Packages        []string      `bson:"packages" json:"packages"`
	Vuxmls          []string      `bson:"vuxmls" json:"vuxmls"`
	Vulnerabilities []string      `bson:"vulnerabilities" json:"vulnerabilities"`
	Pending         int           `bson:"pending" json:"pending"`
	Complete        bool          `bson:"complete" json:"complete"`
	InstanceCount   int           `bson:"instance_count" json:"instance_count"`
	NodeCount       int           `bson:"node_count" json:"node_count"`

	vulns     []*vulnerability.VulnerabilityMini `bson:"-" json:"-"`
	vulnsSet  set.Set                            `bson:"-" json:"-"`
	vuxmlsSet set.Set                            `bson:"-" json:"-"`
	pkgsSet   set.Set                            `bson:"-" json:"-"`
}

func (a *Advisory) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if a.Reference == "" {
		errData = &errortypes.ErrorData{
			Error:   "reference_required",
			Message: "Missing required advisory reference",
		}
		return
	}

	if a.Updated.IsZero() {
		a.Updated = time.Now()
	}

	a.Reference = utils.FilterId(a.Reference)

	if a.Type != "" && !ValidTypes.Contains(a.Type) {
		errData = &errortypes.ErrorData{
			Error:   "invalid_type",
			Message: "Invalid advisory type",
		}
		return
	}

	if a.Severity != "" && !ValidSeverities.Contains(a.Severity) {
		errData = &errortypes.ErrorData{
			Error:   "invalid_severity",
			Message: "Invalid advisory severity",
		}
		return
	}

	if a.Score < 0 || a.Score > Critical {
		errData = &errortypes.ErrorData{
			Error:   "invalid_score",
			Message: "Invalid advisory score",
		}
		return
	}

	a.Description = utils.FilterStrExt(
		a.Description,
		settings.Telemetry.DescriptionLimit,
	)

	if a.Packages == nil {
		a.Packages = []string{}
	}
	if a.Vuxmls == nil {
		a.Vuxmls = []string{}
	}
	if a.Vulnerabilities == nil {
		a.Vulnerabilities = []string{}
	}

	a.Complete = a.Pending == 0

	return
}

func (a *Advisory) scoreVulnerability(
	vuln *vulnerability.VulnerabilityMini) int {

	if vuln == nil {
		return Low
	}

	if vuln.Analysis != nil {
		score := vuln.Analysis.Score
		if score >= 9.0 {
			return Critical
		}
		if score >= 6.0 {
			return High
		}
		if score >= 3.0 {
			return Medium
		}
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

func (a *Advisory) AddVulnerabilities(cveIds []string,
	vulns map[string]*vulnerability.VulnerabilityMini) {

	if a.vulnsSet == nil {
		a.vulnsSet = set.NewSet()
		for _, cveId := range a.Vulnerabilities {
			a.vulnsSet.Add(cveId)
		}
	}

	for _, cveId := range cveIds {
		cveId = strings.ToUpper(cveId)
		if !vulnerability.ValidId(cveId) {
			continue
		}
		if a.vulnsSet.Contains(cveId) {
			continue
		}
		a.vulnsSet.Add(cveId)

		a.Vulnerabilities = append(a.Vulnerabilities, cveId)

		vuln := vulns[cveId]
		if vuln.HasData() {
			a.vulns = append(a.vulns, vuln)
		} else if vuln.IsPending() {
			a.Pending += 1
		}
	}

	a.Complete = a.Pending == 0
}

func (a *Advisory) SetVulnerabilities(
	vulns map[string]*vulnerability.VulnerabilityMini) {

	cveIds := a.Vulnerabilities

	a.Vulnerabilities = []string{}
	a.vulns = nil
	a.vulnsSet = set.NewSet()
	a.Pending = 0

	a.AddVulnerabilities(cveIds, vulns)
}

func (a *Advisory) Reachable(components *vulnerability.Components) bool {
	if a.Pending > 0 || len(a.vulns) == 0 {
		return true
	}

	for _, vuln := range a.vulns {
		if vuln.Analysis.Reachable(components) {
			return true
		}
	}

	return false
}

func (a *Advisory) ResourceState(components *vulnerability.Components) (
	state string, exclusions []string) {

	exclusions = []string{}

	for _, vuln := range a.vulns {
		if !vuln.Analysis.Reachable(components) {
			exclusions = append(exclusions, vuln.Id)
		}
	}

	if a.Reachable(components) {
		state = Affected
	} else {
		state = Unreachable
	}

	return
}

func (a *Advisory) UpdateScore() {
	a.Complete = a.Pending == 0
	top := Low
	for _, vuln := range a.vulns {
		score := a.scoreVulnerability(vuln)
		if score > top {
			top = score
		}
	}
	a.Score = top
}

func (a *Advisory) UpsertModel() mongo.WriteModel {
	return mongo.NewUpdateOneModel().
		SetFilter(&bson.M{
			"organization": a.Organization,
			"reference":    a.Reference,
		}).
		SetUpdate(&bson.M{
			"$set": &bson.M{
				"organization":    a.Organization,
				"reference":       a.Reference,
				"type":            a.Type,
				"updated":         a.Updated,
				"severity":        a.Severity,
				"description":     a.Description,
				"score":           a.Score,
				"packages":        a.Packages,
				"vuxmls":          a.Vuxmls,
				"vulnerabilities": a.Vulnerabilities,
				"pending":         a.Pending,
				"complete":        a.Complete,
				"instance_count":  a.InstanceCount,
				"node_count":      a.NodeCount,
			},
			"$setOnInsert": &bson.M{
				"dismissed": false,
			},
			"$unset": &bson.M{
				"instances":             "",
				"nodes":                 "",
				"unreachable_resources": "",
				"dismissed_resources":   "",
				"exclusion_resources":   "",
			},
		}).
		SetUpsert(true)
}

func (a *Advisory) MergePackages(pkgs []string) {
	if a.pkgsSet == nil {
		a.pkgsSet = set.NewSet()

		curPkgs := a.Packages
		a.Packages = make([]string, 0, len(curPkgs))
		for _, pkg := range curPkgs {
			if a.pkgsSet.Contains(pkg) {
				continue
			}
			a.pkgsSet.Add(pkg)
			a.Packages = append(a.Packages, pkg)
		}
		slices.Sort(a.Packages)
	}

	added := false
	for _, pkg := range pkgs {
		if a.pkgsSet.Contains(pkg) {
			continue
		}
		a.pkgsSet.Add(pkg)
		a.Packages = append(a.Packages, pkg)
		added = true
	}

	if added {
		slices.Sort(a.Packages)
	}
}

func (a *Advisory) MergeVuxml(pkg string, entry *vuxml.VuxmlEntry,
	cveIds []string, vulns map[string]*vulnerability.VulnerabilityMini) {

	a.MergePackages([]string{pkg})

	if a.vuxmlsSet == nil {
		a.vuxmlsSet = set.NewSet()
		for _, vid := range a.Vuxmls {
			a.vuxmlsSet.Add(vid)
		}
	}

	if a.vuxmlsSet.Contains(entry.Vid) {
		return
	}
	a.vuxmlsSet.Add(entry.Vid)
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

	a.AddVulnerabilities(cveIds, vulns)
}

func FromUpdate(updt *telemetry.Update, orgId bson.ObjectID, now time.Time,
	vulns map[string]*vulnerability.VulnerabilityMini) *Advisory {

	adv := &Advisory{
		Organization:    orgId,
		Reference:       updt.Id,
		Type:            RedHat,
		Updated:         now,
		Severity:        updt.Severity,
		Description:     updt.Description,
		Packages:        updt.Packages,
		Vuxmls:          []string{},
		Vulnerabilities: []string{},
		Complete:        true,
	}

	adv.AddVulnerabilities(updt.Vulnerabilities, vulns)

	return adv
}

func NewUpdate(ref string, typ string, orgId bson.ObjectID,
	now time.Time) *Advisory {

	return &Advisory{
		Organization:    orgId,
		Reference:       ref,
		Type:            typ,
		Updated:         now,
		Severity:        "",
		Description:     "",
		Packages:        []string{},
		Vuxmls:          []string{},
		Vulnerabilities: []string{},
		Complete:        true,
	}
}
