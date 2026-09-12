package firewalls

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

// commitFields are the fields written by the admin firewall handler.
var commitFields = set.NewSet(
	"name",
	"comment",
	"organization",
	"roles",
	"ingress",
)

var protocolOptions = []widget.SelectOption{
	{Label: "All Traffic", Value: firewall.All},
	{Label: "ICMP", Value: firewall.Icmp},
	{Label: "TCP", Value: firewall.Tcp},
	{Label: "UDP", Value: firewall.Udp},
	{Label: "Multicast", Value: firewall.Multicast},
	{Label: "Broadcast", Value: firewall.Broadcast},
}

// editor is the firewall settings form mirroring the inputs of the
// detailed firewall view in the web interface.
type editor struct {
	fire   *firewall.Firewall
	create bool
	form   *form.Form

	name         *form.Text
	comment      *form.Area
	organization *form.Select
	roles        *form.Tokens
	ingress      *form.Rows
}

// ruleRow creates the protocol, port and source IP inputs of an ingress
// rule.
func ruleRow(rule *firewall.Rule) *form.Row {
	protocol := form.NewSelect("", protocolOptions, rule.Protocol)
	protocol.Compact = true

	sources := form.NewTokens("", "Source IP range", rule.SourceIps)

	return &form.Row{
		Cells: []form.Input{
			protocol,
			form.NewCell("Port range", rule.Port),
			sources,
		},
		Tag: rule,
	}
}

func newRuleRow() *form.Row {
	row := ruleRow(&firewall.Rule{
		Protocol:  firewall.All,
		SourceIps: []string{},
	})
	row.Tag = nil
	return row
}

func newEditor(fire *firewall.Firewall, orgs []*database.Named) *editor {
	e := &editor{
		fire: fire,
	}

	e.name = form.NewText("Name", "Enter name", fire.Name)
	e.comment = form.NewArea("Comment", "Firewall comment", fire.Comment, 4)

	// Firewalls without an organization are node firewalls
	orgOpts := []widget.SelectOption{{Label: nodeFirewall, Value: ""}}
	orgOpts = append(orgOpts, resource.SelectOptions(orgs)...)
	orgValue := ""
	if !fire.Organization.IsZero() {
		orgValue = fire.Organization.Hex()
	}
	e.organization = form.NewSelect("Organization", orgOpts, orgValue)

	e.roles = form.NewTokens("Roles", "Add role", fire.Roles)

	e.ingress = form.NewRows("Ingress Rules", "No ingress rules",
		"Add Rule", []string{"Protocol", "Port", "Sources"}, newRuleRow)
	for _, rule := range fire.Ingress {
		e.ingress.Add(ruleRow(rule))
	}

	e.form = form.New(
		e.name,
		e.comment,
		e.organization,
		e.roles,
		e.ingress,
	)

	return e
}

func (e *editor) Form() *form.Form {
	return e.form
}

func (e *editor) buildIngress() []*firewall.Rule {
	rules := []*firewall.Rule{}
	for _, row := range e.ingress.Rows() {
		protocol := row.Cells[0].(*form.Select).Value()
		port := row.Cells[1].(*form.Text).Value()
		if protocol == firewall.All || protocol == firewall.Icmp {
			port = ""
		}

		rules = append(rules, &firewall.Rule{
			Protocol:  protocol,
			Port:      port,
			SourceIps: row.Cells[2].(*form.Tokens).Values(),
		})
	}
	return rules
}

// apply sets the form values on the firewall.
func (e *editor) apply(fire *firewall.Firewall) {
	fire.Name = e.name.Value()
	fire.Comment = e.comment.Value()
	fire.Organization, _ = utils.ParseObjectId(e.organization.Value())
	fire.Roles = e.roles.Values()
	fire.Ingress = e.buildIngress()
}

// Save applies the form to the current firewall the same way as the
// admin firewall handler, or inserts a new firewall.
func (e *editor) Save(db *database.Database) (err error) {
	var fire *firewall.Firewall
	if e.create {
		fire = &firewall.Firewall{}
	} else {
		fire, err = firewall.Get(db, e.fire.Id)
		if err != nil {
			return
		}
	}

	e.apply(fire)

	errData, err := fire.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = &errortypes.ParseError{
			errors.New("firewalls: " + errData.Message),
		}
		return
	}

	if e.create {
		err = fire.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"firewall_id": fire.Id.Hex(),
		}).Info("tui: Firewall created")
	} else {
		err = fire.CommitFields(db, commitFields)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"firewall_id": fire.Id.Hex(),
		}).Info("tui: Firewall settings saved")
	}

	err = event.PublishDispatch(db, "firewall.change")
	if err != nil {
		return
	}

	return
}

// newFirewall returns the default new firewall of the web interface, an
// ICMP rule and an SSH rule open to everyone.
func newFirewall() *firewall.Firewall {
	return &firewall.Firewall{
		Name:    "new-firewall",
		Comment: "22/tcp - SSH connections",
		Roles:   []string{},
		Ingress: []*firewall.Rule{
			{
				Protocol:  firewall.Icmp,
				SourceIps: []string{"0.0.0.0/0", "::/0"},
			},
			{
				Protocol:  firewall.Tcp,
				Port:      "22",
				SourceIps: []string{"0.0.0.0/0", "::/0"},
			},
		},
	}
}
