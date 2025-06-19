package security

import (
	"time"

	"github.com/pritunl/pritunl-cloud/agent/utils"
)

var (
	lastReport     *Report
	lastReportTime time.Time
)

type Report struct {
	Updates []*Update `bson:"updates" json:"updates"`
}

type Update struct {
	Advisory string `bson:"advisory" json:"advisory"`
	Severity string `bson:"severity" json:"severity"`
	Package  string `bson:"package" json:"package"`
}

func Refresh() {
	if time.Since(lastReportTime) < 6*time.Hour {
		return
	}
	lastReportTime = time.Now()

	if !utils.IsDnf() {
		return
	}

	lastReport = dnfGetReport()
}

func GetReport() *Report {
	return lastReport
}
