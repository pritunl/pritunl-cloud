package firewall

import (
	"fmt"
	"sort"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/nodeport"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/sirupsen/logrus"
)

func Get(db *database.Database, fireId primitive.ObjectID) (
	fire *Firewall, err error) {

	coll := db.Firewalls()
	fire = &Firewall{}

	err = coll.FindOneId(fireId, fire)
	if err != nil {
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, fireId primitive.ObjectID) (
	fire *Firewall, err error) {

	coll := db.Firewalls()
	fire = &Firewall{}

	err = coll.FindOne(db, &bson.M{
		"_id":          fireId,
		"organization": orgId,
	}).Decode(fire)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	fires []*Firewall, err error) {

	coll := db.Firewalls()
	fires = []*Firewall{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		fire := &Firewall{}
		err = cursor.Decode(fire)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		fires = append(fires, fire)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetRoles(db *database.Database, roles []string) (
	fires []*Firewall, err error) {

	coll := db.Firewalls()
	fires = []*Firewall{}

	cursor, err := coll.Find(
		db,
		&bson.M{
			"$or": []*bson.M{
				&bson.M{
					"organization": nil,
				},
				&bson.M{
					"organization": &bson.M{
						"$exists": false,
					},
				},
			},
			"network_roles": &bson.M{
				"$in": roles,
			},
		},
		&options.FindOptions{
			Sort: &bson.D{
				{"_id", 1},
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		fire := &Firewall{}
		err = cursor.Decode(fire)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		fires = append(fires, fire)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOrgMapRoles(db *database.Database, orgId primitive.ObjectID) (
	fires map[string][]*Firewall, err error) {

	coll := db.Firewalls()
	fires = map[string][]*Firewall{}

	cursor, err := coll.Find(db, &bson.M{
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		fire := &Firewall{}
		err = cursor.Decode(fire)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		for _, role := range fire.NetworkRoles {
			roleFires := fires[role]
			if roleFires == nil {
				roleFires = []*Firewall{}
			}
			fires[role] = append(roleFires, fire)
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOrgRoles(db *database.Database, orgId primitive.ObjectID,
	roles []string) (fires []*Firewall, err error) {

	coll := db.Firewalls()
	fires = []*Firewall{}

	cursor, err := coll.Find(
		db,
		&bson.M{
			"organization": orgId,
			"network_roles": &bson.M{
				"$in": roles,
			},
		},
		&options.FindOptions{
			Sort: &bson.D{
				{"_id", 1},
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		fire := &Firewall{}
		err = cursor.Decode(fire)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		fires = append(fires, fire)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (fires []*Firewall, count int64, err error) {

	coll := db.Firewalls()
	fires = []*Firewall{}

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

	maxPage := count / pageCount
	if count == pageCount {
		maxPage = 0
	}
	page = utils.Min64(page, maxPage)
	skip := utils.Min64(page*pageCount, count)

	cursor, err := coll.Find(
		db,
		query,
		&options.FindOptions{
			Sort: &bson.D{
				{"name", 1},
			},
			Skip:  &skip,
			Limit: &pageCount,
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		fire := &Firewall{}
		err = cursor.Decode(fire)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		fires = append(fires, fire)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, fireId primitive.ObjectID) (err error) {
	coll := db.Firewalls()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": fireId,
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

func RemoveOrg(db *database.Database, orgId, fireId primitive.ObjectID) (
	err error) {

	coll := db.Firewalls()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id":          fireId,
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

func RemoveMulti(db *database.Database, fireIds []primitive.ObjectID) (
	err error) {

	coll := db.Firewalls()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": fireIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveMultiOrg(db *database.Database, orgId primitive.ObjectID,
	fireIds []primitive.ObjectID) (err error) {

	coll := db.Firewalls()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": fireIds,
		},
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func MergeIngress(fires []*Firewall) (rules []*Rule) {
	rules = []*Rule{}
	rulesMap := map[string]*Rule{}
	rulesKey := []string{}

	for _, fire := range fires {
		for _, ingress := range fire.Ingress {
			key := fmt.Sprintf("%s-%s", ingress.Protocol, ingress.Port)
			rule := rulesMap[key]
			if rule == nil {
				rule = &Rule{
					Protocol:  ingress.Protocol,
					Port:      ingress.Port,
					SourceIps: ingress.SourceIps,
				}
				rulesMap[key] = rule
				rulesKey = append(rulesKey, key)
			} else {
				sourceIps := set.NewSet()
				for _, sourceIp := range rule.SourceIps {
					sourceIps.Add(sourceIp)
				}

				for _, sourceIp := range ingress.SourceIps {
					if sourceIps.Contains(sourceIp) {
						continue
					}
					sourceIps.Add(sourceIp)
					rule.SourceIps = append(rule.SourceIps, sourceIp)
				}
			}
		}
	}

	sort.Strings(rulesKey)
	for _, key := range rulesKey {
		rules = append(rules, rulesMap[key])
	}

	return
}

func GetAllIngress(db *database.Database, nodeSelf *node.Node,
	instances []*instance.Instance, specRules map[string][]*Rule,
	nodePortsMap map[string][]*nodeport.Mapping) (
	nodeFirewall []*Rule, firewalls map[string][]*Rule,
	mappings map[string][]*Mapping, err error) {

	if nodeSelf.Firewall {
		fires, e := GetRoles(db, nodeSelf.NetworkRoles)
		if e != nil {
			err = e
			return
		}

		ingress := MergeIngress(fires)
		nodeFirewall = ingress
	}

	nodePortIps := map[string]string{}
	firewalls = map[string][]*Rule{}
	for _, inst := range instances {
		if !inst.IsActive() {
			continue
		}

		namespaces := []string{}
		for i := range inst.Virt.NetworkAdapters {
			namespaces = append(namespaces, vm.GetNamespace(inst.Id, i))
		}

		if len(inst.NodePortIps) > 0 && len(namespaces) > 0 {
			nodePortIps[namespaces[0]] = inst.NodePortIps[0]
		}

		fires, e := GetOrgRoles(db,
			inst.Organization, inst.NetworkRoles)
		if e != nil {
			err = e
			return
		}
		ingress := MergeIngress(fires)

		for _, namespace := range namespaces {
			_, ok := firewalls[namespace]
			if ok {
				logrus.WithFields(logrus.Fields{
					"instance_id": inst.Id.Hex(),
					"namespace":   namespace,
				}).Error("firewall: Namespace conflict")

				err = &errortypes.ParseError{
					errors.New("firewall: Namespace conflict"),
				}
				return
			}

			firewalls[namespace] = ingress
		}
	}

	for namespace, rules := range specRules {
		firewalls[namespace] = append(firewalls[namespace], rules...)
	}

	mappings = map[string][]*Mapping{}
	externalPorts := map[int]string{}
	for namespace, ndePorts := range nodePortsMap {
		for _, ndePort := range ndePorts {
			ipvs := false

			extNamespace := externalPorts[ndePort.ExternalPort]
			if extNamespace != "" {
				ipvs = true

				if extNamespace != "-" {
					for _, mapping := range mappings[extNamespace] {
						if mapping.ExternalPort == ndePort.ExternalPort {
							mapping.Ipvs = true
						}
					}
					externalPorts[ndePort.ExternalPort] = "-"
				}
			} else {
				externalPorts[ndePort.ExternalPort] = namespace
			}

			mappings[namespace] = append(mappings[namespace], &Mapping{
				Ipvs:         ipvs,
				Address:      nodePortIps[namespace],
				Protocol:     ndePort.Protocol,
				ExternalPort: ndePort.ExternalPort,
				InternalPort: ndePort.InternalPort,
			})
		}
	}

	return
}
