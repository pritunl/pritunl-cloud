// Package disks provides the disks tab, a paged list of the instance
// disks in the cluster from the admin perspective.
package disks

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
)

const (
	// cardFields is the number of fields on each disk card.
	cardFields = 5
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "Disks"
}

func (p *Provider) Empty() string {
	return "No disks"
}

func (p *Provider) Singular() string {
	return "disk"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// FilterFields mirrors the disks filter of the web interface.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	orgs, err := resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}

	nodes, err := resource.AllNames(db, db.Nodes(), bson.M{
		"types": node.Hypervisor,
	})
	if err != nil {
		return
	}

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "Disk ID",
			Placeholder: "Disk ID",
		},
		{
			Key:         "name",
			Label:       "Name",
			Placeholder: "Name",
		},
		{
			Key:         "instance",
			Label:       "Instance ID",
			Placeholder: "Instance ID",
		},
		{
			Key:     "organization",
			Label:   "Organization",
			Options: resource.SelectOptions(orgs),
		},
		{
			Key:     "node",
			Label:   "Node",
			Options: resource.SelectOptions(nodes),
		},
	}

	return
}

// nodeInstance is an instance name with its node, the instance select
// of a disk lists the instances on the node of the disk like the web
// interface.
type nodeInstance struct {
	Id   bson.ObjectID `bson:"_id"`
	Name string        `bson:"name"`
	Node bson.ObjectID `bson:"node"`
}

// loadNodeInstances returns the instances on the nodes keyed by node.
func loadNodeInstances(db *database.Database, nodeIds []bson.ObjectID) (
	instances map[bson.ObjectID][]*nodeInstance, err error) {

	instances = map[bson.ObjectID][]*nodeInstance{}
	if len(nodeIds) == 0 {
		return
	}

	cursor, err := db.Instances().Find(
		db,
		bson.M{
			"node": bson.M{
				"$in": nodeIds,
			},
		},
		options.Find().
			SetSort(bson.D{
				{"name", 1},
			}).
			SetProjection(bson.D{
				{"name", 1},
				{"node", 1},
			}),
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		inst := &nodeInstance{}
		err = cursor.Decode(inst)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		instances[inst.Node] = append(instances[inst.Node], inst)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

// instanceOptions converts instances to the choices of the instance
// select.
func instanceOptions(insts []*nodeInstance) []widget.SelectOption {
	opts := []widget.SelectOption{}
	for _, inst := range insts {
		opts = append(opts, widget.SelectOption{
			Label: inst.Name,
			Value: inst.Id.Hex(),
		})
	}
	return opts
}

// Load returns a page of disks with the organization, node, pool and
// image names resolved and the instances of each node for the instance
// select.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	disks, count, err := aggregate.GetDiskPaged(db, &query, page, pageCount)
	if err != nil {
		return
	}

	orgIds := resource.IdSet{}
	nodeIds := resource.IdSet{}
	poolIds := resource.IdSet{}
	imageIds := resource.IdSet{}
	for _, dsk := range disks {
		orgIds.Add(dsk.Organization)
		nodeIds.Add(dsk.Node)
		poolIds.Add(dsk.Pool)
		imageIds.Add(dsk.Image)
	}

	orgNames, err := resource.Names(db, db.Organizations(), orgIds.List())
	if err != nil {
		return
	}

	nodeNames, err := resource.Names(db, db.Nodes(), nodeIds.List())
	if err != nil {
		return
	}

	poolNames, err := resource.Names(db, db.Pools(), poolIds.List())
	if err != nil {
		return
	}

	imageNames, err := resource.Names(db, db.Images(), imageIds.List())
	if err != nil {
		return
	}

	// LVM disks list the instances of every node in the pool
	poolNodes := map[bson.ObjectID][]bson.ObjectID{}
	for _, poolId := range poolIds.List() {
		nodes, e := node.GetAllPool(db, poolId)
		if e != nil {
			err = e
			return
		}

		for _, nde := range nodes {
			poolNodes[poolId] = append(poolNodes[poolId], nde.Id)
			nodeIds.Add(nde.Id)
		}
	}

	nodeInstances, err := loadNodeInstances(db, nodeIds.List())
	if err != nil {
		return
	}

	items := make([]resource.Item, 0, len(disks))
	for _, dsk := range disks {
		insts := []*nodeInstance{}
		if !dsk.Node.IsZero() {
			insts = nodeInstances[dsk.Node]
		} else if !dsk.Pool.IsZero() {
			for _, nodeId := range poolNodes[dsk.Pool] {
				insts = append(insts, nodeInstances[nodeId]...)
			}
		}

		items = append(items, &Item{
			dsk:       dsk,
			instances: instanceOptions(insts),
			org:       orgNames[dsk.Organization],
			node:      nodeNames[dsk.Node],
			pool:      poolNames[dsk.Pool],
			image:     imageNames[dsk.Image],
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}
