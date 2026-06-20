package telemetry

import (
	"bytes"
	"encoding/json"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/tools/commander"
	"github.com/sirupsen/logrus"
)

var (
	cveReg = regexp.MustCompile(`CVE-\d{4}-\d+`)
	vidReg = regexp.MustCompile(
		`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-` +
			`[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
)

type pkgAuditIssue struct {
	Cve []string `json:"cve"`
	Url string   `json:"url"`
}

type pkgAuditPackage struct {
	Version string          `json:"version"`
	Issues  []pkgAuditIssue `json:"issues"`
}

type pkgAuditReport struct {
	Packages map[string]pkgAuditPackage `json:"packages"`
}

var Updates = &Telemetry[[]*Update]{
	TransmitRate: 6 * time.Minute,
	RefreshRate:  6 * time.Hour,
	Refresher:    UpdatesRefresh,
	Validate: func(data []*Update) []*Update {
		if len(data) > 50 {
			return data[:50]
		}
		return data
	},
}

type Update struct {
	Id              string   `bson:"id" json:"id"`
	Type            string   `bson:"type" json:"type"`
	Vulnerabilities []string `bson:"vulnerabilities" json:"vulnerabilities"`
	Severity        string   `bson:"severity" json:"severity"`
	Description     string   `bson:"description" json:"description"`
	Packages        []string `bson:"packages" json:"packages"`
	Score           int      `bson:"-" json:"score"`
}

func (u *Update) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	u.Id = utils.FilterId(u.Id)

	if u.Type != RedHat && u.Type != FreeBsd {
		errData = &errortypes.ErrorData{
			Error:   "invalid_type",
			Message: "Invalid update type",
		}
		return
	}

	for i, cve := range u.Vulnerabilities {
		u.Vulnerabilities[i] = utils.FilterId(cve)
	}

	u.Severity = utils.FilterStr(u.Severity, 64)

	u.Description = utils.FilterStrExt(
		u.Description,
		settings.Telemetry.DescriptionLimit,
	)

	for i, pkg := range u.Packages {
		u.Packages[i] = utils.FilterStr(pkg, 128)
	}

	return
}

func parseRecord(lines []string) (update *Update) {
	updt := &Update{
		Type: RedHat,
	}
	descLines := []string{}
	currentField := ""

	for _, line := range lines {
		colonIdx := strings.Index(line, ":")
		if colonIdx < 0 {
			continue
		}

		prefix := line[:colonIdx]
		value := strings.TrimSpace(line[colonIdx+1:])

		if strings.TrimSpace(prefix) == "" {
			switch currentField {
			case "Description":
				descLines = append(descLines, value)
			case "CVEs":
				if value != "" {
					updt.Vulnerabilities = append(updt.Vulnerabilities,
						cveReg.FindAllString(value, -1)...)
				}
			}
			continue
		}

		field := strings.TrimSpace(prefix)
		currentField = field

		switch field {
		case "Update ID", "Name":
			updt.Id = value
		case "Severity":
			updt.Severity = parseSeverity(value)
		case "Description":
			if value != "" {
				descLines = append(descLines, value)
			}
		case "CVEs":
			if value != "" {
				updt.Vulnerabilities = append(updt.Vulnerabilities,
					cveReg.FindAllString(value, -1)...)
			}
		}
	}

	if !matchAdvisory(updt.Id) {
		return
	}
	if updt.Severity == "" {
		return
	}

	for len(descLines) > 0 && descLines[len(descLines)-1] == "" {
		descLines = descLines[:len(descLines)-1]
	}
	updt.Description = strings.Join(descLines, "\n")

	fullText := strings.Join(lines, "\n")
	cveSet := map[string]bool{}
	deduped := []string{}
	for _, cve := range updt.Vulnerabilities {
		if !cveSet[cve] {
			cveSet[cve] = true
			deduped = append(deduped, cve)
		}
	}
	for _, cve := range cveReg.FindAllString(fullText, -1) {
		if !cveSet[cve] {
			cveSet[cve] = true
			deduped = append(deduped, cve)
		}
	}
	sort.Strings(deduped)
	updt.Vulnerabilities = deduped

	update = updt
	return
}

func updatesList() (advisories map[string][]string, err error) {
	if !IsDnf() {
		return
	}

	resp, err := commander.Exec(&commander.Opt{
		Name: "dnf",
		Args: []string{
			"updateinfo",
			"list",
		},
		Timeout: 90 * time.Second,
		PipeOut: true,
		PipeErr: true,
	})
	if err != nil {
		if resp != nil {
			logrus.WithFields(
				resp.Map(),
			).Error("telemetry: Failed to get dnf security update list")
		}
		return
	}

	advisories = map[string][]string{}
	seen := map[string]map[string]bool{}

	for _, line := range strings.Split(string(resp.Output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Last metadata") ||
			strings.HasPrefix(line, "Updating") ||
			strings.HasPrefix(line, "Repositories") {

			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		adv := parts[0]
		if !matchAdvisory(adv) {
			continue
		}

		pkg := ""
		part1 := strings.ToLower(parts[1])
		part2 := strings.ToLower(parts[2])

		if strings.Contains(part1, moderate) ||
			strings.Contains(part1, important) ||
			strings.Contains(part1, critical) {

			pkg = parts[2]
		} else if len(parts) >= 4 && (strings.Contains(part2, moderate) ||
			strings.Contains(part2, important) ||
			strings.Contains(part2, critical)) {

			pkg = parts[3]
		} else {
			continue
		}

		if pkg == "" {
			continue
		}

		pkgSet, ok := seen[adv]
		if !ok {
			pkgSet = map[string]bool{}
			seen[adv] = pkgSet
		}
		if !pkgSet[pkg] {
			pkgSet[pkg] = true
			advisories[adv] = append(advisories[adv], pkg)
		}
	}

	for adv := range advisories {
		sort.Strings(advisories[adv])
	}

	return
}

func parsePkgAudit(output []byte) (updates []*Update, err error) {
	idx := bytes.IndexByte(output, '{')
	if idx < 0 {
		return
	}

	report := &pkgAuditReport{}
	err = json.Unmarshal(output[idx:], report)
	if err != nil {
		return
	}

	updates = []*Update{}
	seen := map[string]bool{}

	names := make([]string, 0, len(report.Packages))
	for name := range report.Packages {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		pkg := report.Packages[name]
		pkgRef := name + "@" + pkg.Version

		for _, issue := range pkg.Issues {
			vid := vidReg.FindString(issue.Url)
			if vid == "" {
				continue
			}

			key := name + "|" + vid
			if seen[key] {
				continue
			}
			seen[key] = true

			cves := []string{}
			cveSeen := map[string]bool{}
			for _, cve := range issue.Cve {
				for _, match := range cveReg.FindAllString(cve, -1) {
					if !cveSeen[match] {
						cveSeen[match] = true
						cves = append(cves, match)
					}
				}
			}
			sort.Strings(cves)

			updates = append(updates, &Update{
				Id:              vid,
				Type:            FreeBsd,
				Vulnerabilities: cves,
				Packages:        []string{pkgRef},
			})
		}
	}

	return
}

func pkgUpdates() (updates []*Update, err error) {
	resp, _ := commander.Exec(&commander.Opt{
		Name: "pkg",
		Args: []string{
			"audit",
			"--fetch",
			"--raw=json",
		},
		Timeout: 120 * time.Second,
		PipeOut: true,
		PipeErr: true,
	})
	if resp == nil {
		return
	}

	updates, err = parsePkgAudit(resp.Output)
	if err != nil {
		logrus.WithFields(
			resp.Map(),
		).Error("telemetry: Failed to parse pkg audit output")
		updates = nil
		err = nil
		return
	}

	return
}

func UpdatesRefresh() (updates []*Update, err error) {
	if IsDnf() {
		updates, err = dnfUpdates()
		return
	}
	if IsPkg() {
		updates, err = pkgUpdates()
		return
	}

	return
}

func dnfUpdates() (updates []*Update, err error) {
	if !IsDnf() {
		return
	}

	pkgMap, err := updatesList()
	if err != nil {
		return
	}

	resp, err := commander.Exec(&commander.Opt{
		Name: "dnf",
		Args: []string{
			"updateinfo",
			"info",
		},
		Timeout: 120 * time.Second,
		PipeOut: true,
		PipeErr: true,
	})
	if err != nil {
		if resp != nil {
			logrus.WithFields(
				resp.Map(),
			).Error("telemetry: Failed to get dnf updateinfo report")
		}
		return
	}

	updates = []*Update{}
	seen := map[string]bool{}
	var current []string

	flush := func() {
		if len(current) == 0 {
			return
		}
		record := current
		current = nil

		upd := parseRecord(record)
		if upd == nil {
			return
		}
		if seen[upd.Id] {
			return
		}
		pkgs, ok := pkgMap[upd.Id]
		if !ok {
			return
		}
		seen[upd.Id] = true
		upd.Packages = pkgs
		updates = append(updates, upd)
	}

	for _, line := range strings.Split(string(resp.Output), "\n") {
		if isSeparatorLine(line) {
			flush()
			continue
		}
		if strings.HasPrefix(
			strings.ReplaceAll(line, " ", ""), "Name:") {

			flush()
		}
		current = append(current, line)
	}
	flush()

	return
}

func init() {
	Register(Updates)
}
