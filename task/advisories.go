package task

import (
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/advisory"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/manifest"
	"github.com/pritunl/pritunl-cloud/vulnerability"
	"github.com/pritunl/pritunl-cloud/vuxml"
	"github.com/sirupsen/logrus"
)

const (
	advisoriesUpsertBatch = 100
)

var advisoriesUpdate = &Task{
	Name:    "advisories_update",
	Version: 1,
	Hours: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
	Minutes: []int{22},
	Handler: advisoriesUpdateHandler,
}

type advisoryProcessor struct {
	now             time.Time
	vulnerabilities map[string]*vulnerability.VulnerabilityMini
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

		err = a.parseUpdates(db, updts)
		if err != nil {
			return
		}
	}

	err = cursor.Err()
	if err != nil {
		return
	}

	batch := make([]*advisory.Advisory, 0, advisoriesUpsertBatch)
	for _, orgAdvs := range a.advisories {
		for _, adv := range orgAdvs {
			batch = append(batch, adv)
			if len(batch) < advisoriesUpsertBatch {
				continue
			}

			err = advisory.UpsertMulti(db, batch)
			if err != nil {
				return
			}
			batch = batch[:0]
		}
	}

	err = advisory.UpsertMulti(db, batch)
	if err != nil {
		return
	}

	coll := db.Advisories()

	_, err = coll.DeleteMany(db, &bson.M{
		"updated": &bson.M{
			"$lt": a.now,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	err = advisory.SweepResources(db, a.now,
		func(orgId bson.ObjectID, reference string) bool {
			return a.advisories[orgId][reference] != nil
		})
	if err != nil {
		return
	}

	return
}

func (a *advisoryProcessor) loadVulnerabilities(db *database.Database,
	updts *manifest.Updates) (err error) {

	lookupIds := []string{}
	lookupSet := set.NewSet()
	for _, updt := range updts.Updates {
		for _, vulnId := range updt.Vulnerabilities {
			vulnId = strings.ToUpper(vulnId)
			if !vulnerability.ValidId(vulnId) ||
				lookupSet.Contains(vulnId) {

				continue
			}

			_, ok := a.vulnerabilities[vulnId]
			if ok {
				continue
			}

			lookupSet.Add(vulnId)
			lookupIds = append(lookupIds, vulnId)
		}
	}

	if len(lookupIds) == 0 {
		return
	}

	vulns, err := vulnerability.GetMini(db, lookupIds)
	if err != nil {
		return
	}

	newIds := []string{}
	referenceIds := []string{}
	for _, vulnId := range lookupIds {
		vuln := vulns[vulnId]
		if vuln == nil {
			vuln = &vulnerability.VulnerabilityMini{
				Id:   vulnId,
				Sync: vulnerability.Pending,
			}
			newIds = append(newIds, vulnId)
		} else if vuln.NeedsReference() {
			referenceIds = append(referenceIds, vulnId)
		}

		a.vulnerabilities[vulnId] = vuln
	}

	err = vulnerability.Queue(db, a.now, newIds, referenceIds)
	if err != nil {
		return
	}

	return
}

func (a *advisoryProcessor) getAdvisory(orgId bson.ObjectID,
	ref string, newAdv func() *advisory.Advisory) (adv *advisory.Advisory) {

	orgAdvs := a.advisories[orgId]
	if orgAdvs == nil {
		orgAdvs = map[string]*advisory.Advisory{}
		a.advisories[orgId] = orgAdvs
	}

	adv = orgAdvs[ref]
	if adv == nil {
		adv = newAdv()

		orgDismissals := a.dismissals[orgId]
		if orgDismissals != nil {
			dism := orgDismissals[adv.Reference]
			if dism != nil {
				adv.Dismissed = dism.Dismissed
			}
		}

		orgAdvs[ref] = adv
	}

	return
}

func (a *advisoryProcessor) parseUpdates(db *database.Database,
	updts *manifest.Updates) (err error) {

	err = a.loadVulnerabilities(db, updts)
	if err != nil {
		return
	}

	resourceAdvs := []*advisory.Advisory{}
	resourceAdvsSet := set.NewSet()
	for _, updt := range updts.Updates {
		if updt.Id == "" {
			continue
		}

		if updt.Type == advisory.RedHat {
			adv := a.getAdvisory(updts.Organization, updt.Id,
				func() *advisory.Advisory {
					return advisory.FromUpdate(updt, updts.Organization,
						a.now, a.vulnerabilities)
				},
			)

			adv.MergePackages(updt.Packages)

			adv.UpdateScore()

			if !resourceAdvsSet.Contains(adv.Reference) {
				resourceAdvsSet.Add(adv.Reference)
				resourceAdvs = append(resourceAdvs, adv)
			}
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

			for _, pkg := range updt.Packages {
				pkgName, _, _ := strings.Cut(pkg, "@")
				if pkgName == "" {
					continue
				}

				ref, ok := entry.Reference(pkgName)
				if !ok {
					continue
				}

				adv := a.getAdvisory(updts.Organization, ref,
					func() *advisory.Advisory {
						return advisory.NewUpdate(
							ref,
							advisory.FreeBsd,
							updts.Organization,
							a.now,
						)
					},
				)

				adv.MergeVuxml(pkg, entry, updt.Vulnerabilities,
					a.vulnerabilities)

				adv.UpdateScore()

				if !resourceAdvsSet.Contains(adv.Reference) {
					resourceAdvsSet.Add(adv.Reference)
					resourceAdvs = append(resourceAdvs, adv)
				}
			}
		}
	}

	kind := ""
	switch updts.Variant {
	case manifest.InstanceVariant:
		kind = advisory.Instance
	case manifest.NodeVariant:
		kind = advisory.Node
	}

	components := vulnerability.NewComponents(updts.Components())
	orgDismissals := a.dismissals[updts.Organization]
	counter := &advisory.Counter{}
	resources := []*advisory.Resource{}

	for _, adv := range resourceAdvs {
		state, exclusions := adv.ResourceState(components)
		dismissed := orgDismissals[adv.Reference].HasResource(
			updts.Resource)

		if state == advisory.Affected {
			switch kind {
			case advisory.Instance:
				adv.InstanceCount += 1
			case advisory.Node:
				adv.NodeCount += 1
			}
		}

		counter.Add(adv, state, dismissed)

		resources = append(resources, &advisory.Resource{
			Organization: updts.Organization,
			Reference:    adv.Reference,
			Resource:     updts.Resource,
			Kind:         kind,
			State:        state,
			Exclusions:   exclusions,
			Timestamp:    a.now,
		})
	}

	unreachable := 0
	excluded := 0
	for _, res := range resources {
		if res.State == advisory.Unreachable {
			unreachable += 1
		}
		excluded += len(res.Exclusions)
	}

	err = advisory.UpsertResources(db, resources)
	if err != nil {
		return
	}

	if counter.Count != updts.Count || counter.Max != updts.Max ||
		counter.Pending != updts.Pending {

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
					"advisory_count":   counter.Count,
					"advisory_max":     counter.Max,
					"advisory_pending": counter.Pending,
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

		updts.Count = counter.Count
		updts.Max = counter.Max
		updts.Pending = counter.Pending

		_, err = db.Manifests().UpdateOne(db, &bson.M{
			"_id": updts.Id,
		}, &bson.M{
			"$set": &bson.M{
				"count":   counter.Count,
				"max":     counter.Max,
				"pending": counter.Pending,
			},
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}
	}

	return
}

func advisoriesUpdateHandler(db *database.Database) (err error) {
	advProc := &advisoryProcessor{
		vulnerabilities: map[string]*vulnerability.VulnerabilityMini{},
		advisories:      map[bson.ObjectID]map[string]*advisory.Advisory{},
		now:             time.Now(),
	}

	advProc.dismissals, err = advisory.GetDismissals(db)
	if err != nil {
		return
	}

	dismissedAdvs := 0
	dismissedRes := 0
	for _, orgDismissals := range advProc.dismissals {
		for _, dism := range orgDismissals {
			if dism.Dismissed {
				dismissedAdvs += 1
			}
			dismissedRes += dism.Resources.Len()
		}
	}

	err = advProc.Run(db)
	if err != nil {
		return
	}

	advCount := 0
	for _, orgAdvs := range advProc.advisories {
		advCount += len(orgAdvs)
	}

	logrus.WithFields(logrus.Fields{
		"advisories":      advCount,
		"vulnerabilities": len(advProc.vulnerabilities),
		"duration":        time.Since(advProc.now).String(),
	}).Info("task: Advisory task complete")

	return
}

func init() {
	register(advisoriesUpdate)
}
