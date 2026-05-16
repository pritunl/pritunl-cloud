package telemetry

import (
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pritunl/pritunl-cloud/advisory"
	"github.com/pritunl/tools/commander"
	"github.com/sirupsen/logrus"
)

type Update struct {
	Advisory    string               `bson:"advisory" json:"advisory"`
	Cves        []string             `bson:"cves" json:"cves"`
	Severity    string               `bson:"severity" json:"severity"`
	Description string               `bson:"description" json:"description"`
	Packages    []string             `bson:"packages" json:"packages"`
	Details     []*advisory.Advisory `bson:"details" json:"details"`
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

var (
	cveReg = regexp.MustCompile(`CVE-\d{4}-\d+`)
)

func severityRank(severity string) int {
	switch severity {
	case Critical:
		return 0
	case Important:
		return 1
	case Moderate:
		return 2
	default:
		return 3
	}
}

func parseSeverity(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "critical":
		return Critical
	case "important":
		return Important
	case "moderate":
		return Moderate
	}
	return ""
}

func isAllowedAdvisory(id string) bool {
	return strings.HasPrefix(id, "RHSA-") ||
		strings.HasPrefix(id, "ALSA-") ||
		strings.HasPrefix(id, "RLSA-") ||
		strings.HasPrefix(id, "ELSA-") ||
		strings.HasPrefix(id, "FEDORA-")
}

func isAllowedArch(pkg string) bool {
	return strings.HasSuffix(pkg, ".x86_64") ||
		strings.HasSuffix(pkg, ".noarch")
}

func isSeparatorLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) == 0 {
		return false
	}
	for _, r := range trimmed {
		if r != '=' {
			return false
		}
	}
	return true
}

func splitRecords(output string) [][]string {
	lines := strings.Split(output, "\n")
	var records [][]string
	var current []string
	inRecord := false

	flush := func() {
		if len(current) > 0 {
			records = append(records, current)
			current = nil
		}
	}

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if isSeparatorLine(line) {
			flush()
			for i+1 < len(lines) && !isSeparatorLine(lines[i+1]) {
				i++
			}
			if i+1 < len(lines) && isSeparatorLine(lines[i+1]) {
				i++
			}
			inRecord = true
			continue
		}

		if strings.HasPrefix(line, "Name        :") {
			flush()
			inRecord = true
		}

		if inRecord {
			current = append(current, line)
		}
	}

	flush()
	return records
}

func parseRecord(lines []string) *Update {
	upd := &Update{}
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
					upd.Cves = append(upd.Cves,
						cveReg.FindAllString(value, -1)...)
				}
			}
			continue
		}

		field := strings.TrimSpace(prefix)
		currentField = field

		switch field {
		case "Update ID", "Name":
			upd.Advisory = value
		case "Severity":
			upd.Severity = parseSeverity(value)
		case "Description":
			if value != "" {
				descLines = append(descLines, value)
			}
		case "CVEs":
			if value != "" {
				upd.Cves = append(upd.Cves,
					cveReg.FindAllString(value, -1)...)
			}
		}
	}

	if !isAllowedAdvisory(upd.Advisory) {
		return nil
	}
	if upd.Severity == "" {
		return nil
	}

	for len(descLines) > 0 && descLines[len(descLines)-1] == "" {
		descLines = descLines[:len(descLines)-1]
	}
	upd.Description = strings.Join(descLines, "\n")

	fullText := strings.Join(lines, "\n")
	cveSet := map[string]bool{}
	deduped := []string{}
	for _, c := range upd.Cves {
		if !cveSet[c] {
			cveSet[c] = true
			deduped = append(deduped, c)
		}
	}
	for _, c := range cveReg.FindAllString(fullText, -1) {
		if !cveSet[c] {
			cveSet[c] = true
			deduped = append(deduped, c)
		}
	}
	sort.Strings(deduped)
	upd.Cves = deduped

	return upd
}

func updatesList() (advisories map[string][]string, err error) {
	if !IsDnf() {
		return
	}

	var resp *commander.Return
	if HasSevs() {
		resp, err = commander.Exec(&commander.Opt{
			Name: "dnf",
			Args: []string{
				"updateinfo",
				"list",
				"--advisory-severities=Moderate,Important,Critical",
			},
			Timeout: 90 * time.Second,
			PipeOut: true,
			PipeErr: true,
		})
	} else {
		resp, err = commander.Exec(&commander.Opt{
			Name: "dnf",
			Args: []string{
				"updateinfo",
				"list",
				"--sec-severity=Moderate",
				"--sec-severity=Important",
				"--sec-severity=Critical",
			},
			Timeout: 90 * time.Second,
			PipeOut: true,
			PipeErr: true,
		})
	}
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
		if !isAllowedAdvisory(adv) {
			continue
		}

		pkg := ""
		part1 := strings.ToLower(parts[1])
		part2 := strings.ToLower(parts[2])

		if strings.Contains(part1, Moderate) ||
			strings.Contains(part1, Important) ||
			strings.Contains(part1, Critical) {

			pkg = parts[2]
		} else if len(parts) >= 4 && (strings.Contains(part2, Moderate) ||
			strings.Contains(part2, Important) ||
			strings.Contains(part2, Critical)) {

			pkg = parts[3]
		} else {
			continue
		}

		if pkg == "" || !isAllowedArch(pkg) {
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

func UpdatesRefresh() (updates []*Update, err error) {
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
	for _, record := range splitRecords(string(resp.Output)) {
		upd := parseRecord(record)
		if upd == nil {
			continue
		}
		if seen[upd.Advisory] {
			continue
		}
		pkgs, ok := pkgMap[upd.Advisory]
		if !ok {
			continue
		}
		seen[upd.Advisory] = true
		upd.Packages = pkgs
		updates = append(updates, upd)
	}

	sort.Slice(updates, func(i, j int) bool {
		ri := severityRank(updates[i].Severity)
		rj := severityRank(updates[j].Severity)
		if ri != rj {
			return ri < rj
		}
		return updates[i].Advisory < updates[j].Advisory
	})

	return
}

func init() {
	Register(Updates)
}
