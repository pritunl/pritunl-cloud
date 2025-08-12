package completion

import (
	"sort"
	"sync"
	"time"

	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/plan"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/pritunl/pritunl-cloud/unit"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
)

type Completion struct {
	Organizations []*database.Named         `json:"organizations"`
	Authorities   []*database.Named         `json:"authorities"`
	Policies      []*database.Named         `json:"policies"`
	Domains       []*domain.Completion      `json:"domains"`
	Vpcs          []*vpc.Completion         `json:"vpcs"`
	Datacenters   []*datacenter.Completion  `json:"datacenters"`
	Blocks        []*block.Completion       `json:"blocks"`
	Nodes         []*node.Completion        `json:"nodes"`
	Pools         []*pool.Completion        `json:"pools"`
	Zones         []*zone.Completion        `json:"zones"`
	Shapes        []*shape.Completion       `json:"shapes"`
	Images        []*image.Completion       `json:"images"`
	Storages      []*storage.Completion     `json:"storages"`
	Builds        []*Build                  `json:"builds"`
	Instances     []*instance.Completion    `json:"instances"`
	Plans         []*plan.Completion        `json:"plans"`
	Certificates  []*certificate.Completion `json:"certificates"`
	Secrets       []*secret.Completion      `json:"secrets"`
	Pods          []*pod.Completion         `json:"pods"`
	Units         []*unit.Completion        `json:"units"`
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
	query bson.M, projection *bson.M, sort *bson.D, new func() interface{},
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

func GetCompletion(db *database.Database, orgId primitive.ObjectID,
	orgRoles []string) (cmpl *Completion, err error) {

	cmpl = &Completion{}
	query := bson.M{}
	if !orgId.IsZero() {
		query["organization"] = orgId
	}

	releaseImages := map[string][]*image.Completion{}
	otherImages := []*image.Completion{}
	unitsMap := map[primitive.ObjectID]*unit.Completion{}
	buildsMap := map[primitive.ObjectID]*Build{}
	deployments := []*deployment.Deployment{}

	var wg sync.WaitGroup
	errChan := make(chan error, 16)

	wg.Add(1)
	go func() {
		defer wg.Done()

		var orgQuery bson.M
		if !orgId.IsZero() {
			if orgRoles == nil {
				orgRoles = []string{}
			}

			orgQuery = bson.M{
				"roles": bson.M{
					"$in": orgRoles,
				},
			}
		} else {
			orgQuery = bson.M{}
		}

		var orgs []*database.Named
		err := get(
			db,
			db.Organizations(),
			orgQuery,
			&bson.M{
				"_id":  1,
				"name": 1,
			},
			&bson.D{
				{"name", 1},
			},
			func() interface{} {
				return &database.Named{}
			},
			func(item interface{}) {
				orgs = append(
					orgs,
					item.(*database.Named),
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

		var authrs []*database.Named
		err := get(
			db,
			db.Authorities(),
			query,
			&bson.M{
				"_id":  1,
				"name": 1,
			},
			&bson.D{
				{"name", 1},
			},
			func() interface{} {
				return &database.Named{}
			},
			func(item interface{}) {
				authrs = append(
					authrs,
					item.(*database.Named),
				)
			},
		)
		if err != nil {
			errChan <- err
			return
		}
		cmpl.Authorities = authrs
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		var domains []*domain.Completion
		err := get(
			db,
			db.Domains(),
			query,
			&bson.M{
				"_id":          1,
				"name":         1,
				"organization": 1,
			},
			&bson.D{
				{"name", 1},
			},
			func() interface{} {
				return &domain.Completion{}
			},
			func(item interface{}) {
				domains = append(
					domains,
					item.(*domain.Completion),
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

		var vpcs []*vpc.Completion
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
			&bson.D{
				{"name", 1},
			},
			func() interface{} {
				return &vpc.Completion{}
			},
			func(item interface{}) {
				vpcs = append(
					vpcs,
					item.(*vpc.Completion),
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

		var dcQuery bson.M
		if !orgId.IsZero() {
			dcQuery = bson.M{
				"$or": []bson.M{
					bson.M{
						"match_organizations": false,
					},
					bson.M{
						"organizations": orgId,
					},
				},
			}
		} else {
			dcQuery = bson.M{}
		}

		var datacenters []*datacenter.Completion
		err := get(
			db,
			db.Datacenters(),
			dcQuery,
			&bson.M{
				"_id":          1,
				"name":         1,
				"network_mode": 1,
			},
			&bson.D{
				{"name", 1},
			},
			func() interface{} {
				return &datacenter.Completion{}
			},
			func(item interface{}) {
				datacenters = append(
					datacenters,
					item.(*datacenter.Completion),
				)
			},
		)
		if err != nil {
			errChan <- err
			return
		}
		cmpl.Datacenters = datacenters
	}()

	if orgId.IsZero() {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var blocks []*block.Completion
			err := get(
				db,
				db.Blocks(),
				query,
				&bson.M{
					"_id":  1,
					"name": 1,
					"type": 1,
				},
				&bson.D{
					{"name", 1},
				},
				func() interface{} {
					return &block.Completion{}
				},
				func(item interface{}) {
					blocks = append(
						blocks,
						item.(*block.Completion),
					)
				},
			)
			if err != nil {
				errChan <- err
				return
			}
			cmpl.Blocks = blocks
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		var nodes []*node.Completion
		err := get(
			db,
			db.Nodes(),
			bson.M{},
			&bson.M{
				"_id":   1,
				"name":  1,
				"zone":  1,
				"types": 1,
			},
			&bson.D{
				{"name", 1},
			},
			func() interface{} {
				return &node.Completion{}
			},
			func(item interface{}) {
				nde := item.(*node.Completion)

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

		var pools []*pool.Completion
		err := get(
			db,
			db.Pools(),
			query,
			&bson.M{
				"_id":  1,
				"name": 1,
				"zone": 1,
			},
			&bson.D{
				{"name", 1},
			},
			func() interface{} {
				return &pool.Completion{}
			},
			func(item interface{}) {
				pools = append(
					pools,
					item.(*pool.Completion),
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

		var zones []*zone.Completion
		err := get(
			db,
			db.Zones(),
			bson.M{},
			&bson.M{
				"_id":        1,
				"name":       1,
				"datacenter": 1,
			},
			&bson.D{
				{"name", 1},
			},
			func() interface{} {
				return &zone.Completion{}
			},
			func(item interface{}) {
				zones = append(
					zones,
					item.(*zone.Completion),
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

		var shapes []*shape.Completion
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
			&bson.D{
				{"name", 1},
			},
			func() interface{} {
				return &shape.Completion{}
			},
			func(item interface{}) {
				shapes = append(
					shapes,
					item.(*shape.Completion),
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
			&bson.D{
				{"name", 1},
			},
			func() interface{} {
				return &image.Completion{}
			},
			func(item interface{}) {
				img := item.(*image.Completion)
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

		var storages []*storage.Completion
		if orgId.IsZero() {
			err := get(
				db,
				db.Storages(),
				bson.M{},
				&bson.M{
					"_id":  1,
					"name": 1,
					"type": 1,
				},
				&bson.D{
					{"name", 1},
				},
				func() interface{} {
					return &storage.Completion{}
				},
				func(item interface{}) {
					storages = append(
						storages,
						item.(*storage.Completion),
					)
				},
			)
			if err != nil {
				errChan <- err
				return
			}
		}
		cmpl.Storages = storages
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		var instances []*instance.Completion
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
			&bson.D{
				{"name", 1},
			},
			func() interface{} {
				return &instance.Completion{}
			},
			func(item interface{}) {
				instances = append(
					instances,
					item.(*instance.Completion),
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

		var plans []*plan.Completion
		err := get(
			db,
			db.Plans(),
			query,
			&bson.M{
				"_id":          1,
				"name":         1,
				"organization": 1,
			},
			&bson.D{
				{"name", 1},
			},
			func() interface{} {
				return &plan.Completion{}
			},
			func(item interface{}) {
				plans = append(
					plans,
					item.(*plan.Completion),
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

		var certificates []*certificate.Completion
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
			&bson.D{
				{"name", 1},
			},
			func() interface{} {
				return &certificate.Completion{}
			},
			func(item interface{}) {
				certificates = append(
					certificates,
					item.(*certificate.Completion),
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

		var secrets []*secret.Completion
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
			&bson.D{
				{"name", 1},
			},
			func() interface{} {
				return &secret.Completion{}
			},
			func(item interface{}) {
				secrets = append(
					secrets,
					item.(*secret.Completion),
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

		var pods []*pod.Completion
		err := get(
			db,
			db.Pods(),
			query,
			&bson.M{
				"_id":          1,
				"name":         1,
				"organization": 1,
			},
			&bson.D{
				{"name", 1},
			},
			func() interface{} {
				return &pod.Completion{}
			},
			func(item interface{}) {
				pods = append(
					pods,
					item.(*pod.Completion),
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

		var units []*unit.Completion
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
			&bson.D{
				{"name", 1},
			},
			func() interface{} {
				return &unit.Completion{}
			},
			func(item interface{}) {
				unt := item.(*unit.Completion)

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

		var deplyQuery bson.M
		if !orgId.IsZero() {
			deplyQuery = bson.M{
				"organizations": orgId,
				"kind":          "image",
			}
		} else {
			deplyQuery = bson.M{
				"kind": "image",
			}
		}

		err = get(
			db,
			db.Deployments(),
			deplyQuery,
			&bson.M{
				"_id":          1,
				"name":         1,
				"pod":          1,
				"unit":         1,
				"organization": 1,
				"kind":         1,
				"state":        1,
				"status":       1,
				"image":        1,
				"image_data":   1,
				"tags":         1,
			},
			&bson.D{
				{"timestamp", -1},
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
		var latestImg *image.Completion

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
	sort.Sort(image.CompletionsSort(cmpl.Images))

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
