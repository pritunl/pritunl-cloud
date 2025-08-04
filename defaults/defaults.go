package defaults

import (
	"fmt"
	"math/rand"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/authority"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/organization"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
	"github.com/sirupsen/logrus"
)

func initStorage(db *database.Database) (err error) {
	stores, err := storage.GetAll(db)
	if err != nil {
		return
	}

	if len(stores) == 0 {
		store := &storage.Storage{
			Name:     "pritunl-images",
			Type:     storage.Public,
			Endpoint: "images.pritunl.com",
			Bucket:   "stable",
			Insecure: false,
		}

		errData, e := store.Validate(db)
		if e != nil {
			err = e
			return
		}

		if errData != nil {
			err = &errortypes.ApiError{
				errors.Newf(
					"defaults: Storage validate error %s",
					errData.Message,
				),
			}
			return
		}

		err = store.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"storage": store.Id.Hex(),
		}).Info("defaults: Created default storage")

		event.PublishDispatch(db, "storage.change")
	}

	return
}

func initOrganization(db *database.Database) (
	defaultOrg primitive.ObjectID, err error) {

	orgs, err := organization.GetAll(db)
	if err != nil {
		return
	}

	if len(orgs) == 0 {
		org := &organization.Organization{
			Name:  "org",
			Roles: []string{"org"},
		}

		errData, e := org.Validate(db)
		if e != nil {
			err = e
			return
		}

		if errData != nil {
			err = &errortypes.ApiError{
				errors.Newf(
					"defaults: Organization validate error %s",
					errData.Message,
				),
			}
			return
		}

		err = org.Insert(db)
		if err != nil {
			return
		}

		defaultOrg = org.Id

		logrus.WithFields(logrus.Fields{
			"organization": org.Id.Hex(),
		}).Info("defaults: Created default organization")

		event.PublishDispatch(db, "organization.change")
	} else {
		for _, org := range orgs {
			if defaultOrg.IsZero() || org.Name == "org" {
				defaultOrg = org.Id
			}
		}
	}

	return
}

func initDatacenter(db *database.Database) (
	defaultDc primitive.ObjectID, err error) {

	dcs, err := datacenter.GetAll(db)
	if err != nil {
		return
	}

	if len(dcs) == 0 {
		stores, e := storage.GetAll(db)
		if e != nil {
			err = e
			return
		}

		publicStorages := []primitive.ObjectID{}
		for _, store := range stores {
			if store.Endpoint == "images.pritunl.com" &&
				store.Bucket == "stable" {

				publicStorages = append(publicStorages, store.Id)
				break
			}
		}

		dc := &datacenter.Datacenter{
			Name:           "us-west-1",
			NetworkMode:    datacenter.Default,
			PublicStorages: publicStorages,
		}

		errData, e := dc.Validate(db)
		if e != nil {
			err = e
			return
		}

		if errData != nil {
			err = &errortypes.ApiError{
				errors.Newf(
					"defaults: Datacenter validate error %s",
					errData.Message,
				),
			}
			return
		}

		err = dc.Insert(db)
		if err != nil {
			return
		}

		defaultDc = dc.Id

		logrus.WithFields(logrus.Fields{
			"datacenter": dc.Id.Hex(),
		}).Info("defaults: Created default datacenter")

		event.PublishDispatch(db, "datacenter.change")
	} else {
		for _, dc := range dcs {
			if defaultDc.IsZero() || dc.Name == "us-west-1" {
				defaultDc = dc.Id
			}
		}
	}

	return
}

func initZone(db *database.Database, defaultDc primitive.ObjectID) (
	err error) {

	zones, err := zone.GetAll(db)
	if err != nil {
		return
	}

	if len(zones) == 0 {
		zne := &zone.Zone{
			Name:       "us-west-1a",
			Datacenter: defaultDc,
		}

		errData, e := zne.Validate(db)
		if e != nil {
			err = e
			return
		}

		if errData != nil {
			err = &errortypes.ApiError{
				errors.Newf(
					"defaults: Zone validate error %s",
					errData.Message,
				),
			}
			return
		}

		err = zne.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"zone": zne.Id.Hex(),
		}).Info("defaults: Created default zone")

		event.PublishDispatch(db, "zone.change")
	}

	return
}

func initVpc(db *database.Database, defaultOrg,
	defaultDc primitive.ObjectID) (err error) {

	if defaultOrg.IsZero() {
		return
	}

	vcs, err := vpc.GetAll(db, &bson.M{})
	if err != nil {
		return
	}

	if len(vcs) == 0 {
		start, end, step := 100, 220, 4
		randomStep := rand.Intn((end-start)/step + 1)
		netNum := start + (randomStep * step)

		vc := &vpc.Vpc{
			Name:         "vpc",
			Organization: defaultOrg,
			Datacenter:   defaultDc,
			VpcId:        utils.RandInt(1000, 4090),
			Network:      fmt.Sprintf("10.%d.0.0/14", netNum),
			Subnets: []*vpc.Subnet{
				&vpc.Subnet{
					Name:    "primary",
					Network: fmt.Sprintf("10.%d.2.0/23", netNum),
				},
			},
		}

		vc.InitVpc()

		errData, e := vc.Validate(db)
		if e != nil {
			err = e
			return
		}

		if errData != nil {
			err = &errortypes.ApiError{
				errors.Newf(
					"defaults: VPC validate error %s",
					errData.Message,
				),
			}
			return
		}

		err = vc.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"vpc": vc.Id.Hex(),
		}).Info("defaults: Created default VPC")

		event.PublishDispatch(db, "vpc.change")
	}

	return
}

func initFirewall(db *database.Database, defaultOrg primitive.ObjectID) (
	err error) {

	if defaultOrg.IsZero() {
		return
	}

	fires, err := firewall.GetAll(db, &bson.M{})
	if err != nil {
		return
	}

	if len(fires) == 0 {
		fire := &firewall.Firewall{
			Name:         "instance",
			Organization: defaultOrg,
			Roles: []string{
				"instance",
			},
			Ingress: []*firewall.Rule{
				&firewall.Rule{
					SourceIps: []string{
						"0.0.0.0/0",
						"::/0",
					},
					Protocol: firewall.Icmp,
				},
				&firewall.Rule{
					SourceIps: []string{
						"0.0.0.0/0",
						"::/0",
					},
					Protocol: firewall.Tcp,
					Port:     "22",
				},
			},
		}

		errData, e := fire.Validate(db)
		if e != nil {
			err = e
			return
		}

		if errData != nil {
			err = &errortypes.ApiError{
				errors.Newf(
					"defaults: Firewall validate error %s",
					errData.Message,
				),
			}
			return
		}

		err = fire.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"firewall": fire.Id.Hex(),
		}).Info("defaults: Created default firewall")

		event.PublishDispatch(db, "firewall.change")
	}

	return
}

func initAuthority(db *database.Database, defaultOrg primitive.ObjectID) (
	err error) {

	if defaultOrg.IsZero() {
		return
	}

	authrs, err := authority.GetAll(db, &bson.M{})
	if err != nil {
		return
	}

	if len(authrs) == 0 {
		authr := &authority.Authority{
			Name:         "cloud",
			Type:         authority.SshKey,
			Organization: defaultOrg,
			NetworkRoles: []string{
				"instance",
			},
		}

		errData, e := authr.Validate(db)
		if e != nil {
			err = e
			return
		}

		if errData != nil {
			err = &errortypes.ApiError{
				errors.Newf(
					"defaults: Authority validate error %s",
					errData.Message,
				),
			}
			return
		}

		err = authr.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"authority": authr.Id.Hex(),
		}).Info("defaults: Created default authority")

		event.PublishDispatch(db, "authority.change")
	}

	return
}

func initNode(db *database.Database, defaultOrg primitive.ObjectID) (
	err error) {

	if defaultOrg.IsZero() {
		return
	}

	if !node.Self.Zone.IsZero() {
		return
	}

	dcs, err := datacenter.GetAll(db)
	if err != nil {
		return
	}

	zones, err := zone.GetAll(db)
	if err != nil {
		return
	}

	nodes, err := node.GetAll(db)
	if err != nil {
		return
	}

	if len(dcs) != 1 || len(zones) != 1 || len(nodes) != 1 {
		return
	}

	node.Self.Datacenter = zones[0].Datacenter
	node.Self.Zone = zones[0].Id
	node.Self.Roles = []string{
		"shape-m2",
	}
	node.Self.HostNat = true
	node.Self.InternalInterfaces = []string{
		settings.Hypervisor.HostNetworkName,
	}

	errData, err := node.Self.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = errData.GetError()
		return
	}

	err = node.Self.Commit(db)
	if err != nil {
		return
	}

	shpe := &shape.Shape{
		Name:       "m2-small",
		Datacenter: node.Self.Datacenter,
		Memory:     2048,
		Processors: 1,
		Flexible:   true,
		Roles: []string{
			"shape-m2",
		},
		Type:     shape.Instance,
		DiskType: shape.Qcow2,
	}

	errData, err = shpe.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = errData.GetError()
		return
	}

	err = shpe.Insert(db)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"node": node.Self.Id.Hex(),
	}).Info("defaults: Configured default node")

	event.PublishDispatch(db, "node.change")

	return
}

func Defaults() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	err = initStorage(db)
	if err != nil {
		return
	}

	defaultOrg, err := initOrganization(db)
	if err != nil {
		return
	}

	defaultDc, err := initDatacenter(db)
	if err != nil {
		return
	}

	err = initZone(db, defaultDc)
	if err != nil {
		return
	}

	err = initVpc(db, defaultOrg, defaultDc)
	if err != nil {
		return
	}

	err = initFirewall(db, defaultOrg)
	if err != nil {
		return
	}

	err = initAuthority(db, defaultOrg)
	if err != nil {
		return
	}

	err = initNode(db, defaultOrg)
	if err != nil {
		return
	}

	return
}
