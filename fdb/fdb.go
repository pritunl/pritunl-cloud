package fdb

import (
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

var (
	curRecords  set.Set
	lastRefresh time.Time
)

type Record struct {
	Mac   string
	Iface string
}

func getRecords() (records set.Set, err error) {
	records = set.NewSet()

	output, err := utils.ExecOutput("", "bridge", "fdb")
	if err != nil {
		return
	}

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 || len(fields[0]) != 17 ||
			(!strings.HasPrefix(fields[0], "00:") &&
				!strings.HasPrefix(fields[0], "04:")) ||
			fields[1] != "dev" {

			continue
		}

		iface := fields[2]
		if len(iface) != 14 || !strings.HasPrefix(iface, "j") {
			continue
		}

		static := false
		sticky := false
		vlan := false
		for _, field := range fields[3:] {
			if field == "static" {
				static = true
			} else if field == "sticky" {
				sticky = true
			} else if field == "vlan" {
				vlan = true
			}
		}
		if !static || !sticky || vlan {
			continue
		}

		records.Add(Record{
			Mac:   fields[0],
			Iface: iface,
		})
	}

	return
}

func ApplyState(stat *state.State) (err error) {
	newRecords := set.NewSet()
	for _, inst := range stat.Instances() {
		if !inst.IsActive() {
			continue
		}

		iface := vm.GetIfaceNodeInternal(inst.Id, 0)
		if !stat.HasInterfaces(iface) {
			continue
		}

		newRecords.Add(Record{
			Mac:   vm.GetMacAddr(inst.Id, inst.Vpc),
			Iface: iface,
		})
		newRecords.Add(Record{
			Mac:   vm.GetMacAddrInternal(inst.Id, inst.Vpc),
			Iface: iface,
		})
	}

	if curRecords == nil {
		records, e := getRecords()
		if e != nil {
			err = e
			return
		}

		curRecords = records
		lastRefresh = time.Now()
	} else if time.Since(lastRefresh) > refreshTtl {
		records, e := getRecords()
		if e != nil {
			err = e
			return
		}

		driftAdd := records.Copy()
		driftAdd.Subtract(curRecords)
		driftRem := curRecords.Copy()
		driftRem.Subtract(records)
		if driftAdd.Len() > 0 || driftRem.Len() > 0 {
			logrus.WithFields(logrus.Fields{
				"added":   driftAdd.Len(),
				"removed": driftRem.Len(),
			}).Warning("fdb: Unexpected fdb table change")
		}

		curRecords = records
		lastRefresh = time.Now()
	}

	remRecords := curRecords.Copy()
	remRecords.Subtract(newRecords)
	if remRecords.Len() > 0 {
		logrus.WithFields(logrus.Fields{
			"entries": remRecords.Len(),
		}).Info("fdb: Removing instance fdb entries")
	}
	for recordInf := range remRecords.Iter() {
		recrd := recordInf.(Record)

		_, err = utils.ExecCombinedOutputLogged(
			[]string{
				"Cannot find device",
				"No such file",
			},
			"bridge", "fdb",
			"del", recrd.Mac,
			"dev", recrd.Iface,
			"master",
		)
		if err != nil {
			return
		}

		curRecords.Remove(recrd)
	}

	addRecords := newRecords.Copy()
	addRecords.Subtract(curRecords)
	if addRecords.Len() > 0 {
		logrus.WithFields(logrus.Fields{
			"entries": addRecords.Len(),
		}).Info("fdb: Adding instance fdb entries")
	}
	for recordInf := range addRecords.Iter() {
		recrd := recordInf.(Record)

		output, e := utils.ExecCombinedOutputLogged(
			[]string{
				"Cannot find device",
			},
			"bridge", "fdb",
			"replace", recrd.Mac,
			"dev", recrd.Iface,
			"master", "static", "sticky",
		)
		if e != nil {
			err = e
			return
		}

		if strings.Contains(output, "Cannot find device") {
			continue
		}

		curRecords.Add(recrd)
	}

	return
}
