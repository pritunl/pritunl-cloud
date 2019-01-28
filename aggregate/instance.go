package aggregate

import (
	"context"
	"fmt"
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/authority"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
	"sort"
	"strings"
)

type InstancePipe struct {
	instance.Instance `bson:",inline"`
	NodeDocs          []*node.Node `bson:"node_docs"`
	DiskDocs          []*disk.Disk `bson:"disk_docs"`
}

type InstanceInfo struct {
	Node          string   `json:"node"`
	Disks         []string `json:"disks"`
	FirewallRules []string `json:"firewall_rules"`
	Authorities   []string `json:"authorities"`
}

type InstanceAggregate struct {
	instance.Instance
	Info *InstanceInfo `json:"info"`
}

func GetInstancePaged(db *database.Database, query *bson.M, page,
pageCount int64) (insts []*InstanceAggregate, count int64, err error) {

	coll := db.Instances()
	insts = []*InstanceAggregate{}

	count, err = coll.Count(context.Background(), query)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	page = utils.Min64(page, count/pageCount)
	skip := utils.Min64(page*pageCount, count)

	cursor, err := coll.Aggregate(context.Background(), []*bson.M{
		&bson.M{
			"$match": query,
		},
		&bson.M{
			"$sort": &bson.M{
				"name": 1,
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
				"from":         "nodes",
				"localField":   "node",
				"foreignField": "_id",
				"as":           "node_docs",
			},
		},
		&bson.M{
			"$lookup": &bson.M{
				"from":         "disks",
				"localField":   "_id",
				"foreignField": "instance",
				"as":           "disk_docs",
			},
		},
		//&bson.M{
		//	"$lookup": &bson.M{
		//		"from": "firewalls",
		//		"let": &bson.M{
		//			"organization":  "$organization",
		//			"network_roles": "$network_roles",
		//		},
		//		"pipeline": []*bson.M{
		//			&bson.M{
		//				"$match": &bson.M{
		//					"$expr": &bson.M{
		//						"$and": []*bson.M{
		//							&bson.M{
		//								"$eq": []string{
		//									"$organization",
		//									"$$organization",
		//								},
		//							},
		//						},
		//					},
		//				},
		//			},
		//		},
		//		"as": "firewall_docs",
		//	},
		//},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(context.Background())

	firesOrg := map[primitive.ObjectID]map[string][]*firewall.Firewall{}
	firesRoles := map[primitive.ObjectID]set.Set{}
	authrsOrg := map[primitive.ObjectID]map[string][]*authority.Authority{}
	authrsRoles := map[primitive.ObjectID]set.Set{}

	for cursor.Next(context.Background()) {
		doc := &InstancePipe{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		info := &InstanceInfo{
			Node:          "Unknown",
			Disks:         []string{},
			FirewallRules: []string{},
			Authorities:   []string{},
		}

		if len(doc.NodeDocs) > 0 {
			info.Node = doc.NodeDocs[0].Name
		}

		for _, dsk := range doc.DiskDocs {
			info.Disks = append(
				info.Disks,
				fmt.Sprintf("%s: %s", dsk.Index, dsk.Name),
			)
		}

		fires := firesOrg[doc.Organization]
		if fires == nil {
			fires, err = firewall.GetOrgMapRoles(db, doc.Organization)
			if err != nil {
				return
			}

			for _, roleFires := range fires {
				for _, fire := range roleFires {
					if _, ok := firesRoles[fire.Id]; ok {
						continue
					}

					roles := set.NewSet()
					for _, role := range fire.NetworkRoles {
						roles.Add(role)
					}
					firesRoles[fire.Id] = roles
				}
			}

			firesOrg[doc.Organization] = fires
		}

		authrs := authrsOrg[doc.Organization]
		if authrs == nil {
			authrs, err = authority.GetOrgMapRoles(db, doc.Organization)
			if err != nil {
				return
			}

			for _, roleAuthrs := range authrs {
				for _, authr := range roleAuthrs {
					if _, ok := authrsRoles[authr.Id]; ok {
						continue
					}

					roles := set.NewSet()
					for _, role := range authr.NetworkRoles {
						roles.Add(role)
					}
					authrsRoles[authr.Id] = roles
				}
			}

			authrsOrg[doc.Organization] = authrs
		}

		curFires := set.NewSet()

		firewallRules := map[string]set.Set{}
		firewallRulesKeys := []string{}
		authrNames := set.NewSet()
		for _, role := range doc.NetworkRoles {
			roleFires := fires[role]
			if roleFires != nil {
				for _, fire := range roleFires {
					if curFires.Contains(fire.Id) {
						continue
					}
					curFires.Add(fire.Id)

					for _, rule := range fire.Ingress {
						key := rule.Protocol
						if rule.Port != "" {
							key += ":" + rule.Port
						}

						rules := firewallRules[key]
						if rules == nil {
							rules = set.NewSet()
							firewallRules[key] = rules
							firewallRulesKeys = append(
								firewallRulesKeys,
								key,
							)
						}

						for _, sourceIp := range rule.SourceIps {
							rules.Add(sourceIp)
						}
					}
				}
			}

			roleAuthrs := authrs[role]
			if roleAuthrs != nil {
				for _, authr := range roleAuthrs {
					authrNames.Add(authr.Name)
				}
			}
		}

		sort.Strings(firewallRulesKeys)
		for _, key := range firewallRulesKeys {
			rules := firewallRules[key]

			vals := []string{}
			for rule := range rules.Iter() {
				vals = append(vals, rule.(string))
			}
			sort.Strings(vals)

			info.FirewallRules = append(
				info.FirewallRules,
				key+" - "+strings.Join(vals, ", "),
			)
		}

		for authr := range authrNames.Iter() {
			info.Authorities = append(info.Authorities, authr.(string))
		}
		sort.Strings(info.Authorities)

		inst := &InstanceAggregate{
			Instance: doc.Instance,
			Info:     info,
		}

		insts = append(insts, inst)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
