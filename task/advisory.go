package task

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/telemetry"
	"github.com/pritunl/pritunl-cloud/vulnerability"
	"github.com/sirupsen/logrus"
)

var advisoryData = &Task{
	Name:    "advisory",
	Version: 1,
	Hours:   []int{0, 3, 6, 9, 12, 15, 18, 21},
	Minutes: []int{22},
	Handler: advisoryDataHandler,
}

func advisoryDataHandler(db *database.Database) (err error) {
	vulnerabilities := map[string]*vulnerability.Vulnerability{}

	coll := db.Instances()

	cursor, err := coll.Find(db, &bson.M{})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		inst := &instance.Instance{}
		err = cursor.Decode(inst)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		if inst.Guest == nil {
			continue
		}

		vulns := []*vulnerability.Vulnerability{}
		updtsData := map[string]*telemetry.UpdateData{}
		for _, updt := range inst.Guest.Updates {
			updtData := &telemetry.UpdateData{}
			updtVulns := []*vulnerability.Vulnerability{}

			for _, vulnId := range updt.Vulnerabilities {
				vuln, ok := vulnerabilities[vulnId]
				if !ok {
					vuln, err = vulnerability.GetOneLimit(db, vulnId)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"vulnerability": vulnId,
							"error":         err,
						}).Error("task: Failed to query vulnerability")
						err = nil
						vuln = nil
					}
					vulnerabilities[vulnId] = vuln
				}

				if vuln != nil {
					vulns = append(vulns, vuln)
					updtVulns = append(updtVulns, vuln)
				}
			}

			updtData.Score = updt.GetScore(updtVulns)
			updtsData[updt.Id] = updtData
		}
		inst.Guest.UpdatesData = updtsData
		inst.Guest.Vulnerabilities = vulns

		err = inst.CommitFields(db, set.NewSet(
			"guest.updates_data",
			"guest.vulnerabilities",
		))
		if err != nil {
			return
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Nodes()

	cursor, err = coll.Find(db, &bson.M{})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		nde := &node.Node{}
		err = cursor.Decode(nde)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		vulns := []*vulnerability.Vulnerability{}
		updtsData := map[string]*telemetry.UpdateData{}
		for _, updt := range nde.Updates {
			updtData := &telemetry.UpdateData{}
			updtVulns := []*vulnerability.Vulnerability{}

			for _, vulnId := range updt.Vulnerabilities {
				vuln, ok := vulnerabilities[vulnId]
				if !ok {
					vuln, err = vulnerability.GetOneLimit(db, vulnId)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"vulnerability": vulnId,
							"error":         err,
						}).Error("task: Failed to query vulnerability")
						err = nil
						vuln = nil
					}
					vulnerabilities[vulnId] = vuln
				}

				if vuln != nil {
					vulns = append(vulns, vuln)
					updtVulns = append(updtVulns, vuln)
				}
			}

			updtData.Score = updt.GetScore(updtVulns)
			updtsData[updt.Id] = updtData
		}
		nde.UpdatesData = updtsData
		nde.Vulnerabilities = vulns

		err = nde.CommitFields(
			db, set.NewSet("updates_data", "vulnerabilities"))
		if err != nil {
			return
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func init() {
	register(advisoryData)
}
