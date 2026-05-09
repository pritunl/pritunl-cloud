package aggregate

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo"
	"github.com/pritunl/pritunl-cloud/authority"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/drive"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/iso"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/pci"
	"github.com/pritunl/pritunl-cloud/usb"
	"github.com/pritunl/pritunl-cloud/vm"
)

type InstancePipe struct {
	instance.Instance `bson:",inline"`
	NodeDocs          []*node.Node             `bson:"node_docs"`
	DatacenterDocs    []*datacenter.Datacenter `bson:"datacenter_docs"`
	DiskDocs          []*disk.Disk             `bson:"disk_docs"`
}

type InstancesPipe struct {
	Metadata  []*Metadata     `bson:"meta"`
	Instances []*InstancePipe `bson:"instances"`
}

type InstanceInfo struct {
	Node          string              `json:"node"`
	NodePublicIp  string              `json:"node_public_ip"`
	Mtu           int                 `json:"mtu"`
	Iscsi         bool                `json:"iscsi"`
	Disks         []string            `json:"disks"`
	FirewallRules map[string]string   `json:"firewall_rules"`
	Authorities   []string            `json:"authorities"`
	Isos          []*iso.Iso          `json:"isos"`
	UsbDevices    []*usb.Device       `json:"usb_devices"`
	PciDevices    []*pci.Device       `json:"pci_devices"`
	DriveDevices  []*drive.Device     `json:"drive_devices"`
	CloudSubnets  []*node.CloudSubnet `json:"cloud_subnets"`
}

type InstanceAggregate struct {
	instance.Instance
	Info *InstanceInfo `json:"info"`
}

func GetInstancePaged(db *database.Database, query *bson.M, page,
	pageCount int64) (insts []*InstanceAggregate, count int64, err error) {

	coll := db.Instances()
	insts = []*InstanceAggregate{}

	if pageCount == 0 {
		pageCount = 20
	}
	skip := page * pageCount

	firesOrg := map[bson.ObjectID]map[string][]*firewall.Firewall{}
	firesRoles := map[bson.ObjectID]set.Set{}
	authrsOrg := map[bson.ObjectID]map[string][]*authority.Authority{}
	authrsRoles := map[bson.ObjectID]set.Set{}

	addInstance := func(doc *InstancePipe) error {
		info := &InstanceInfo{
			Node:          "Unknown",
			Disks:         []string{},
			FirewallRules: map[string]string{},
			Authorities:   []string{},
			CloudSubnets:  []*node.CloudSubnet{},
		}

		var nde *node.Node
		var dc *datacenter.Datacenter

		if len(doc.NodeDocs) > 0 {
			nde = doc.NodeDocs[0]
		}

		if len(doc.DatacenterDocs) > 0 {
			dc = doc.DatacenterDocs[0]
		}

		if nde != nil {
			info.Node = nde.Name
			if len(nde.PublicIps) > 0 {
				info.NodePublicIp = nde.PublicIps[0]
			}
			info.Iscsi = nde.Iscsi

			info.Isos = nde.LocalIsos

			info.CloudSubnets = nde.GetCloudSubnetsName()

			if nde.UsbPassthrough {
				info.UsbDevices = nde.UsbDevices
			}

			if nde.PciDevices != nil {
				info.PciDevices = nde.PciDevices
			}

			if nde.InstanceDrives != nil {
				info.DriveDevices = nde.InstanceDrives
			}
		}

		if nde != nil && dc != nil {
			info.Mtu = dc.GetInstanceMtu()
		}

		for _, dsk := range doc.DiskDocs {
			info.Disks = append(
				info.Disks,
				fmt.Sprintf("%s: %s", dsk.Index, dsk.Name),
			)
		}

		fires := firesOrg[doc.Organization]
		if fires == nil {
			var e error
			fires, e = firewall.GetOrgMapRoles(db, doc.Organization)
			if e != nil {
				return e
			}

			for _, roleFires := range fires {
				for _, fire := range roleFires {
					if _, ok := firesRoles[fire.Id]; ok {
						continue
					}

					roles := set.NewSet()
					for _, role := range fire.Roles {
						roles.Add(role)
					}
					firesRoles[fire.Id] = roles
				}
			}

			firesOrg[doc.Organization] = fires
		}

		authrs := authrsOrg[doc.Organization]
		if authrs == nil {
			var e error
			authrs, e = authority.GetOrgMapRoles(db, doc.Organization)
			if e != nil {
				return e
			}

			for _, roleAuthrs := range authrs {
				for _, authr := range roleAuthrs {
					if _, ok := authrsRoles[authr.Id]; ok {
						continue
					}

					roles := set.NewSet()
					for _, role := range authr.Roles {
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
		for _, role := range doc.Roles {
			roleFires := fires[role]
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

			roleAuthrs := authrs[role]
			for _, authr := range roleAuthrs {
				authrNames.Add(authr.Name)
			}
		}

		if !doc.Instance.Deployment.IsZero() {
			doc.Instance.LoadVirt(nil, nil)

			specRules, _, e := firewall.GetSpecRulesSlow(
				db, doc.Instance.Node, []*instance.Instance{&doc.Instance})
			if e != nil {
				return e
			}

			instNamespace := vm.GetNamespace(doc.Instance.Id, 0)
			for namespace, rules := range specRules {
				if namespace != instNamespace {
					continue
				}

				for _, rule := range rules {
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

		sort.Strings(firewallRulesKeys)
		for _, key := range firewallRulesKeys {
			rules := firewallRules[key]

			vals := []string{}
			for rule := range rules.Iter() {
				vals = append(vals, rule.(string))
			}
			sort.Strings(vals)

			info.FirewallRules[key] = strings.Join(vals, ", ")
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
		return nil
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
					"from":         "datacenters",
					"localField":   "datacenter",
					"foreignField": "_id",
					"as":           "datacenter_docs",
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
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}
		defer cursor.Close(db)

		for cursor.Next(db) {
			doc := &InstancePipe{}
			err = cursor.Decode(doc)
			if err != nil {
				err = database.ParseError(err)
				return
			}

			err = addInstance(doc)
			if err != nil {
				return
			}
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
					"name": 1,
				},
			},
			&bson.M{
				"$facet": &bson.M{
					"meta": []*bson.M{
						&bson.M{
							"$count": "count",
						},
					},
					"instances": []*bson.M{
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
								"from":         "datacenters",
								"localField":   "datacenter",
								"foreignField": "_id",
								"as":           "datacenter_docs",
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

		doc := &InstancesPipe{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		if len(doc.Metadata) > 0 {
			count = doc.Metadata[0].Count
		}

		for _, instDoc := range doc.Instances {
			err = addInstance(instDoc)
			if err != nil {
				return
			}
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
