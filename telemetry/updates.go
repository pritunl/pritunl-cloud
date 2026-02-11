package telemetry

import (
	"sort"
	"strings"
	"time"

	"github.com/pritunl/tools/commander"
	"github.com/sirupsen/logrus"
)

type Update struct {
	Advisories []string `bson:"advisories" json:"advisories"`
	Cves       []string `bson:"cves" json:"cves"`
	Severity   string   `bson:"severity" json:"severity"`
	Package    string   `bson:"package" json:"package"`
}

var Updates = &Telemetry[[]*Update]{
	TransmitRate: 6 * time.Minute,
	RefreshRate:  6 * time.Hour,
	Refresher:    updatesRefresh,
	Validate: func(data []*Update) []*Update {
		if len(data) > 50 {
			return data[:50]
		}
		return data
	},
}

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

func updatesRefresh() (updates []*Update, err error) {
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
		logrus.WithFields(
			resp.Map(),
		).Error("security: Failed to get dnf security update report")
		return
	}

	lines := strings.Split(string(resp.Output), "\n")

	packages := map[string]*Update{}

	for _, line := range lines {
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

		advisory := parts[0]
		severity := ""
		pkg := ""
		part1 := parts[1]
		part2 := parts[2]
		part3 := ""
		if len(parts) >= 4 {
			part3 = parts[3]
		}

		if !strings.HasPrefix(advisory, "RHSA-") &&
			!strings.HasPrefix(advisory, "ALSA-") &&
			!strings.HasPrefix(advisory, "RLSA-") &&
			!strings.HasPrefix(advisory, "ELSA-") &&
			!strings.HasPrefix(advisory, "FEDORA-") {

			continue
		}

		if strings.Contains(strings.ToLower(part1), Moderate) {
			severity = Moderate
			pkg = part2
		} else if strings.Contains(strings.ToLower(part1), Important) {
			severity = Important
			pkg = part2
		} else if strings.Contains(strings.ToLower(part1), Critical) {
			severity = Critical
			pkg = part2
		} else if strings.Contains(strings.ToLower(part2), Moderate) {
			severity = Moderate
			pkg = part3
		} else if strings.Contains(strings.ToLower(part2), Important) {
			severity = Important
			pkg = part3
		} else if strings.Contains(strings.ToLower(part2), Critical) {
			severity = Critical
			pkg = part3
		} else {
			continue
		}

		upd, ok := packages[pkg]
		if !ok {
			upd = &Update{
				Severity: severity,
				Package:  pkg,
			}
			packages[pkg] = upd
		}

		found := false
		for _, a := range upd.Advisories {
			if a == advisory {
				found = true
				break
			}
		}

		if !found {
			upd.Advisories = append(upd.Advisories, advisory)
		}

		if severityRank(severity) < severityRank(upd.Severity) {
			upd.Severity = severity
		}
	}

	cveLines := []string{}
	resp, err = commander.Exec(&commander.Opt{
		Name: "dnf",
		Args: []string{
			"updateinfo",
			"list",
			"--with-cve",
		},
		Timeout: 60 * time.Second,
		PipeOut: true,
		PipeErr: true,
	})
	if err != nil {
		logrus.WithFields(
			resp.Map(),
		).Error("security: Failed to get dnf security cve report")
		err = nil
	} else {
		cveLines = strings.Split(string(resp.Output), "\n")
	}

	for _, line := range cveLines {
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

		cve := parts[0]
		pkg := ""
		part1 := parts[1]
		part2 := parts[2]
		part3 := ""
		if len(parts) >= 4 {
			part3 = parts[3]
		}

		if strings.Contains(strings.ToLower(part1), Moderate) ||
			strings.Contains(strings.ToLower(part1), Important) ||
			strings.Contains(strings.ToLower(part1), Critical) {

			pkg = part2
		} else if strings.Contains(strings.ToLower(part2), Moderate) ||
			strings.Contains(strings.ToLower(part2), Important) ||
			strings.Contains(strings.ToLower(part2), Critical) {

			pkg = part3
		}

		if pkg == "" {
			continue
		}

		upd, ok := packages[pkg]
		if !ok {
			continue
		}

		found := false
		for _, c := range upd.Cves {
			if c == cve {
				found = true
				break
			}
		}

		if !found {
			upd.Cves = append(upd.Cves, cve)
		}
	}

	updates = []*Update{}
	for _, upd := range packages {
		updates = append(updates, upd)
	}

	sort.Slice(updates, func(i, j int) bool {
		ri := severityRank(updates[i].Severity)
		rj := severityRank(updates[j].Severity)
		if ri != rj {
			return ri < rj
		}
		return updates[i].Package < updates[j].Package
	})

	return
}

func init() {
	Register(Updates)
}
