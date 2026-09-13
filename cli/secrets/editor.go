package secrets

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/sirupsen/logrus"
)

// commitFields are the fields written by the admin secret handler.
var commitFields = set.NewSet(
	"name",
	"comment",
	"organization",
	"type",
	"key",
	"value",
	"data",
	"region",
	"public_key",
	"private_key",
)

var typeOptions = []widget.SelectOption{
	{Label: "AWS", Value: secret.AWS},
	{Label: "Cloudflare", Value: secret.Cloudflare},
	{Label: "Oracle Cloud", Value: secret.OracleCloud},
	{Label: "Google Cloud", Value: secret.GoogleCloud},
	{Label: "JSON", Value: secret.Json},
}

// typeLabels are the labels and placeholders of the key inputs of each
// secret type from the web interface, empty labels hide the input.
type typeLabels struct {
	key            string
	keyPlaceholder string
	keyArea        bool
	value          string
	valPlaceholder string
	region         string
	publicKey      string
}

func labelsOf(typ string) typeLabels {
	switch typ {
	case secret.Cloudflare:
		return typeLabels{
			key:            "Cloudflare Token",
			keyPlaceholder: "Token",
		}
	case secret.OracleCloud:
		return typeLabels{
			key:            "Oracle Cloud Tenancy OCID",
			keyPlaceholder: "Tenancy OCID",
			value:          "Oracle Cloud User OCID",
			valPlaceholder: "User OCID",
			region:         "Oracle Cloud Region",
			publicKey:      "Oracle Cloud Public Key",
		}
	case secret.GoogleCloud:
		return typeLabels{
			key:            "Google Cloud Service Account JSON",
			keyPlaceholder: "Service Account",
			keyArea:        true,
		}
	case secret.Json:
		return typeLabels{}
	default:
		return typeLabels{
			key:            "AWS Key ID",
			keyPlaceholder: "Key ID",
			value:          "AWS Secret ID",
			valPlaceholder: "Key ID",
			region:         "AWS Region",
		}
	}
}

// editor is the secret settings form mirroring the inputs of the
// detailed secret view in the web interface, the key inputs change
// their labels and visibility with the type.
type editor struct {
	secr   *secret.Secret
	create bool
	form   *form.Form

	name         *form.Text
	comment      *form.Area
	key          *form.Text
	keyArea      *form.Area
	value        *form.Text
	region       *form.Text
	data         *form.Area
	typ          *form.Select
	organization *form.Select
	publicKey    *form.Area
}

func newEditor(secr *secret.Secret, orgs []*database.Named) *editor {
	e := &editor{
		secr: secr,
	}

	labels := labelsOf(secr.Type)

	e.name = form.NewText("Name", "Name", secr.Name)
	e.comment = form.NewArea("Comment", "Secret comment", secr.Comment, 4)

	e.key = form.NewText(labels.key, labels.keyPlaceholder, secr.Key)
	e.key.Hide = func() bool {
		labels := labelsOf(e.typ.Value())
		return labels.key == "" || labels.keyArea
	}
	e.keyArea = form.NewArea(labels.key, labels.keyPlaceholder, secr.Key, 6)
	e.keyArea.Hide = func() bool {
		labels := labelsOf(e.typ.Value())
		return labels.key == "" || !labels.keyArea
	}

	e.value = form.NewText(labels.value, labels.valPlaceholder, secr.Value)
	e.value.Hide = func() bool {
		return labelsOf(e.typ.Value()).value == ""
	}

	e.region = form.NewText(labels.region, "Region", secr.Region)
	e.region.Hide = func() bool {
		return labelsOf(e.typ.Value()).region == ""
	}

	e.data = form.NewArea("JSON Data", "Secret JSON data", secr.Data, 8)
	e.data.Hide = func() bool {
		return e.typ.Value() != secret.Json
	}

	e.typ = form.NewSelect("Type", typeOptions, secr.Type)
	e.typ.OnChange = e.applyType

	orgOpts := resource.WithSelect(nodeSecret, resource.SelectOptions(orgs))
	e.organization = form.NewSelect("Organization", orgOpts,
		resource.IdHex(secr.Organization))

	e.publicKey = form.NewArea(labels.publicKey, "Oracle Cloud Public Key",
		secr.PublicKey, 6)
	e.publicKey.Hide = func() bool {
		return labelsOf(e.typ.Value()).publicKey == ""
	}
	e.publicKey.Lock = func() bool {
		return true
	}

	e.form = form.New(
		e.name,
		e.comment,
		e.key,
		e.keyArea,
		e.value,
		e.region,
		e.data,
		e.typ,
		e.organization,
		e.publicKey,
	)

	return e
}

// applyType updates the labels of the key inputs for the selected type
// like the web interface.
func (e *editor) applyType() {
	labels := labelsOf(e.typ.Value())
	e.key.Label = labels.key
	e.keyArea.Label = labels.key
	e.value.Label = labels.value
	e.region.Label = labels.region
	e.publicKey.Label = labels.publicKey
}

func (e *editor) Form() *form.Form {
	return e.form
}

// keyValue returns the key from the text or area input shown for the
// type, the two inputs hold the same field.
func (e *editor) keyValue() string {
	if labelsOf(e.typ.Value()).keyArea {
		return e.keyArea.Value()
	}
	return e.key.Value()
}

// apply sets the form values on the secret.
func (e *editor) apply(secr *secret.Secret) {
	secr.Name = e.name.Value()
	secr.Comment = e.comment.Value()
	secr.Organization = resource.ParseId(e.organization.Value())
	secr.Type = e.typ.Value()
	secr.Key = e.keyValue()
	secr.Value = e.value.Value()
	secr.Data = e.data.Value()
	secr.Region = e.region.Value()
}

// Save applies the form to the current secret the same way as the admin
// secret handler, or inserts a new secret.
func (e *editor) Save(db *database.Database) (err error) {
	var secr *secret.Secret
	if e.create {
		secr = &secret.Secret{}
	} else {
		secr, err = secret.Get(db, e.secr.Id)
		if err != nil {
			return
		}
	}

	e.apply(secr)

	errData, err := secr.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = &errortypes.ParseError{
			errors.New("secrets: " + errData.Message),
		}
		return
	}

	if e.create {
		err = secr.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"secret_id": secr.Id.Hex(),
		}).Info("tui: Secret created")
	} else {
		err = secr.CommitFields(db, commitFields)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"secret_id": secr.Id.Hex(),
		}).Info("tui: Secret settings saved")
	}

	err = event.PublishDispatch(db, "secret.change")
	if err != nil {
		return
	}

	return
}
