package aggregate

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/organization"
	"github.com/pritunl/pritunl-cloud/plan"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
)

type Completion struct {
	Organizations []*organization.Organization `json:"organizations"`
	Domains       []*domain.Domain             `json:"domains"`
	Vpcs          []*vpc.Vpc                   `json:"vpcs"`
	Datacenters   []*datacenter.Datacenter     `json:"datacenters"`
	Nodes         []*node.Node                 `json:"nodes"`
	Pools         []*pool.Pool                 `json:"pools"`
	Zones         []*zone.Zone                 `json:"zones"`
	Shapes        []*shape.Shape               `json:"shapes"`
	Images        []*image.Image               `json:"images"`
	Instances     []*instance.Instance         `json:"instances"`
	Plans         []*plan.Plan                 `json:"plans"`
	Certificates  []*certificate.Certificate   `json:"certificates"`
	Secrets       []*secret.Secret             `json:"secrets"`
	Pods          []*pod.Pod                   `json:"pods"`
	Units         []*pod.Unit                  `json:"units"`
}

func get(db *database.Database, coll *database.Collection,
	query, projection *bson.M, new func() interface{},
	add func(interface{})) (err error) {

	cursor, err := coll.Find(db, query, &options.FindOptions{
		Projection: projection,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		item := new()
		err = cursor.Decode(item)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		add(item)
	}

	return
}

func GetCompletion(db *database.Database, orgId primitive.ObjectID) (
	cmpl *Completion, err error) {

	cmpl = &Completion{}
	query := &bson.M{}

	err = get(
		db,
		db.Organizations(),
		&bson.M{},
		&bson.M{
			"_id":  1,
			"name": 1,
		},
		func() interface{} {
			return &organization.Organization{}
		},
		func(item interface{}) {
			cmpl.Organizations = append(
				cmpl.Organizations,
				item.(*organization.Organization),
			)
		},
	)
	if err != nil {
		return
	}

	err = get(
		db,
		db.Domains(),
		query,
		&bson.M{
			"_id":          1,
			"name":         1,
			"organization": 1,
		},
		func() interface{} {
			return &domain.Domain{}
		},
		func(item interface{}) {
			cmpl.Domains = append(
				cmpl.Domains,
				item.(*domain.Domain),
			)
		},
	)
	if err != nil {
		return
	}

	err = get(
		db,
		db.Vpcs(),
		query,
		&bson.M{
			"_id":          1,
			"name":         1,
			"organization": 1,
			"vpc_id":       1,
			"network":      1,
			"subnets":      1,
			"datacenter":   1,
		},
		func() interface{} {
			return &vpc.Vpc{}
		},
		func(item interface{}) {
			cmpl.Vpcs = append(
				cmpl.Vpcs,
				item.(*vpc.Vpc),
			)
		},
	)
	if err != nil {
		return
	}

	err = get(
		db,
		db.Datacenters(),
		&bson.M{},
		&bson.M{
			"_id":                 1,
			"name":                1,
			"match_organizations": 1,
			"organizations":       1,
		},
		func() interface{} {
			return &datacenter.Datacenter{}
		},
		func(item interface{}) {
			cmpl.Datacenters = append(
				cmpl.Datacenters,
				item.(*datacenter.Datacenter),
			)
		},
	)
	if err != nil {
		return
	}

	err = get(
		db,
		db.Nodes(),
		&bson.M{},
		&bson.M{
			"_id":              1,
			"name":             1,
			"zone":             1,
			"types":            1,
			"timestamp":        1,
			"cpu_units":        1,
			"memory_units":     1,
			"cpu_units_res":    1,
			"memory_units_res": 1,
		},
		func() interface{} {
			return &node.Node{}
		},
		func(item interface{}) {
			nde := item.(*node.Node)

			if !nde.IsHypervisor() {
				return
			}

			cmpl.Nodes = append(
				cmpl.Nodes,
				nde,
			)
		},
	)
	if err != nil {
		return
	}

	err = get(
		db,
		db.Pools(),
		query,
		&bson.M{
			"_id":          1,
			"name":         1,
			"organization": 1,
			"zone":         1,
		},
		func() interface{} {
			return &pool.Pool{}
		},
		func(item interface{}) {
			cmpl.Pools = append(
				cmpl.Pools,
				item.(*pool.Pool),
			)
		},
	)
	if err != nil {
		return
	}

	err = get(
		db,
		db.Zones(),
		&bson.M{},
		&bson.M{
			"_id":        1,
			"name":       1,
			"datacenter": 1,
		},
		func() interface{} {
			return &zone.Zone{}
		},
		func(item interface{}) {
			cmpl.Zones = append(
				cmpl.Zones,
				item.(*zone.Zone),
			)
		},
	)
	if err != nil {
		return
	}

	err = get(
		db,
		db.Shapes(),
		&bson.M{},
		&bson.M{
			"_id":        1,
			"name":       1,
			"datacenter": 1,
			"flexible":   1,
			"memory":     1,
			"processors": 1,
		},
		func() interface{} {
			return &shape.Shape{}
		},
		func(item interface{}) {
			cmpl.Shapes = append(
				cmpl.Shapes,
				item.(*shape.Shape),
			)
		},
	)
	if err != nil {
		return
	}

	err = get(
		db,
		db.Images(),
		query,
		&bson.M{
			"_id":          1,
			"name":         1,
			"organization": 1,
			"type":         1,
			"firmware":     1,
			"key":          1,
			"storage":      1,
		},
		func() interface{} {
			return &image.Image{}
		},
		func(item interface{}) {
			cmpl.Images = append(
				cmpl.Images,
				item.(*image.Image),
			)
		},
	)
	if err != nil {
		return
	}

	err = get(
		db,
		db.Instances(),
		query,
		&bson.M{
			"_id":          1,
			"name":         1,
			"organization": 1,
			"zone":         1,
			"vpc":          1,
			"subnet":       1,
			"node":         1,
		},
		func() interface{} {
			return &instance.Instance{}
		},
		func(item interface{}) {
			cmpl.Instances = append(
				cmpl.Instances,
				item.(*instance.Instance),
			)
		},
	)
	if err != nil {
		return
	}

	err = get(
		db,
		db.Plans(),
		query,
		&bson.M{
			"_id":          1,
			"name":         1,
			"organization": 1,
		},
		func() interface{} {
			return &plan.Plan{}
		},
		func(item interface{}) {
			cmpl.Plans = append(
				cmpl.Plans,
				item.(*plan.Plan),
			)
		},
	)
	if err != nil {
		return
	}

	err = get(
		db,
		db.Certificates(),
		query,
		&bson.M{
			"_id":          1,
			"name":         1,
			"organization": 1,
		},
		func() interface{} {
			return &certificate.Certificate{}
		},
		func(item interface{}) {
			cmpl.Certificates = append(
				cmpl.Certificates,
				item.(*certificate.Certificate),
			)
		},
	)
	if err != nil {
		return
	}

	err = get(
		db,
		db.Secrets(),
		query,
		&bson.M{
			"_id":          1,
			"name":         1,
			"organization": 1,
		},
		func() interface{} {
			return &secret.Secret{}
		},
		func(item interface{}) {
			cmpl.Secrets = append(
				cmpl.Secrets,
				item.(*secret.Secret),
			)
		},
	)
	if err != nil {
		return
	}

	err = get(
		db,
		db.Pods(),
		query,
		&bson.M{
			"_id":          1,
			"name":         1,
			"organization": 1,
			"units":        1,
		},
		func() interface{} {
			return &pod.Pod{}
		},
		func(item interface{}) {
			pd := item.(*pod.Pod)

			cmpl.Pods = append(
				cmpl.Pods,
				pd,
			)

			for _, unit := range pd.Units {
				cmpl.Units = append(cmpl.Units, unit)
			}
		},
	)
	if err != nil {
		return
	}

	return
}
