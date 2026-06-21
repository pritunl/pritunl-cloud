package task

import (
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/advisory"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/manifest"
	"github.com/pritunl/pritunl-cloud/vulnerability"
	"github.com/pritunl/pritunl-cloud/vuxml"
	"github.com/sirupsen/logrus"
)

var advisoryData = &Task{
	Name:    "advisory",
	Version: 1,
	Hours:   []int{0, 3, 6, 9, 12, 15, 18, 21},
	Minutes: []int{22},
	Handler: advisoryDataHandler,
}

type advisoryProcessor struct {
	now             time.Time
	vulnerabilities map[string]*vulnerability.Vulnerability
	advisories      map[bson.ObjectID]map[string]*advisory.Advisory
	vuxmlDb         map[string]*vuxml.VuxmlEntry
	dismissals      map[bson.ObjectID]map[string]*advisory.Dismissal
}

func (a *advisoryProcessor) Run(db *database.Database) (err error) {
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

		err = a.parseUpdates(db, *updts)
		if err != nil {
			return
		}
	}

	err = cursor.Err()
	if err != nil {
		return
	}

	coll := db.Advisories()

	for orgId, orgAdvs := range a.advisories {
		for advId, adv := range orgAdvs {

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
					"vuxmls":          adv.Vuxmls,
					"vulnerabilities": adv.Vulnerabilities,
					"instances":       adv.Instances,
					"nodes":           adv.Nodes,
				},
				"$setOnInsert": &bson.M{
					"dismissed":           false,
					"dismissed_resources": []bson.ObjectID{},
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
			"$lt": a.now,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (a *advisoryProcessor) parseUpdates(db *database.Database,
	updts manifest.Updates) (err error) {

	orgAdvs := a.advisories[updts.Organization]
	if orgAdvs == nil {
		orgAdvs = map[string]*advisory.Advisory{}
		a.advisories[updts.Organization] = orgAdvs
	}
	orgDismissals := a.dismissals[updts.Organization]

	resourceAdvs := []*advisory.Advisory{}
	resourceAdvsSet := set.NewSet()
	for _, updt := range updts.Updates {
		if updt.Id == "" {
			continue
		}

		if updt.Type == advisory.RedHat {
			updtVulns := []*vulnerability.Vulnerability{}

			for _, vulnId := range updt.Vulnerabilities {
				vuln, ok := a.vulnerabilities[vulnId]
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
					a.vulnerabilities[vulnId] = vuln
				}

				if vuln != nil {
					updtVulns = append(updtVulns, vuln)
				}
			}

			adv := orgAdvs[updt.Id]
			if adv == nil {
				adv = advisory.FromUpdate(
					updt, updts.Organization, a.now, updtVulns)

				if orgDismissals != nil {
					dism := orgDismissals[adv.Reference]
					if dism != nil {
						adv.Dismissed = dism.Dismissed
						adv.DismissedResources = dism.DismissedResources
					}
				}

				orgAdvs[updt.Id] = adv
			}

			adv.MergePackages(updt.Packages)

			adv.UpdateScore()

			switch updts.Variant {
			case manifest.InstanceVariant:
				adv.Instances = append(adv.Instances, updts.Resource)
			case manifest.NodeVariant:
				adv.Nodes = append(adv.Nodes, updts.Resource)
			}

			resourceAdvs = append(resourceAdvs, adv)
		} else if updt.Type == advisory.FreeBsd {
			if a.vuxmlDb == nil {
				a.vuxmlDb, err = vuxml.Load()
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"error": err,
					}).Error("task: Failed to load FreeBSD vuxml")
					return
				}
			}

			entry := a.vuxmlDb[updt.Id]
			if entry == nil {
				continue
			}

			updtVulns := []*vulnerability.Vulnerability{}

			for _, vulnId := range updt.Vulnerabilities {
				vuln, ok := a.vulnerabilities[vulnId]
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
					a.vulnerabilities[vulnId] = vuln
				}

				if vuln != nil {
					updtVulns = append(updtVulns, vuln)
				}
			}

			for _, pkg := range updt.Packages {
				pkgName, _, _ := strings.Cut(pkg, "@")
				if pkgName == "" {
					continue
				}

				ref, ok := entry.Reference(pkgName)
				if !ok {
					continue
				}

				adv := orgAdvs[ref]
				if adv == nil {
					adv = advisory.NewUpdate(
						ref,
						advisory.FreeBsd,
						updts.Organization,
						a.now,
					)

					if orgDismissals != nil {
						dism := orgDismissals[adv.Reference]
						if dism != nil {
							adv.Dismissed = dism.Dismissed
							adv.DismissedResources = dism.DismissedResources
						}
					}

					orgAdvs[ref] = adv
				}

				adv.MergeVuxml(pkg, entry, updtVulns)

				adv.UpdateScore()

				if !resourceAdvsSet.Contains(adv.Reference) {
					resourceAdvsSet.Add(adv.Reference)

					switch updts.Variant {
					case manifest.InstanceVariant:
						adv.Instances = append(
							adv.Instances, updts.Resource)
					case manifest.NodeVariant:
						adv.Nodes = append(adv.Nodes, updts.Resource)
					}

					resourceAdvs = append(resourceAdvs, adv)
				}
			}
		}
	}

	advCount, advMax := advisory.CountResource(
		updts.Resource, resourceAdvs)

	if advCount != updts.Count || advMax != updts.Max {
		var resourceColl *database.Collection
		switch updts.Variant {
		case manifest.InstanceVariant:
			resourceColl = db.Instances()
		case manifest.NodeVariant:
			resourceColl = db.Nodes()
		}

		if resourceColl != nil {
			_, err = resourceColl.UpdateOne(db, &bson.M{
				"_id": updts.Resource,
			}, &bson.M{
				"$set": &bson.M{
					"advisory_count": advCount,
					"advisory_max":   advMax,
				},
			})
			if err != nil {
				err = database.ParseError(err)
				if _, ok := err.(*database.NotFoundError); ok {
					err = nil
				} else {
					return
				}
			}
		}

		updts.Count = advCount
		updts.Max = advMax

		_, err = db.Manifests().UpdateOne(db, &bson.M{
			"_id": updts.Id,
		}, &bson.M{
			"$set": &bson.M{
				"count": advCount,
				"max":   advMax,
			},
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}
	}

	return
}

func advisoryDataHandler(db *database.Database) (err error) {
	advProc := &advisoryProcessor{
		vulnerabilities: map[string]*vulnerability.Vulnerability{},
		advisories:      map[bson.ObjectID]map[string]*advisory.Advisory{},
		now:             time.Now(),
	}

	advProc.dismissals, err = advisory.GetDismissals(db)
	if err != nil {
		return
	}

	err = advProc.Run(db)
	if err != nil {
		return
	}

	return
}

func init() {
	register(advisoryData)
}
