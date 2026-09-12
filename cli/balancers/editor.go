package balancers

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/balancer"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/sirupsen/logrus"
)

// commitFields are the fields written by the admin balancer handler.
var commitFields = set.NewSet(
	"name",
	"comment",
	"state",
	"type",
	"organization",
	"datacenter",
	"certificates",
	"websockets",
	"domains",
	"backends",
	"check_path",
)

var typeOptions = []widget.SelectOption{
	{Label: "HTTP", Value: balancer.Http},
}

var protocolOptions = []widget.SelectOption{
	{Label: "HTTP", Value: "http"},
	{Label: "HTTPS", Value: "https"},
}

// editor is the balancer settings form mirroring the inputs of the
// detailed balancer view in the web interface.
type editor struct {
	balnc  *balancer.Balancer
	names  *names
	create bool
	form   *form.Form

	name         *form.Text
	comment      *form.Area
	state        *form.Toggle
	typ          *form.Select
	datacenter   *form.Select
	domains      *form.Rows
	backends     *form.Rows
	websockets   *form.Toggle
	organization *form.Select
	certificates *form.Tokens
	checkPath    *form.Text
}

// domainRow creates the domain and host inputs of an external domain.
func domainRow(domn *balancer.Domain) *form.Row {
	return &form.Row{
		Cells: []form.Input{
			form.NewCell("Domain", domn.Domain),
			form.NewCell("Host", domn.Host),
		},
		Tag: domn,
	}
}

func newDomainRow() *form.Row {
	row := domainRow(&balancer.Domain{})
	row.Tag = nil
	return row
}

// backendRow creates the protocol, hostname and port inputs of an
// internal backend.
func backendRow(backend *balancer.Backend) *form.Row {
	protocol := form.NewSelect("", protocolOptions, backend.Protocol)
	protocol.Compact = true

	return &form.Row{
		Cells: []form.Input{
			protocol,
			form.NewCell("Hostname", backend.Hostname),
			form.NewNumberCell("Port", backend.Port),
		},
		Tag: backend,
	}
}

// newBackendRow returns the default backend of the web add button.
func newBackendRow() *form.Row {
	row := backendRow(&balancer.Backend{
		Protocol: "http",
		Port:     80,
	})
	row.Tag = nil
	return row
}

func hexList(ids []bson.ObjectID) []string {
	hexes := []string{}
	for _, id := range ids {
		hexes = append(hexes, id.Hex())
	}
	return hexes
}

func newEditor(balnc *balancer.Balancer, nms *names) *editor {
	e := &editor{
		balnc: balnc,
		names: nms,
	}

	e.name = form.NewText("Name", "Enter name", balnc.Name)
	e.comment = form.NewArea("Comment", "Load balancer comment",
		balnc.Comment, 4)
	e.state = form.NewToggle("Active", balnc.State)
	e.typ = form.NewSelect("Type", typeOptions, balnc.Type)

	dcOpts := resource.SelectOptions(nms.datacenters)
	if len(dcOpts) > 0 {
		dcOpts = resource.WithSelect("Select Datacenter", dcOpts)
	} else {
		dcOpts = resource.WithSelect("No Datacenters", nil)
	}
	e.datacenter = form.NewSelect("Datacenter", dcOpts,
		resource.IdHex(balnc.Datacenter))

	e.domains = form.NewRows("External Domains", "No domains",
		"Add Domain", []string{"Domain", "Host"}, newDomainRow)
	for _, domn := range balnc.Domains {
		e.domains.Add(domainRow(domn))
	}

	e.backends = form.NewRows("Internal Backends", "No backends",
		"Add Backend", []string{"Protocol", "Hostname", "Port"},
		newBackendRow)
	for _, backend := range balnc.Backends {
		e.backends.Add(backendRow(backend))
	}

	e.websockets = form.NewToggle("WebSockets", balnc.WebSockets)

	orgOpts := resource.SelectOptions(nms.orgs)
	if len(orgOpts) > 0 {
		orgOpts = resource.WithSelect("Select Organization", orgOpts)
	} else {
		orgOpts = resource.WithSelect("No Organizations", nil)
	}
	e.organization = form.NewSelect("Organization", orgOpts,
		resource.IdHex(balnc.Organization))
	e.organization.OnChange = e.applyOrganization

	e.certificates = form.NewTokensSelect("Certificates",
		e.certificateOptions(), hexList(balnc.Certificates))

	e.checkPath = form.NewText("Health Check Path", "Enter path",
		balnc.CheckPath)

	e.form = form.New(
		e.name,
		e.comment,
		e.state,
		e.typ,
		e.datacenter,
		e.domains,
		e.backends,
		e.websockets,
		e.organization,
		e.certificates,
		e.checkPath,
	)

	return e
}

// certificateOptions lists the certificates of the selected
// organization like the web interface.
func (e *editor) certificateOptions() []widget.SelectOption {
	return resource.OrgSelectOptions(e.names.certificates,
		resource.ParseId(e.organization.Value()))
}

// applyOrganization updates the certificate choices when the
// organization changes.
func (e *editor) applyOrganization() {
	e.certificates.Options = e.certificateOptions()
}

func (e *editor) Form() *form.Form {
	return e.form
}

func (e *editor) buildDomains() []*balancer.Domain {
	domains := []*balancer.Domain{}
	for _, row := range e.domains.Rows() {
		domains = append(domains, &balancer.Domain{
			Domain: row.Cells[0].(*form.Text).Value(),
			Host:   row.Cells[1].(*form.Text).Value(),
		})
	}
	return domains
}

func (e *editor) buildBackends() []*balancer.Backend {
	backends := []*balancer.Backend{}
	for _, row := range e.backends.Rows() {
		backends = append(backends, &balancer.Backend{
			Protocol: row.Cells[0].(*form.Select).Value(),
			Hostname: row.Cells[1].(*form.Text).Value(),
			Port:     row.Cells[2].(*form.Number).Int(),
		})
	}
	return backends
}

func (e *editor) buildCertificates() []bson.ObjectID {
	certs := []bson.ObjectID{}
	for _, hex := range e.certificates.Values() {
		certId := resource.ParseId(hex)
		if !certId.IsZero() {
			certs = append(certs, certId)
		}
	}
	return certs
}

// apply sets the form values on the balancer.
func (e *editor) apply(balnc *balancer.Balancer) {
	balnc.Name = e.name.Value()
	balnc.Comment = e.comment.Value()
	balnc.State = e.state.Value()
	balnc.Type = e.typ.Value()
	balnc.Organization = resource.ParseId(e.organization.Value())
	balnc.Datacenter = resource.ParseId(e.datacenter.Value())
	balnc.Certificates = e.buildCertificates()
	balnc.WebSockets = e.websockets.Value()
	balnc.Domains = e.buildDomains()
	balnc.Backends = e.buildBackends()
	balnc.CheckPath = e.checkPath.Value()
}

// Save applies the form to the current balancer the same way as the
// admin balancer handler, or inserts a new balancer.
func (e *editor) Save(db *database.Database) (err error) {
	var balnc *balancer.Balancer
	if e.create {
		balnc = &balancer.Balancer{}
	} else {
		balnc, err = balancer.Get(db, e.balnc.Id)
		if err != nil {
			return
		}
	}

	e.apply(balnc)

	errData, err := balnc.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = &errortypes.ParseError{
			errors.New("balancers: " + errData.Message),
		}
		return
	}

	if e.create {
		err = balnc.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"balancer_id": balnc.Id.Hex(),
		}).Info("tui: Balancer created")
	} else {
		err = balnc.CommitFields(db, commitFields)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"balancer_id": balnc.Id.Hex(),
		}).Info("tui: Balancer settings saved")
	}

	err = event.PublishDispatch(db, "balancer.change")
	if err != nil {
		return
	}

	return
}
