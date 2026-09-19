package advisory

import (
	"slices"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vulnerability"
)

func Get(db *database.Database, advId bson.ObjectID) (
	adv *Advisory, err error) {

	coll := db.Advisories()
	adv = &Advisory{}

	err = coll.FindOneId(advId, adv)
	if err != nil {
		return
	}

	return
}

func GetOne(db *database.Database, query *bson.M) (adv *Advisory, err error) {
	coll := db.Advisories()
	adv = &Advisory{}

	err = coll.FindOne(db, query).Decode(adv)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, advId bson.ObjectID) (
	adv *Advisory, err error) {

	coll := db.Advisories()
	adv = &Advisory{}

	err = coll.FindOne(db, &bson.M{
		"_id":          advId,
		"organization": orgId,
	}).Decode(adv)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M,
	opts ...options.Lister[options.FindOptions]) (
	advisories []*Advisory, err error) {

	coll := db.Advisories()
	advisories = []*Advisory{}

	cursor, err := coll.Find(db, query, opts...)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		adv := &Advisory{}
		err = cursor.Decode(adv)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		advisories = append(advisories, adv)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

var countProjection = &bson.M{
	"organization": 1,
	"reference":    1,
	"dismissed":    1,
	"updated":      1,
	"score":        1,
	"complete":     1,
}

func GetResourceAdvisories(db *database.Database,
	resourceId bson.ObjectID) (advisories []*Advisory,
	resources map[bson.ObjectID]*Resource, err error) {

	advisories, resources, err = getResourceAdvisories(db, resourceId)
	if err != nil {
		return
	}

	return
}

func getResourceAdvisories(db *database.Database,
	resourceId bson.ObjectID, opts ...options.Lister[options.FindOptions]) (
	advisories []*Advisory, resources map[bson.ObjectID]*Resource,
	err error) {

	advisories = []*Advisory{}
	resources = map[bson.ObjectID]*Resource{}

	joins, err := GetResources(db, &bson.M{
		"resource": resourceId,
	})
	if err != nil {
		return
	}

	orgRefs := map[bson.ObjectID][]string{}
	joinsMap := map[bson.ObjectID]map[string]*Resource{}
	for _, res := range joins {
		orgRefs[res.Organization] = append(
			orgRefs[res.Organization], res.Reference)

		orgJoins := joinsMap[res.Organization]
		if orgJoins == nil {
			orgJoins = map[string]*Resource{}
			joinsMap[res.Organization] = orgJoins
		}
		orgJoins[res.Reference] = res
	}

	for orgId, refs := range orgRefs {
		advs, e := GetAll(db, &bson.M{
			"organization": orgId,
			"reference": &bson.M{
				"$in": refs,
			},
		}, opts...)
		if e != nil {
			err = e
			return
		}

		for _, adv := range advs {
			res := joinsMap[orgId][adv.Reference]
			if res == nil || res.Timestamp.Before(adv.Updated) {
				continue
			}

			advisories = append(advisories, adv)
			resources[adv.Id] = res
		}
	}

	slices.SortFunc(advisories, func(a, b *Advisory) int {
		return strings.Compare(a.Reference, b.Reference)
	})

	return
}

func updateResource(db *database.Database, resId bson.ObjectID,
	kind string) (err error) {

	advisories, resources, err := getResourceAdvisories(db, resId,
		options.Find().SetProjection(countProjection))
	if err != nil {
		return
	}

	counter := &Counter{}
	for _, adv := range advisories {
		res := resources[adv.Id]
		counter.Add(adv, res.State, res.Dismissed)
	}

	var coll *database.Collection
	if kind == Node {
		coll = db.Nodes()
	} else {
		coll = db.Instances()
	}

	_, err = coll.UpdateOne(db, &bson.M{
		"_id": resId,
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

	return
}

func UpdateInstance(db *database.Database, instId bson.ObjectID) (err error) {
	err = updateResource(db, instId, Instance)
	if err != nil {
		return
	}

	return
}

func UpdateNode(db *database.Database, nodeId bson.ObjectID) (err error) {
	err = updateResource(db, nodeId, Node)
	if err != nil {
		return
	}

	return
}

func UpdateResource(db *database.Database, resId bson.ObjectID) (err error) {
	coll := db.Instances()

	count, err := coll.CountDocuments(db, &bson.M{
		"_id": resId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	if count > 0 {
		err = UpdateInstance(db, resId)
		if err != nil {
			return
		}
	}

	coll = db.Nodes()

	count, err = coll.CountDocuments(db, &bson.M{
		"_id": resId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	if count > 0 {
		err = UpdateNode(db, resId)
		if err != nil {
			return
		}
	}

	return
}

func UpdateResourceOrg(db *database.Database,
	resId, orgId bson.ObjectID) (err error) {

	coll := db.Instances()

	count, err := coll.CountDocuments(db, &bson.M{
		"_id":          resId,
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	if count > 0 {
		err = UpdateInstance(db, resId)
		if err != nil {
			return
		}
	}

	return
}

func updateResources(db *database.Database, advs ...*Advisory) (err error) {
	updated := set.NewSet()

	for _, adv := range advs {
		resources, e := GetResources(db, &bson.M{
			"organization": adv.Organization,
			"reference":    adv.Reference,
		}, options.Find().SetProjection(&bson.M{
			"resource": 1,
			"kind":     1,
		}))
		if e != nil {
			err = e
			return
		}

		for _, res := range resources {
			if updated.Contains(res.Resource) {
				continue
			}
			updated.Add(res.Resource)

			err = updateResource(db, res.Resource, res.Kind)
			if err != nil {
				return
			}
		}
	}

	return
}

func UpsertMulti(db *database.Database, advs []*Advisory) (err error) {
	if len(advs) == 0 {
		return
	}

	models := make([]mongo.WriteModel, 0, len(advs))
	for _, adv := range advs {
		models = append(models, adv.UpsertModel())
	}

	coll := db.Advisories()

	_, err = coll.BulkWrite(db, models,
		options.BulkWrite().SetOrdered(false))
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (advisories []*Advisory, count int64, err error) {

	coll := db.Advisories()
	advisories = []*Advisory{}

	if len(*query) == 0 {
		count, err = coll.EstimatedDocumentCount(db)
		if err != nil {
			err = database.ParseError(err)
			return
		}
	} else {
		count, err = coll.CountDocuments(db, query)
		if err != nil {
			err = database.ParseError(err)
			return
		}
	}

	if pageCount == 0 {
		pageCount = 20
	}
	maxPage := count / pageCount
	if count == pageCount {
		maxPage = 0
	}
	page = utils.Min64(page, maxPage)
	skip := utils.Min64(page*pageCount, count)

	cursor, err := coll.Find(
		db,
		query,
		options.Find().
			SetSort(bson.D{{"reference", 1}}).
			SetSkip(skip).
			SetLimit(pageCount),
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		adv := &Advisory{}
		err = cursor.Decode(adv)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		advisories = append(advisories, adv)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, advId bson.ObjectID) (err error) {
	coll := db.Advisories()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": advId,
	})
	if err != nil {
		err = database.ParseError(err)
		switch err.(type) {
		case *database.NotFoundError:
			err = nil
		default:
			return
		}
	}

	return
}

func RemoveOrg(db *database.Database, orgId, advId bson.ObjectID) (
	err error) {

	coll := db.Advisories()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id":          advId,
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		switch err.(type) {
		case *database.NotFoundError:
			err = nil
		default:
			return
		}
	}

	return
}

func RemoveMulti(db *database.Database, advIds []bson.ObjectID) (
	err error) {

	coll := db.Advisories()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": advIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveMultiOrg(db *database.Database, orgId bson.ObjectID,
	advIds []bson.ObjectID) (err error) {

	coll := db.Advisories()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": advIds,
		},
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func updateDismiss(db *database.Database, adv *Advisory,
	dismiss, restore bool, dismissals, restores []bson.ObjectID) (
	changed bool, err error) {

	if dismiss || restore {
		coll := db.Advisories()

		adv.Dismissed = dismiss

		_, err = coll.UpdateOne(db, &bson.M{
			"_id": adv.Id,
		}, &bson.M{
			"$set": &bson.M{
				"dismissed": adv.Dismissed,
			},
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}

		changed = true
	}

	coll := db.AdvisoryResources()

	if len(dismissals) > 0 {
		_, err = coll.UpdateMany(db, &bson.M{
			"organization": adv.Organization,
			"reference":    adv.Reference,
			"resource": &bson.M{
				"$in": dismissals,
			},
		}, &bson.M{
			"$set": &bson.M{
				"dismissed": true,
			},
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}

		changed = true
	}

	if len(restores) > 0 {
		_, err = coll.UpdateMany(db, &bson.M{
			"organization": adv.Organization,
			"reference":    adv.Reference,
			"resource": &bson.M{
				"$in": restores,
			},
		}, &bson.M{
			"$set": &bson.M{
				"dismissed": false,
			},
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}

		changed = true
	}

	return
}

func UpdateDismiss(db *database.Database, advId bson.ObjectID,
	dismiss, restore bool, dismissals, restores []bson.ObjectID) (err error) {

	adv, err := Get(db, advId)
	if err != nil {
		return
	}

	changed, err := updateDismiss(
		db, adv, dismiss, restore, dismissals, restores)
	if err != nil {
		return
	}
	if !changed {
		return
	}

	if dismiss || restore {
		err = updateResources(db, adv)
		if err != nil {
			return
		}

		return
	}

	for _, resourceId := range dismissals {
		err = UpdateResource(db, resourceId)
		if err != nil {
			return
		}
	}

	for _, resourceId := range restores {
		err = UpdateResource(db, resourceId)
		if err != nil {
			return
		}
	}

	return
}

func UpdateDismissOrg(db *database.Database, orgId, advId bson.ObjectID,
	dismiss, restore bool, dismissals, restores []bson.ObjectID) (err error) {

	adv, err := GetOrg(db, orgId, advId)
	if err != nil {
		return
	}

	changed, err := updateDismiss(
		db, adv, dismiss, restore, dismissals, restores)
	if err != nil {
		return
	}
	if !changed {
		return
	}

	if dismiss || restore {
		err = updateResources(db, adv)
		if err != nil {
			return
		}

		return
	}

	for _, resourceId := range dismissals {
		err = UpdateResourceOrg(db, resourceId, orgId)
		if err != nil {
			return
		}
	}

	for _, resourceId := range restores {
		err = UpdateResourceOrg(db, resourceId, orgId)
		if err != nil {
			return
		}
	}

	return
}

func updateMulti(db *database.Database, query *bson.M,
	dismiss bool) (err error) {

	(*query)["dismissed"] = &bson.M{
		"$ne": dismiss,
	}

	advs, err := GetAll(db, query, options.Find().SetProjection(&bson.M{
		"organization": 1,
		"reference":    1,
	}))
	if err != nil {
		return
	}

	if len(advs) == 0 {
		return
	}

	advIds := make([]bson.ObjectID, 0, len(advs))
	for _, adv := range advs {
		advIds = append(advIds, adv.Id)
	}

	coll := db.Advisories()

	_, err = coll.UpdateMany(db, &bson.M{
		"_id": &bson.M{
			"$in": advIds,
		},
	}, &bson.M{
		"$set": &bson.M{
			"dismissed": dismiss,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	err = updateResources(db, advs...)
	if err != nil {
		return
	}

	return
}

func UpdateMulti(db *database.Database, advIds []bson.ObjectID,
	dismiss, restore bool) (err error) {

	if !dismiss && !restore {
		return
	}

	err = updateMulti(db, &bson.M{
		"_id": &bson.M{
			"$in": advIds,
		},
	}, dismiss)
	if err != nil {
		return
	}

	return
}

func UpdateMultiOrg(db *database.Database, orgId bson.ObjectID,
	advIds []bson.ObjectID, dismiss, restore bool) (err error) {

	if !dismiss && !restore {
		return
	}

	err = updateMulti(db, &bson.M{
		"_id": &bson.M{
			"$in": advIds,
		},
		"organization": orgId,
	}, dismiss)
	if err != nil {
		return
	}

	return
}

func refreshVulnerability(db *database.Database, adv *Advisory,
	cveId string) (err error) {

	if !slices.Contains(adv.Vulnerabilities, cveId) {
		err = &errortypes.NotFoundError{
			errors.New("advisory: Vulnerability not found in advisory"),
		}
		return
	}

	vuln, err := vulnerability.GetOneForce(db, cveId)
	if err != nil {
		return
	}
	if vuln == nil {
		err = &errortypes.NotFoundError{
			errors.New("advisory: Vulnerability not found"),
		}
		return
	}

	vulns, err := vulnerability.GetMini(db, adv.Vulnerabilities)
	if err != nil {
		return
	}

	prevScore := adv.Score
	prevComplete := adv.Complete
	adv.SetVulnerabilities(vulns)
	adv.UpdateScore()

	coll := db.Advisories()

	_, err = coll.UpdateOne(db, &bson.M{
		"_id": adv.Id,
	}, &bson.M{
		"$set": &bson.M{
			"score":    adv.Score,
			"pending":  adv.Pending,
			"complete": adv.Complete,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if adv.Score != prevScore || adv.Complete != prevComplete {
		err = updateResources(db, adv)
		if err != nil {
			return
		}
	}

	return
}

func RefreshVulnerability(db *database.Database, advId bson.ObjectID,
	cveId string) (err error) {

	adv, err := Get(db, advId)
	if err != nil {
		return
	}

	err = refreshVulnerability(db, adv, cveId)
	if err != nil {
		return
	}

	return
}

func RefreshVulnerabilityOrg(db *database.Database, orgId, advId bson.ObjectID,
	cveId string) (err error) {

	adv, err := GetOrg(db, orgId, advId)
	if err != nil {
		return
	}

	err = refreshVulnerability(db, adv, cveId)
	if err != nil {
		return
	}

	return
}
