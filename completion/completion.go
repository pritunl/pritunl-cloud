package completion

import (
	"sort"
	"sync"
	"time"

	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/deployment"
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
	"github.com/pritunl/pritunl-cloud/unit"
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
	Builds        []*Build                     `json:"builds"`
	Instances     []*instance.Instance         `json:"instances"`
	Plans         []*plan.Plan                 `json:"plans"`
	Certificates  []*certificate.Certificate   `json:"certificates"`
	Secrets       []*secret.Secret             `json:"secrets"`
	Pods          []*pod.Pod                   `json:"pods"`
	Units         []*unit.Unit                 `json:"units"`
}

type Build struct {
	Id           primitive.ObjectID `json:"id"`
	Name         string             `json:"name"`
	Pod          primitive.ObjectID `json:"pod"`
	Organization primitive.ObjectID `json:"organization"`
	Tags         []*BuildTag        `json:"tags"`
}

type BuildTag struct {
	Tag       string    `json:"tag"`
	Timestamp time.Time `json:"timestamp"`
}

func get(db *database.Database, coll *database.Collection,
	query bson.M, projection *bson.M, sort *bson.M, new func() interface{},
	add func(interface{})) (err error) {

	opts := &options.FindOptions{
		Projection: projection,
	}
	if sort != nil {
		opts.Sort = sort
	}

	cursor, err := coll.Find(db, query, opts)
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
	query := bson.M{}
	if !orgId.IsZero() {
		query["organization"] = orgId
	}

	releaseImages := map[string][]*image.Image{}
	otherImages := []*image.Image{}
	unitsMap := map[primitive.ObjectID]*unit.Unit{}
	buildsMap := map[primitive.ObjectID]*Build{}
	deployments := []*deployment.Deployment{}

	var wg sync.WaitGroup
	errChan := make(chan error, 16)

	wg.Add(1)
	go func() {
		defer wg.Done()

		var orgs []*organization.Organization
		err := get(
			db,
			db.Organizations(),
			bson.M{},
			&bson.M{
				"_id":  1,
				"name": 1,
			},
			nil,
			func() interface{} {
				return &organization.Organization{}
			},
			func(item interface{}) {
				orgs = append(
					orgs,
					item.(*organization.Organization),
				)
			},
		)
		if err != nil {
			errChan <- err
			return
		}
		cmpl.Organizations = orgs
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		var domains []*domain.Domain
		err := get(
			db,
			db.Domains(),
			query,
			&bson.M{
				"_id":          1,
				"name":         1,
				"organization": 1,
			},
			nil,
			func() interface{} {
				return &domain.Domain{}
			},
			func(item interface{}) {
				domains = append(
					domains,
					item.(*domain.Domain),
				)
			},
		)
		if err != nil {
			errChan <- err
			return
		}
		cmpl.Domains = domains
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		var vpcs []*vpc.Vpc
		err := get(
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
			nil,
			func() interface{} {
				return &vpc.Vpc{}
			},
			func(item interface{}) {
				vpcs = append(
					vpcs,
					item.(*vpc.Vpc),
				)
			},
		)
		if err != nil {
			errChan <- err
			return
		}
		cmpl.Vpcs = vpcs
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		var datacenters []*datacenter.Datacenter
		err := get(
			db,
			db.Datacenters(),
			bson.M{},
			&bson.M{
				"_id":                 1,
				"name":                1,
				"match_organizations": 1,
				"organizations":       1,
			},
			nil,
			func() interface{} {
				return &datacenter.Datacenter{}
			},
			func(item interface{}) {
				datacenters = append(
					datacenters,
					item.(*datacenter.Datacenter),
				)
			},
		)
		if err != nil {
			errChan <- err
			return
		}
		cmpl.Datacenters = datacenters
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		var nodes []*node.Node
		err := get(
			db,
			db.Nodes(),
			bson.M{},
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
			nil,
			func() interface{} {
				return &node.Node{}
			},
			func(item interface{}) {
				nde := item.(*node.Node)

				if !nde.IsHypervisor() {
					return
				}

				nodes = append(
					nodes,
					nde,
				)
			},
		)
		if err != nil {
			errChan <- err
			return
		}
		cmpl.Nodes = nodes
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		var pools []*pool.Pool
		err := get(
			db,
			db.Pools(),
			query,
			&bson.M{
				"_id":          1,
				"name":         1,
				"organization": 1,
				"zone":         1,
			},
			nil,
			func() interface{} {
				return &pool.Pool{}
			},
			func(item interface{}) {
				pools = append(
					pools,
					item.(*pool.Pool),
				)
			},
		)
		if err != nil {
			errChan <- err
			return
		}
		cmpl.Pools = pools
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		var zones []*zone.Zone
		err := get(
			db,
			db.Zones(),
			bson.M{},
			&bson.M{
				"_id":        1,
				"name":       1,
				"datacenter": 1,
			},
			nil,
			func() interface{} {
				return &zone.Zone{}
			},
			func(item interface{}) {
				zones = append(
					zones,
					item.(*zone.Zone),
				)
			},
		)
		if err != nil {
			errChan <- err
			return
		}
		cmpl.Zones = zones
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		var shapes []*shape.Shape
		err := get(
			db,
			db.Shapes(),
			bson.M{},
			&bson.M{
				"_id":        1,
				"name":       1,
				"datacenter": 1,
				"flexible":   1,
				"memory":     1,
				"processors": 1,
			},
			nil,
			func() interface{} {
				return &shape.Shape{}
			},
			func(item interface{}) {
				shapes = append(
					shapes,
					item.(*shape.Shape),
				)
			},
		)
		if err != nil {
			errChan <- err
			return
		}
		cmpl.Shapes = shapes
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := get(
			db,
			db.Images(),
			query,
			&bson.M{
				"_id":          1,
				"name":         1,
				"release":      1,
				"build":        1,
				"organization": 1,
				"deployment":   1,
				"type":         1,
				"firmware":     1,
				"key":          1,
				"storage":      1,
			},
			nil,
			func() interface{} {
				return &image.Image{}
			},
			func(item interface{}) {
				img := item.(*image.Image)
				if img.Release != "" {
					releaseImages[img.Release] = append(
						releaseImages[img.Release],
						img,
					)
				} else {
					otherImages = append(otherImages, img)
				}
			},
		)
		if err != nil {
			errChan <- err
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		var instances []*instance.Instance
		err := get(
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
			nil,
			func() interface{} {
				return &instance.Instance{}
			},
			func(item interface{}) {
				instances = append(
					instances,
					item.(*instance.Instance),
				)
			},
		)
		if err != nil {
			errChan <- err
			return
		}
		cmpl.Instances = instances
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		var plans []*plan.Plan
		err := get(
			db,
			db.Plans(),
			query,
			&bson.M{
				"_id":          1,
				"name":         1,
				"organization": 1,
			},
			nil,
			func() interface{} {
				return &plan.Plan{}
			},
			func(item interface{}) {
				plans = append(
					plans,
					item.(*plan.Plan),
				)
			},
		)
		if err != nil {
			errChan <- err
			return
		}
		cmpl.Plans = plans
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		var certificates []*certificate.Certificate
		err := get(
			db,
			db.Certificates(),
			query,
			&bson.M{
				"_id":          1,
				"name":         1,
				"organization": 1,
				"type":         1,
			},
			nil,
			func() interface{} {
				return &certificate.Certificate{}
			},
			func(item interface{}) {
				certificates = append(
					certificates,
					item.(*certificate.Certificate),
				)
			},
		)
		if err != nil {
			errChan <- err
			return
		}
		cmpl.Certificates = certificates
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		var secrets []*secret.Secret
		err := get(
			db,
			db.Secrets(),
			query,
			&bson.M{
				"_id":          1,
				"name":         1,
				"organization": 1,
				"type":         1,
			},
			nil,
			func() interface{} {
				return &secret.Secret{}
			},
			func(item interface{}) {
				secrets = append(
					secrets,
					item.(*secret.Secret),
				)
			},
		)
		if err != nil {
			errChan <- err
			return
		}
		cmpl.Secrets = secrets
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		var pods []*pod.Pod
		err := get(
			db,
			db.Pods(),
			query,
			&bson.M{
				"_id":          1,
				"name":         1,
				"organization": 1,
			},
			nil,
			func() interface{} {
				return &pod.Pod{}
			},
			func(item interface{}) {
				pods = append(
					pods,
					item.(*pod.Pod),
				)
			},
		)
		if err != nil {
			errChan <- err
			return
		}
		cmpl.Pods = pods
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		var units []*unit.Unit
		err := get(
			db,
			db.Units(),
			query,
			&bson.M{
				"_id":          1,
				"pod":          1,
				"organization": 1,
				"name":         1,
				"kind":         1,
			},
			nil,
			func() interface{} {
				return &unit.Unit{}
			},
			func(item interface{}) {
				unt := item.(*unit.Unit)

				units = append(
					units,
					unt,
				)

				unitsMap[unt.Id] = unt
			},
		)
		if err != nil {
			errChan <- err
			return
		}
		cmpl.Units = units
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		err = get(
			db,
			db.Deployments(),
			query,
			&bson.M{
				"_id":          1,
				"name":         1,
				"pod":          1,
				"unit":         1,
				"organization": 1,
				"kind":         1,
				"state":        1,
				"status":       1,
				"image_id":     1,
				"image_name":   1,
				"tags":         1,
			},
			&bson.M{
				"timestamp": -1,
			},
			func() interface{} {
				return &deployment.Deployment{}
			},
			func(item interface{}) {
				deployments = append(deployments, item.(*deployment.Deployment))
			},
		)
		if err != nil {
			errChan <- err
			return
		}
	}()

	wg.Wait()
	close(errChan)

	for e := range errChan {
		if e != nil {
			err = e
			return
		}
	}

	for _, imgs := range releaseImages {
		tags := []string{"latest"}
		var latestImg *image.Image

		for _, img := range imgs {
			tags = append(tags, img.Build)

			if latestImg == nil {
				latestImg = img
			} else if img.Build > latestImg.Build {
				latestImg = img
			}
		}

		latestImg.Name = latestImg.Release
		latestImg.Tags = tags
		cmpl.Images = append(cmpl.Images, latestImg)
	}
	sort.Sort(image.ImagesSort(cmpl.Images))

	cmpl.Images = append(
		cmpl.Images,
		otherImages...,
	)

	for _, deply := range deployments {
		if !deply.ImageReady() {
			return
		}

		unt := unitsMap[deply.Unit]
		if unt == nil {
			return
		}

		build := buildsMap[deply.Unit]
		if build == nil {
			build = &Build{
				Id:           deply.Unit,
				Name:         unt.Name,
				Pod:          unt.Pod,
				Organization: unt.Organization,
				Tags: []*BuildTag{
					&BuildTag{
						Tag:       "latest",
						Timestamp: deply.Timestamp,
					},
				},
			}
			buildsMap[deply.Unit] = build
		}

		for _, tag := range deply.Tags {
			build.Tags = append(build.Tags, &BuildTag{
				Tag:       tag,
				Timestamp: deply.Timestamp,
			})
		}
	}

	for _, build := range buildsMap {
		cmpl.Builds = append(cmpl.Builds, build)
	}

	return
}
