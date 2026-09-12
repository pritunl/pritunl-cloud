package domains

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/sirupsen/logrus"
)

// commitFields are the fields written by the admin domain handler, the
// records are synced separately.
var commitFields = set.NewSet(
	"name",
	"comment",
	"organization",
	"type",
	"secret",
	"root_domain",
)

var typeOptions = []widget.SelectOption{
	{Label: "Local", Value: domain.Local},
	{Label: "AWS", Value: domain.AWS},
	{Label: "Cloudflare", Value: domain.Cloudflare},
	{Label: "Oracle Cloud", Value: domain.OracleCloud},
	{Label: "Google Cloud", Value: "google_cloud"},
}

var recordTypeOptions = []widget.SelectOption{
	{Label: "A", Value: domain.A},
	{Label: "AAAA", Value: domain.AAAA},
	{Label: "CNAME", Value: domain.CNAME},
}

// editor is the domain settings form mirroring the inputs of the
// detailed domain view in the web interface.
type editor struct {
	domn   *domain.Domain
	names  *names
	create bool
	form   *form.Form

	name         *form.Text
	comment      *form.Area
	rootDomain   *form.Text
	records      *form.Rows
	organization *form.Select
	typ          *form.Select
	secret       *form.Select
}

// recordRow creates the type, sub domain and value inputs of a record.
func recordRow(rec *domain.Record) *form.Row {
	typ := form.NewSelect("", recordTypeOptions, rec.Type)
	typ.Compact = true

	return &form.Row{
		Cells: []form.Input{
			typ,
			form.NewCell("Sub Domain", rec.SubDomain),
			form.NewCell("IP Address", rec.Value),
		},
		Tag: rec,
	}
}

func newRecordRow() *form.Row {
	row := recordRow(&domain.Record{
		Type: domain.A,
	})
	row.Tag = nil
	return row
}

func newEditor(domn *domain.Domain, nms *names) *editor {
	e := &editor{
		domn:  domn,
		names: nms,
	}

	e.name = form.NewText("Name", "Enter name", domn.Name)
	e.comment = form.NewArea("Comment", "Domain comment", domn.Comment, 4)
	e.rootDomain = form.NewText("Domain", "Enter domain", domn.RootDomain)

	e.records = form.NewRows("Domain Records", "No records", "Add Record",
		[]string{"Type", "Sub Domain", "Value"}, newRecordRow)
	for _, rec := range domn.Records {
		e.records.Add(recordRow(rec))
	}

	orgOpts := resource.SelectOptions(nms.orgs)
	if len(orgOpts) == 0 {
		orgOpts = resource.WithSelect("No Organizations", nil)
	}
	e.organization = form.NewSelect("Organization", orgOpts,
		resource.IdHex(domn.Organization))

	e.typ = form.NewSelect("Provider", typeOptions, domn.Type)

	e.secret = form.NewSelect("Provider API Secret", nil,
		resource.IdHex(domn.Secret))
	e.secret.Dynamic = e.secretOptions
	e.secret.Hide = func() bool {
		typ := e.typ.Value()
		return typ == "" || typ == domain.Local
	}

	e.form = form.New(
		e.name,
		e.comment,
		e.rootDomain,
		e.records,
		e.organization,
		e.typ,
		e.secret,
	)

	return e
}

// secretOptions lists the secrets of the selected organization like the
// web interface.
func (e *editor) secretOptions() []widget.SelectOption {
	opts := resource.OrgSelectOptions(e.names.secrets,
		resource.ParseId(e.organization.Value()))
	if len(opts) == 0 {
		return resource.WithSelect("No Secrets", nil)
	}
	return resource.WithSelect("Select Secret", opts)
}

func (e *editor) Form() *form.Form {
	return e.form
}

// buildRecords returns the records of the form the same way the web
// interface sends them, kept records carry their id, added records the
// insert operation and removed records the delete operation.
func (e *editor) buildRecords() []*domain.Record {
	records := []*domain.Record{}

	for _, row := range e.records.Rows() {
		rec := &domain.Record{
			Type:      row.Cells[0].(*form.Select).Value(),
			SubDomain: row.Cells[1].(*form.Text).Value(),
			Value:     row.Cells[2].(*form.Text).Value(),
		}

		if orig, ok := row.Tag.(*domain.Record); ok {
			rec.Id = orig.Id
		} else {
			rec.Operation = domain.INSERT
		}

		records = append(records, rec)
	}

	for _, tag := range e.records.Removed() {
		orig, ok := tag.(*domain.Record)
		if !ok {
			continue
		}

		records = append(records, &domain.Record{
			Id:        orig.Id,
			Type:      orig.Type,
			SubDomain: orig.SubDomain,
			Value:     orig.Value,
			Operation: domain.DELETE,
		})
	}

	return records
}

// apply sets the form values on the domain.
func (e *editor) apply(domn *domain.Domain) {
	domn.Name = e.name.Value()
	domn.Comment = e.comment.Value()
	domn.Organization = resource.ParseId(e.organization.Value())
	domn.Type = e.typ.Value()
	domn.Secret = resource.ParseId(e.secret.Value())
	domn.RootDomain = e.rootDomain.Value()
}

func saveError(errData *errortypes.ErrorData) error {
	return &errortypes.ParseError{
		errors.New("domains: " + errData.Message),
	}
}

// Save applies the form to the current domain and syncs its records the
// same way as the admin domain handler, or inserts a new domain.
func (e *editor) Save(db *database.Database) (err error) {
	if e.create {
		domn := &domain.Domain{}
		e.apply(domn)

		errData, er := domn.Validate(db)
		if er != nil {
			err = er
			return
		}
		if errData != nil {
			err = saveError(errData)
			return
		}

		err = domn.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"domain_id": domn.Id.Hex(),
		}).Info("tui: Domain created")
	} else {
		domn, er := domain.Get(db, e.domn.Id)
		if er != nil {
			err = er
			return
		}

		err = domn.LoadRecords(db, true)
		if err != nil {
			return
		}

		domn.PreCommit()
		e.apply(domn)

		errData, er := domn.Validate(db)
		if er != nil {
			err = er
			return
		}
		if errData != nil {
			err = saveError(errData)
			return
		}

		err = domn.CommitFields(db, commitFields)
		if err != nil {
			return
		}

		errData, err = domn.SyncRecords(db, e.buildRecords())
		if err != nil {
			return
		}
		if errData != nil {
			err = saveError(errData)
			return
		}

		logrus.WithFields(logrus.Fields{
			"domain_id": domn.Id.Hex(),
		}).Info("tui: Domain settings saved")
	}

	err = event.PublishDispatch(db, "domain.change")
	if err != nil {
		return
	}

	return
}
