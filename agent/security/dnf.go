package security

import (
	"strings"
	"time"

	"github.com/pritunl/tools/commander"
	"github.com/sirupsen/logrus"
)

func dnfGetReport() (report *Report) {
	report = &Report{
		Updates: []*Update{},
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

			if strings.Contains(strings.ToLower(severity), "moderate") {
				moderateUpdates = append(moderateUpdates, &Update{
					Advisory: advisory,
					Severity: "moderate",
					Package:  pkg,
				})
			} else if strings.Contains(strings.ToLower(severity), "important") {
				importantUpdates = append(importantUpdates, &Update{
					Advisory: advisory,
					Severity: "important",
					Package:  pkg,
				})
			} else if strings.Contains(strings.ToLower(severity), "critical") {
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

	report.Updates = append(report.Updates, criticalUpdates...)
	report.Updates = append(report.Updates, importantUpdates...)
	report.Updates = append(report.Updates, moderateUpdates...)

	return
}
