package vpcs

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/sirupsen/logrus"
)

// commitFields are the fields written by the admin VPC handler.
var commitFields = set.NewSet(
	"name",
	"comment",
	"icmp_redirects",
	"routes",
	"maps",
	"arps",
	"subnets",
)

// editor is the VPC settings form mirroring the inputs of the detailed
// VPC view in the web interface, the organization and datacenter are
// only chosen when creating a VPC.
type editor struct {
	vc     *vpc.Vpc
	names  *names
	create bool
	form   *form.Form

	name          *form.Text
	comment       *form.Area
	network       *form.Text
	subnets       *form.Rows
	maps          *form.Rows
	organization  *form.Select
	datacenter    *form.Select
	icmpRedirects *form.Toggle
	routes        *form.Rows
	arps          *form.Rows
}

// subnetRow creates the name and network inputs of a subnet, the
// network of an existing subnet cannot be changed.
func subnetRow(sub *vpc.Subnet) *form.Row {
	network := form.NewCell("Network", sub.Network)
	if !sub.Id.IsZero() {
		network.Lock = func() bool {
			return true
		}
	}

	return &form.Row{
		Cells: []form.Input{
			form.NewCell("Name", sub.Name),
			network,
		},
		Tag: sub,
	}
}

func newSubnetRow() *form.Row {
	row := subnetRow(&vpc.Subnet{})
	row.Tag = nil
	return row
}

// pairRow creates a destination and target style row with two text
// cells.
func pairRow(first, second string, tag interface{}) *form.Row {
	return &form.Row{
		Cells: []form.Input{
			form.NewCell("", first),
			form.NewCell("", second),
		},
		Tag: tag,
	}
}

func newRouteRow() *form.Row {
	return pairRow("", "", nil)
}

func newEditor(vc *vpc.Vpc, nms *names, create bool) *editor {
	e := &editor{
		vc:     vc,
		names:  nms,
		create: create,
	}

	e.name = form.NewText("Name", "Enter name", vc.Name)
	e.comment = form.NewArea("Comment", "VPC comment", vc.Comment, 4)
	e.network = form.NewText("Network", "Enter network", vc.Network)

	e.subnets = form.NewRows("Subnets", "No subnets", "Add Subnet",
		[]string{"Name", "Network"}, newSubnetRow)
	for _, sub := range vc.Subnets {
		e.subnets.Add(subnetRow(sub))
	}

	e.maps = form.NewRows("Network Maps", "No network maps", "Add Map",
		[]string{"Destination", "Target"}, newRouteRow)
	for _, mp := range vc.Maps {
		e.maps.Add(pairRow(mp.Destination, mp.Target, mp))
	}

	orgOpts := resource.SelectOptions(nms.orgs)
	if len(orgOpts) == 0 {
		orgOpts = resource.WithSelect("No Organizations", nil)
	}
	e.organization = form.NewSelect("Organization", orgOpts,
		resource.IdHex(vc.Organization))

	dcOpts := resource.SelectOptions(nms.datacenters)
	if len(dcOpts) > 0 {
		dcOpts = resource.WithSelect("Select Datacenter", dcOpts)
	} else {
		dcOpts = resource.WithSelect("No Datacenters", nil)
	}
	e.datacenter = form.NewSelect("Datacenter", dcOpts,
		resource.IdHex(vc.Datacenter))

	e.icmpRedirects = form.NewToggle("ICMP Redirects", vc.IcmpRedirects)

	e.routes = form.NewRows("Route Table", "Default route only",
		"Add Route", []string{"Destination", "Target"}, newRouteRow)
	for _, route := range vc.Routes {
		e.routes.Add(pairRow(route.Destination, route.Target, route))
	}

	e.arps = form.NewRows("Custom ARP", "No ARP entries", "Add ARP",
		[]string{"IP Address", "Mac Address"}, newRouteRow)
	for _, arp := range vc.Arps {
		e.arps.Add(pairRow(arp.Ip, arp.Mac, arp))
	}

	inputs := []form.Input{
		e.name,
		e.comment,
		e.network,
		e.subnets,
		e.maps,
	}
	if create {
		inputs = append(inputs, e.organization, e.datacenter)
	}
	inputs = append(inputs,
		e.icmpRedirects,
		e.routes,
		e.arps,
	)

	e.form = form.New(inputs...)

	return e
}

func (e *editor) Form() *form.Form {
	return e.form
}

func cell(row *form.Row, col int) string {
	return row.Cells[col].(*form.Text).Value()
}

// buildSubnets returns the subnets keeping the ids of existing subnets
// so their addresses are preserved.
func (e *editor) buildSubnets() []*vpc.Subnet {
	subnets := []*vpc.Subnet{}
	for _, row := range e.subnets.Rows() {
		sub := &vpc.Subnet{
			Name:    cell(row, 0),
			Network: cell(row, 1),
		}
		if orig, ok := row.Tag.(*vpc.Subnet); ok {
			sub.Id = orig.Id
		}
		subnets = append(subnets, sub)
	}
	return subnets
}

func (e *editor) buildRoutes() []*vpc.Route {
	routes := []*vpc.Route{}
	for _, row := range e.routes.Rows() {
		routes = append(routes, &vpc.Route{
			Destination: cell(row, 0),
			Target:      cell(row, 1),
		})
	}
	return routes
}

func (e *editor) buildMaps() []*vpc.Map {
	maps := []*vpc.Map{}
	for _, row := range e.maps.Rows() {
		maps = append(maps, &vpc.Map{
			Destination: cell(row, 0),
			Target:      cell(row, 1),
		})
	}
	return maps
}

func (e *editor) buildArps() []*vpc.Arp {
	arps := []*vpc.Arp{}
	for _, row := range e.arps.Rows() {
		arps = append(arps, &vpc.Arp{
			Ip:  cell(row, 0),
			Mac: cell(row, 1),
		})
	}
	return arps
}

// apply sets the form values on the VPC, the network, organization and
// datacenter are only set on creation like the admin handler.
func (e *editor) apply(vc *vpc.Vpc) {
	vc.Name = e.name.Value()
	vc.Comment = e.comment.Value()
	vc.IcmpRedirects = e.icmpRedirects.Value()
	vc.Routes = e.buildRoutes()
	vc.Maps = e.buildMaps()
	vc.Arps = e.buildArps()
	vc.Subnets = e.buildSubnets()

	if e.create {
		vc.Network = e.network.Value()
		vc.Organization = resource.ParseId(e.organization.Value())
		vc.Datacenter = resource.ParseId(e.datacenter.Value())
	}
}

func saveError(errData *errortypes.ErrorData) error {
	return &errortypes.ParseError{
		errors.New("vpcs: " + errData.Message),
	}
}

// Save applies the form to the current VPC the same way as the admin VPC
// handler, or inserts a new VPC.
func (e *editor) Save(db *database.Database) (err error) {
	if e.create {
		vc := &vpc.Vpc{}
		e.apply(vc)
		vc.InitVpc()

		errData, er := vc.Validate(db)
		if er != nil {
			err = er
			return
		}
		if errData != nil {
			err = saveError(errData)
			return
		}

		err = vc.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"vpc_id": vc.Id.Hex(),
		}).Info("tui: VPC created")
	} else {
		vc, er := vpc.Get(db, e.vc.Id)
		if er != nil {
			err = er
			return
		}

		vc.PreCommit()
		e.apply(vc)

		errData, er := vc.Validate(db)
		if er != nil {
			err = er
			return
		}
		if errData != nil {
			err = saveError(errData)
			return
		}

		errData, err = vc.PostCommit(db)
		if err != nil {
			return
		}
		if errData != nil {
			err = saveError(errData)
			return
		}

		err = vc.CommitFields(db, commitFields)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"vpc_id": vc.Id.Hex(),
		}).Info("tui: VPC settings saved")
	}

	err = event.PublishDispatch(db, "vpc.change")
	if err != nil {
		return
	}

	return
}
