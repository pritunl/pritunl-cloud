package aggregate

import (
	"sort"
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/advisory"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vulnerability"
)

type AdvisoryInstanceInfo struct {
	Id              bson.ObjectID `json:"id"`
	Name            string        `json:"name"`
	State           string        `json:"state"`
	Dismissed       bool          `json:"dismissed"`
	Status          string        `json:"status"`
	Timestamp       time.Time     `json:"timestamp"`
	Uptime          string        `json:"uptime"`
	PublicIps       []string      `json:"public_ips"`
	PublicIps6      []string      `json:"public_ips6"`
	PrivateIps      []string      `json:"private_ips"`
	PrivateIps6     []string      `json:"private_ips6"`
	CloudPublicIps  []string      `json:"cloud_public_ips"`
	CloudPublicIps6 []string      `json:"cloud_public_ips6"`
}

type AdvisoryNodeInfo struct {
	Id         bson.ObjectID `json:"id"`
	Name       string        `json:"name"`
	State      string        `json:"state"`
	Dismissed  bool          `json:"dismissed"`
	Timestamp  time.Time     `json:"timestamp"`
	PublicIps  []string      `json:"public_ips"`
	PublicIps6 []string      `json:"public_ips6"`
	PrivateIps []string      `json:"private_ips"`
}

type AdvisoryCounts struct {
	Nodes                int64 `json:"nodes"`
	UnreachableNodes     int64 `json:"unreachable_nodes"`
	DismissedNodes       int64 `json:"dismissed_nodes"`
	Instances            int64 `json:"instances"`
	UnreachableInstances int64 `json:"unreachable_instances"`
	DismissedInstances   int64 `json:"dismissed_instances"`
	Active               int64 `json:"active"`
}

type AdvisoryDetail struct {
	Id              bson.ObjectID                  `json:"id"`
	Vulnerabilities []*vulnerability.Vulnerability `json:"vulnerabilities"`
	InstancesInfo   []*AdvisoryInstanceInfo        `json:"instances_info"`
	NodesInfo       []*AdvisoryNodeInfo            `json:"nodes_info"`
	Counts          *AdvisoryCounts                `json:"counts"`
	Exclusions      map[string]int64               `json:"exclusions"`
	Page            int64                          `json:"page"`
	PageCount       int64                          `json:"page_count"`
	Count           int64                          `json:"count"`
}

type ResourceAdvisory struct {
	advisory.Advisory
	VulnerabilityDocs []*vulnerability.Vulnerability `json:"vulnerability_docs"`
	Resource          *advisory.Resource             `json:"resource"`
}

type advisoryCountDoc struct {
	Id struct {
		Kind      string `bson:"kind"`
		State     string `bson:"state"`
		Dismissed bool   `bson:"dismissed"`
	} `bson:"_id"`
	Count int64 `bson:"count"`
}

type advisoryExclusionDoc struct {
	Id    string `bson:"_id"`
	Count int64  `bson:"count"`
}

func getAdvisoryCounts(db *database.Database, query *bson.M) (
	counts *AdvisoryCounts, err error) {

	coll := db.AdvisoryResources()
	counts = &AdvisoryCounts{}

	cursor, err := coll.Aggregate(db, []*bson.M{
		&bson.M{
			"$match": query,
		},
		&bson.M{
			"$group": &bson.M{
				"_id": &bson.M{
					"kind":      "$kind",
					"state":     "$state",
					"dismissed": "$dismissed",
				},
				"count": &bson.M{
					"$sum": 1,
				},
			},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		doc := &advisoryCountDoc{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		isNode := doc.Id.Kind == advisory.Node

		if !doc.Id.Dismissed {
			counts.Active += doc.Count
		}

		switch {
		case doc.Id.Dismissed && isNode:
			counts.DismissedNodes += doc.Count
		case doc.Id.Dismissed:
			counts.DismissedInstances += doc.Count
		case doc.Id.State == advisory.Unreachable && isNode:
			counts.UnreachableNodes += doc.Count
		case doc.Id.State == advisory.Unreachable:
			counts.UnreachableInstances += doc.Count
		case isNode:
			counts.Nodes += doc.Count
		default:
			counts.Instances += doc.Count
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func getAdvisoryExclusions(db *database.Database, query *bson.M) (
	exclusions map[string]int64, err error) {

	coll := db.AdvisoryResources()
	exclusions = map[string]int64{}

	cursor, err := coll.Aggregate(db, []*bson.M{
		&bson.M{
			"$match": query,
		},
		&bson.M{
			"$match": &bson.M{
				"dismissed": &bson.M{
					"$ne": true,
				},
			},
		},
		&bson.M{
			"$unwind": "$exclusions",
		},
		&bson.M{
			"$group": &bson.M{
				"_id": "$exclusions",
				"count": &bson.M{
					"$sum": 1,
				},
			},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		doc := &advisoryExclusionDoc{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		exclusions[doc.Id] = doc.Count
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func getAdvisoryInstances(db *database.Database,
	resources []*advisory.Resource) (
	infos []*AdvisoryInstanceInfo, err error) {

	infos = []*AdvisoryInstanceInfo{}

	instIds := []bson.ObjectID{}
	for _, res := range resources {
		if res.Kind != advisory.Node {
			instIds = append(instIds, res.Resource)
		}
	}

	if len(instIds) == 0 {
		return
	}

	coll := db.Instances()

	cursor, err := coll.Find(db, &bson.M{
		"_id": &bson.M{
			"$in": instIds,
		},
	}, options.Find().SetProjection(&bson.M{
		"name":              1,
		"action":            1,
		"state":             1,
		"timestamp":         1,
		"public_ips":        1,
		"public_ips6":       1,
		"private_ips":       1,
		"private_ips6":      1,
		"cloud_public_ips":  1,
		"cloud_public_ips6": 1,
	}))
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	insts := map[bson.ObjectID]*instance.Instance{}
	for cursor.Next(db) {
		inst := &instance.Instance{}
		err = cursor.Decode(inst)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		insts[inst.Id] = inst
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	for _, res := range resources {
		inst := insts[res.Resource]
		if inst == nil || res.Kind == advisory.Node {
			continue
		}

		inst.Json(true)

		uptime := ""
		if !inst.Timestamp.IsZero() && inst.IsActive() {
			uptime = systemd.FormatUptime(inst.Timestamp)
		}

		infos = append(infos, &AdvisoryInstanceInfo{
			Id:              inst.Id,
			Name:            inst.Name,
			State:           res.State,
			Dismissed:       res.Dismissed,
			Status:          inst.Status,
			Timestamp:       inst.Timestamp,
			Uptime:          uptime,
			PublicIps:       inst.PublicIps,
			PublicIps6:      inst.PublicIps6,
			PrivateIps:      inst.PrivateIps,
			PrivateIps6:     inst.PrivateIps6,
			CloudPublicIps:  inst.CloudPublicIps,
			CloudPublicIps6: inst.CloudPublicIps6,
		})
	}

	return
}

func getAdvisoryNodes(db *database.Database,
	resources []*advisory.Resource) (infos []*AdvisoryNodeInfo, err error) {

	infos = []*AdvisoryNodeInfo{}

	nodeIds := []bson.ObjectID{}
	for _, res := range resources {
		if res.Kind == advisory.Node {
			nodeIds = append(nodeIds, res.Resource)
		}
	}

	if len(nodeIds) == 0 {
		return
	}

	coll := db.Nodes()

	cursor, err := coll.Find(db, &bson.M{
		"_id": &bson.M{
			"$in": nodeIds,
		},
	}, options.Find().SetProjection(&bson.M{
		"name":        1,
		"timestamp":   1,
		"public_ips":  1,
		"public_ips6": 1,
		"private_ips": 1,
	}))
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	ndes := map[bson.ObjectID]*node.Node{}
	for cursor.Next(db) {
		nde := &node.Node{}
		err = cursor.Decode(nde)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		ndes[nde.Id] = nde
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	for _, res := range resources {
		nde := ndes[res.Resource]
		if nde == nil || res.Kind != advisory.Node {
			continue
		}

		privateIps := []string{}
		for _, privateIp := range nde.PrivateIps {
			privateIps = append(privateIps, privateIp)
		}
		sort.Strings(privateIps)

		infos = append(infos, &AdvisoryNodeInfo{
			Id:         nde.Id,
			Name:       nde.Name,
			State:      res.State,
			Dismissed:  res.Dismissed,
			Timestamp:  nde.Timestamp,
			PublicIps:  nde.PublicIps,
			PublicIps6: nde.PublicIps6,
			PrivateIps: privateIps,
		})
	}

	return
}

func GetAdvisoryDetail(db *database.Database, adv *advisory.Advisory,
	page, pageCount int64) (detail *AdvisoryDetail, err error) {

	query := &bson.M{
		"organization": adv.Organization,
		"reference":    adv.Reference,
		"timestamp": &bson.M{
			"$gte": adv.Updated,
		},
	}

	vulns, err := vulnerability.GetMulti(db, adv.Vulnerabilities)
	if err != nil {
		return
	}

	counts, err := getAdvisoryCounts(db, query)
	if err != nil {
		return
	}

	exclusions, err := getAdvisoryExclusions(db, query)
	if err != nil {
		return
	}

	count := counts.Active + counts.DismissedNodes +
		counts.DismissedInstances

	if pageCount <= 0 {
		pageCount = 20
	}
	maxPage := count / pageCount
	if count > 0 && count%pageCount == 0 {
		maxPage -= 1
	}
	page = utils.Min64(utils.Max64(page, 0), maxPage)

	resources, err := advisory.GetResources(
		db,
		query,
		options.Find().
			SetSort(&bson.D{
				{"kind", -1},
				{"dismissed", 1},
				{"state", 1},
				{"resource", 1},
			}).
			SetSkip(page*pageCount).
			SetLimit(pageCount),
	)
	if err != nil {
		return
	}

	instancesInfo, err := getAdvisoryInstances(db, resources)
	if err != nil {
		return
	}

	nodesInfo, err := getAdvisoryNodes(db, resources)
	if err != nil {
		return
	}

	detail = &AdvisoryDetail{
		Id:              adv.Id,
		Vulnerabilities: vulns,
		InstancesInfo:   instancesInfo,
		NodesInfo:       nodesInfo,
		Counts:          counts,
		Exclusions:      exclusions,
		Page:            page,
		PageCount:       pageCount,
		Count:           count,
	}

	return
}

func GetResourceAdvisories(db *database.Database,
	resourceId bson.ObjectID) (advisories []*ResourceAdvisory, err error) {

	advisories = []*ResourceAdvisory{}

	advs, resources, err := advisory.GetResourceAdvisories(db, resourceId)
	if err != nil {
		return
	}

	cveIds := []string{}
	for _, adv := range advs {
		cveIds = append(cveIds, adv.Vulnerabilities...)
	}

	vulns, err := vulnerability.GetMulti(db, cveIds)
	if err != nil {
		return
	}

	vulnsMap := map[string]*vulnerability.Vulnerability{}
	for _, vuln := range vulns {
		vulnsMap[vuln.Id] = vuln
	}

	for _, adv := range advs {
		vulnDocs := []*vulnerability.Vulnerability{}
		for _, cveId := range adv.Vulnerabilities {
			vuln := vulnsMap[cveId]
			if vuln != nil {
				vulnDocs = append(vulnDocs, vuln)
			}
		}

		advisories = append(advisories, &ResourceAdvisory{
			Advisory:          *adv,
			VulnerabilityDocs: vulnDocs,
			Resource:          resources[adv.Id],
		})
	}

	return
}
