package aggregate

import (
	"sort"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo"
	"github.com/pritunl/pritunl-cloud/advisory"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/systemd"
)

type AdvisoryPipe struct {
	advisory.Advisory `bson:",inline"`
	InstanceDocs      []*instance.Instance `bson:"instance_docs"`
	NodeDocs          []*node.Node         `bson:"node_docs"`
}

type AdvisoriesPipe struct {
	Metadata   []*Metadata     `bson:"meta"`
	Advisories []*AdvisoryPipe `bson:"advisories"`
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

	if pageCount == 0 {
		pageCount = 20
	}
	skip := page * pageCount

	addAdvisory := func(doc *AdvisoryPipe) {
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

	var cursor *mongo.Cursor
	if len(*query) == 0 {
		waiter := &sync.WaitGroup{}
		var countErr error

		waiter.Add(1)
		go func() {
			defer waiter.Done()

			count, countErr = coll.EstimatedDocumentCount(db)
			if countErr != nil {
				countErr = database.ParseError(countErr)
				return
			}
		}()

		cursor, err = coll.Aggregate(db, []*bson.M{
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
					"from":         "instances",
					"localField":   "instances",
					"foreignField": "_id",
					"as":           "instance_docs",
				},
			},
			&bson.M{
				"$lookup": &bson.M{
					"from":         "nodes",
					"localField":   "nodes",
					"foreignField": "_id",
					"as":           "node_docs",
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

			addAdvisory(doc)
		}

		waiter.Wait()
		if countErr != nil {
			err = countErr
			return
		}
	} else {
		cursor, err = coll.Aggregate(db, []*bson.M{
			&bson.M{
				"$match": query,
			},
			&bson.M{
				"$sort": &bson.M{
					"reference": 1,
				},
			},
			&bson.M{
				"$facet": &bson.M{
					"meta": []*bson.M{
						&bson.M{
							"$count": "count",
						},
					},
					"advisories": []*bson.M{
						&bson.M{
							"$skip": skip,
						},
						&bson.M{
							"$limit": pageCount,
						},
						&bson.M{
							"$lookup": &bson.M{
								"from":         "instances",
								"localField":   "instances",
								"foreignField": "_id",
								"as":           "instance_docs",
							},
						},
						&bson.M{
							"$lookup": &bson.M{
								"from":         "nodes",
								"localField":   "nodes",
								"foreignField": "_id",
								"as":           "node_docs",
							},
						},
					},
				},
			},
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}
		defer cursor.Close(db)

		if !cursor.Next(db) {
			err = &database.NotFoundError{
				errors.New("aggregate: Not found"),
			}
			return
		}

		doc := &AdvisoriesPipe{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		if len(doc.Metadata) > 0 {
			count = doc.Metadata[0].Count
		}

		for _, advDoc := range doc.Advisories {
			addAdvisory(advDoc)
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
