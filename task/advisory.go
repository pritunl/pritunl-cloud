package task

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/advisory"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/manifest"
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
	advisories := map[bson.ObjectID]map[string]*advisory.Advisory{}
	now := time.Now()

	cursor, err := manifest.FindUpdates(db)
	if err != nil {
		return
	}
	defer cursor.Close()

	for cursor.Next() {
		updts, e := cursor.Decode()
		if e != nil {
			err = e
			return
		}

		for _, updt := range updts.Updates {
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
					updtVulns = append(updtVulns, vuln)
				}
			}

			orgAdvs := advisories[updts.Organization]
			if orgAdvs == nil {
				orgAdvs = map[string]*advisory.Advisory{}
				advisories[updts.Organization] = orgAdvs
			}

			adv := orgAdvs[updt.Id]
			if adv == nil {
				adv = advisory.FromUpdate(
					updt, updts.Organization, now, updtVulns)
				orgAdvs[updt.Id] = adv
			}

			switch updts.Variant {
			case manifest.InstanceVariant:
				adv.Instances = append(adv.Instances, updts.Resource)
			case manifest.NodeVariant:
				adv.Nodes = append(adv.Nodes, updts.Resource)
			}
		}
	}

	err = cursor.Err()
	if err != nil {
		return
	}

	coll := db.Advisories()

	for orgId, orgAdvs := range advisories {
		for advId, adv := range orgAdvs {
			adv.UpdateScore()

			_, err = coll.UpdateOne(db, &bson.M{
				"organization": orgId,
				"reference":    advId,
			}, &bson.M{
				"$set": &bson.M{
					"organization":    orgId,
					"reference":       advId,
					"type":            adv.Type,
					"updated":         adv.Updated,
					"severity":        adv.Severity,
					"description":     adv.Description,
					"score":           adv.Score,
					"packages":        adv.Packages,
					"vulnerabilities": adv.Vulnerabilities,
					"instances":       adv.Instances,
					"nodes":           adv.Nodes,
				},
			}, options.UpdateOne().SetUpsert(true))
			if err != nil {
				err = database.ParseError(err)
				return
			}
		}
	}

	_, err = coll.DeleteMany(db, &bson.M{
		"updated": &bson.M{
			"$lt": now,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func init() {
	register(advisoryData)
}
