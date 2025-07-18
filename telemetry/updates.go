package telemetry

import (
	"strings"
	"time"

	"github.com/pritunl/tools/commander"
	"github.com/sirupsen/logrus"
)

type Update struct {
	Advisory string `bson:"advisory" json:"advisory"`
	Severity string `bson:"severity" json:"severity"`
	Package  string `bson:"package" json:"package"`
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

func updatesRefresh() (updates []*Update, err error) {
	if !IsDnf() {
		return
	}

	resp, err := commander.Exec(&commander.Opt{
		Name: "dnf",
		Args: []string{
			"updateinfo",
			"list",
			"--sec-severity=Moderate",
			"--sec-severity=Important",
			"--sec-severity=Critical",
		},
		Timeout: 30 * time.Second,
		PipeOut: true,
		PipeErr: true,
	})
	if err != nil {
		logrus.WithFields(
			resp.Map(),
		).Error("security: Failed to get dnf security update report")
		return
	}

	lines := strings.Split(string(resp.Output), "\n")

	moderateUpdates := []*Update{}
	importantUpdates := []*Update{}
	criticalUpdates := []*Update{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Last metadata") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 3 {
			advisory := parts[0]
			severity := parts[1]
			pkg := parts[2]

			if strings.Contains(
				strings.ToLower(severity), "moderate") {

				moderateUpdates = append(moderateUpdates, &Update{
					Advisory: advisory,
					Severity: "moderate",
					Package:  pkg,
				})
			} else if strings.Contains(
				strings.ToLower(severity), "important") {

				importantUpdates = append(importantUpdates, &Update{
					Advisory: advisory,
					Severity: "important",
					Package:  pkg,
				})
			} else if strings.Contains(
				strings.ToLower(severity), "critical") {

				criticalUpdates = append(criticalUpdates, &Update{
					Advisory: advisory,
					Severity: "critical",
					Package:  pkg,
				})
			} else {
				continue
			}
		}
	}

	updates = append(updates, criticalUpdates...)
	updates = append(updates, importantUpdates...)
	updates = append(updates, moderateUpdates...)

	return
}

func init() {
	Register(Updates)
}
