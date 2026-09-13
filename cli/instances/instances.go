// Package instances provides the instances tab, a paged list of the
// instances in the cluster from the admin perspective.
package instances

import (
	"fmt"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/vpc"
)

const (
	// cardFields is the number of fields on each instance card.
	cardFields = 8
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Title() string {
	return "Instances"
}

func (p *Provider) Empty() string {
	return "No instances"
}

func (p *Provider) Singular() string {
	return "instance"
}

func (p *Provider) FieldCount() int {
	return cardFields
}

// FilterFields mirrors the instances filter of the web interface.
func (p *Provider) FilterFields(db *database.Database) (
	fields []resource.FilterField, err error) {

	nodes, err := resource.AllNames(db, db.Nodes(), bson.M{
		"types": node.Hypervisor,
	})
	if err != nil {
		return
	}

	zones, err := resource.AllNames(db, db.Zones(), bson.M{})
	if err != nil {
		return
	}

	orgs, err := resource.AllNames(db, db.Organizations(), bson.M{})
	if err != nil {
		return
	}

	vpcs, err := vpc.GetAllNames(db, &bson.M{})
	if err != nil {
		return
	}

	vpcOpts := []widget.SelectOption{}
	subnetOpts := []widget.SelectOption{}
	for _, vc := range vpcs {
		vpcOpts = append(vpcOpts, widget.SelectOption{
			Label: vc.Name,
			Value: vc.Id.Hex(),
		})

		for _, sub := range vc.Subnets {
			subnetOpts = append(subnetOpts, widget.SelectOption{
				Label: fmt.Sprintf("%s - %s - %s",
					vc.Name, sub.Name, sub.Network),
				Value: sub.Id.Hex(),
			})
		}
	}

	fields = []resource.FilterField{
		{
			Key:         "id",
			Label:       "Instance ID",
			Placeholder: "Instance ID",
		},
		{
			Key:         "name",
			Label:       "Name",
			Placeholder: "Name",
		},
		{
			Key:         "comment",
			Label:       "Comment",
			Placeholder: "Comment",
		},
		{
			Key:         "role",
			Label:       "Network Role",
			Placeholder: "Network Role",
		},
		{
			Key:         "address",
			Label:       "IP Address",
			Placeholder: "IP Address",
		},
		{
			Key:         "network_namespace",
			Label:       "Network Namespace",
			Placeholder: "Network Namespace",
		},
		{
			Key:     "node",
			Label:   "Node",
			Options: resource.SelectOptions(nodes),
		},
		{
			Key:     "zone",
			Label:   "Zone",
			Options: resource.SelectOptions(zones),
		},
		{
			Key:     "vpc",
			Label:   "VPC",
			Options: vpcOpts,
		},
		{
			Key:     "subnet",
			Label:   "Subnet",
			Options: subnetOpts,
		},
		{
			Key:     "organization",
			Label:   "Organization",
			Options: resource.SelectOptions(orgs),
		},
	}

	return
}

// Load returns a page of instances with the node, zone and organization
// names of the page resolved.
func (p *Provider) Load(db *database.Database, filter resource.Filter,
	page, pageCount int64) (result *resource.Page, err error) {

	query := buildQuery(filter)

	insts, count, err := instance.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		return
	}

	nodeIds := resource.IdSet{}
	zoneIds := resource.IdSet{}
	orgIds := resource.IdSet{}
	for _, inst := range insts {
		nodeIds.Add(inst.Node)
		zoneIds.Add(inst.Zone)
		orgIds.Add(inst.Organization)
	}

	nodeNames, err := resource.Names(db, db.Nodes(), nodeIds.List())
	if err != nil {
		return
	}

	zoneNames, err := resource.Names(db, db.Zones(), zoneIds.List())
	if err != nil {
		return
	}

	orgNames, err := resource.Names(db, db.Organizations(), orgIds.List())
	if err != nil {
		return
	}

	// VPCs and subnets are the settings select options
	vpcs, err := vpc.GetAllNames(db, &bson.M{})
	if err != nil {
		return
	}

	items := make([]resource.Item, 0, len(insts))
	for _, inst := range insts {
		inst.Json(false)

		items = append(items, &Item{
			inst: inst,
			vpcs: vpcs,
			node: nodeNames[inst.Node],
			zone: zoneNames[inst.Zone],
			org:  orgNames[inst.Organization],
		})
	}

	result = &resource.Page{
		Items: items,
		Count: count,
	}

	return
}
