package aggregate

import (
	"sort"
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/advisory"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/utils"
)

type AdvisoryPipe struct {
	advisory.Advisory `bson:",inline"`
	InstanceDocs      []*instance.Instance `bson:"instance_docs"`
	NodeDocs          []*node.Node         `bson:"node_docs"`
}

type AdvisoryInstanceInfo struct {
	Id              bson.ObjectID `json:"id"`
	Name            string        `json:"name"`
	Action          string        `json:"action"`
	State           string        `json:"state"`
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
	Timestamp  time.Time     `json:"timestamp"`
	PublicIps  []string      `json:"public_ips"`
	PublicIps6 []string      `json:"public_ips6"`
	PrivateIps []string      `json:"private_ips"`
}

type AdvisoryAggregate struct {
	advisory.Advisory
	InstancesInfo []*AdvisoryInstanceInfo `json:"instances_info"`
	NodesInfo     []*AdvisoryNodeInfo     `json:"nodes_info"`
}

func GetAdvisoryPaged(db *database.Database, query *bson.M, page,
	pageCount int64) (advisories []*AdvisoryAggregate, count int64,
	err error) {

	coll := db.Advisories()
	advisories = []*AdvisoryAggregate{}

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

	cursor, err := coll.Aggregate(db, []*bson.M{
		&bson.M{
			"$match": query,
		},
		&bson.M{
			"$sort": &bson.M{
				"reference": 1,
			},
		},
		&bson.M{
			"$skip": skip,
		},
		&bson.M{
			"$limit": pageCount,
		},
		&bson.M{
			"$lookup": &bson.M{
				"from": "instances",
				"let": &bson.M{
					"instance_ids": "$instances",
				},
				"pipeline": []*bson.M{
					&bson.M{
						"$match": &bson.M{
							"$expr": &bson.M{
								"$in": bson.A{"$_id", "$$instance_ids"},
							},
						},
					},
					&bson.M{
						"$project": &bson.M{
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
						},
					},
				},
				"as": "instance_docs",
			},
		},
		&bson.M{
			"$lookup": &bson.M{
				"from": "nodes",
				"let": &bson.M{
					"node_ids": "$nodes",
				},
				"pipeline": []*bson.M{
					&bson.M{
						"$match": &bson.M{
							"$expr": &bson.M{
								"$in": bson.A{"$_id", "$$node_ids"},
							},
						},
					},
					&bson.M{
						"$project": &bson.M{
							"name":        1,
							"timestamp":   1,
							"public_ips":  1,
							"public_ips6": 1,
							"private_ips": 1,
						},
					},
				},
				"as": "node_docs",
			},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		doc := &AdvisoryPipe{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		instancesInfo := []*AdvisoryInstanceInfo{}
		for _, inst := range doc.InstanceDocs {
			uptime := ""
			if !inst.Timestamp.IsZero() && inst.IsActive() {
				uptime = systemd.FormatUptime(inst.Timestamp)
			}

			instancesInfo = append(instancesInfo, &AdvisoryInstanceInfo{
				Id:              inst.Id,
				Name:            inst.Name,
				Action:          inst.Action,
				State:           inst.State,
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

		nodesInfo := []*AdvisoryNodeInfo{}
		for _, nde := range doc.NodeDocs {
			privateIps := []string{}
			for _, privateIp := range nde.PrivateIps {
				privateIps = append(privateIps, privateIp)
			}
			sort.Strings(privateIps)

			nodesInfo = append(nodesInfo, &AdvisoryNodeInfo{
				Id:         nde.Id,
				Name:       nde.Name,
				Timestamp:  nde.Timestamp,
				PublicIps:  nde.PublicIps,
				PublicIps6: nde.PublicIps6,
				PrivateIps: privateIps,
			})
		}

		adv := &AdvisoryAggregate{
			Advisory:      doc.Advisory,
			InstancesInfo: instancesInfo,
			NodesInfo:     nodesInfo,
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
