package policies

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/policy"
	"github.com/sirupsen/logrus"
)

// commitFields are the fields written by the admin policy handler.
var commitFields = set.NewSet(
	"name",
	"comment",
	"disabled",
	"roles",
	"rules",
	"admin_secondary",
	"user_secondary",
	"admin_device_secondary",
	"user_device_secondary",
)

// ruleInputs are the inputs of one policy rule mirroring the web rule
// component: a switch enabling the rule, the disable user switch and
// the permitted values chosen from a list or typed for networks.
type ruleInputs struct {
	typ     string
	enabled *form.Toggle
	disable *form.Toggle
	values  *form.Tokens
}

// newRule creates the inputs of a rule type with the labels of the web
// rule component, options nil means values are typed networks.
func newRule(typ, toggleLabel, valuesLabel string,
	options []widget.SelectOption, rule *policy.Rule) *ruleInputs {

	r := &ruleInputs{
		typ: typ,
	}

	enabled := rule != nil
	values := []string{}
	disable := false
	if enabled {
		values = rule.Values
		disable = rule.Disable
	}

	r.enabled = form.NewToggle(toggleLabel, enabled)

	r.disable = form.NewToggle("Disabled user on failure", disable)
	r.disable.Hide = func() bool {
		return !r.enabled.Value()
	}

	if options != nil {
		r.values = form.NewTokensSelect(valuesLabel, options, values)
	} else {
		r.values = form.NewTokens(valuesLabel, "Add network", values)
	}
	r.values.Hide = r.disable.Hide

	return r
}

func (r *ruleInputs) inputs() []form.Input {
	return []form.Input{r.enabled, r.disable, r.values}
}

// rule returns the rule or nil when the rule is off.
func (r *ruleInputs) rule() *policy.Rule {
	if !r.enabled.Value() {
		return nil
	}
	return &policy.Rule{
		Type:    r.typ,
		Disable: r.disable.Value(),
		Values:  r.values.Values(),
	}
}

// secondaryInputs are the two-factor switch and provider select of the
// admin or user login.
type secondaryInputs struct {
	enabled  *form.Toggle
	provider *form.Select
}

// hasOption reports whether the provider id is one of the providers.
func hasOption(opts []widget.SelectOption, value string) bool {
	for _, opt := range opts {
		if opt.Value == value {
			return true
		}
	}
	return false
}

// newSecondary creates the two-factor inputs, the switch is on when
// the provider is set and still exists like the web interface. The
// switch cannot be turned on without providers.
func newSecondary(toggleLabel, selectLabel string,
	providers []widget.SelectOption, current bson.ObjectID) *secondaryInputs {

	s := &secondaryInputs{}

	value := resource.IdHex(current)
	enabled := value != "" && hasOption(providers, value)

	s.enabled = form.NewToggle(toggleLabel, enabled)

	opts := providers
	if len(opts) == 0 {
		opts = []widget.SelectOption{{Label: "None", Value: ""}}
	}
	s.provider = form.NewSelect(selectLabel, opts, value)
	s.provider.Hide = func() bool {
		return !s.enabled.Value()
	}

	s.enabled.OnChange = func() {
		if s.enabled.Value() && len(providers) == 0 {
			s.enabled.SetValue(false)
		}
	}

	return s
}

func (s *secondaryInputs) inputs() []form.Input {
	return []form.Input{s.enabled, s.provider}
}

// id returns the provider id or the zero id when two-factor is off.
func (s *secondaryInputs) id() bson.ObjectID {
	if !s.enabled.Value() {
		return bson.NilObjectID
	}
	return resource.ParseId(s.provider.Value())
}

// editor is the policy settings form mirroring the inputs of the
// detailed policy view in the web interface.
type editor struct {
	polcy  *policy.Policy
	create bool
	form   *form.Form

	name              *form.Text
	comment           *form.Area
	roles             *form.Tokens
	adminSecondary    *secondaryInputs
	userSecondary     *secondaryInputs
	whitelistNetworks *ruleInputs
	blacklistNetworks *ruleInputs
	enabled           *form.Toggle
	location          *ruleInputs
	operatingSystem   *ruleInputs
	browser           *ruleInputs
	adminDevice       *form.Toggle
	userDevice        *form.Toggle
}

func newEditor(polcy *policy.Policy, providers []widget.SelectOption) *editor {
	e := &editor{
		polcy: polcy,
	}

	rules := polcy.Rules
	if rules == nil {
		rules = map[string]*policy.Rule{}
	}

	e.name = form.NewText("Name", "Enter name", polcy.Name)
	e.comment = form.NewArea("Comment", "Policy comment", polcy.Comment, 4)
	e.roles = form.NewTokens("Roles", "Add role", polcy.Roles)

	e.adminSecondary = newSecondary("Admin two-factor authentication",
		"Admin Two-Factor Provider", providers, polcy.AdminSecondary)
	e.userSecondary = newSecondary("User two-factor authentication",
		"User Two-Factor Provider", providers, polcy.UserSecondary)

	e.whitelistNetworks = newRule(policy.WhitelistNetworks,
		"Permitted network policies", "Permitted Networks", nil,
		rules[policy.WhitelistNetworks])
	e.blacklistNetworks = newRule(policy.BlacklistNetworks,
		"Blocked network policies", "Blocked Networks", nil,
		rules[policy.BlacklistNetworks])

	e.enabled = form.NewToggle("Enabled", !polcy.Disabled)

	e.location = newRule(policy.Location, "Location policies",
		"Permitted Locations", locationOptions, rules[policy.Location])
	e.operatingSystem = newRule(policy.OperatingSystem,
		"Operating system policies", "Permitted Operating Systems",
		operatingSystemOptions, rules[policy.OperatingSystem])
	e.browser = newRule(policy.Browser, "Browser policies",
		"Permitted Browsers", browserOptions, rules[policy.Browser])

	e.adminDevice = form.NewToggle("Admin WebAuthn device authentication",
		polcy.AdminDeviceSecondary)
	e.userDevice = form.NewToggle("User WebAuthn device authentication",
		polcy.UserDeviceSecondary)

	inputs := []form.Input{
		e.name,
		e.comment,
		e.roles,
	}
	inputs = append(inputs, e.adminSecondary.inputs()...)
	inputs = append(inputs, e.userSecondary.inputs()...)
	inputs = append(inputs, e.whitelistNetworks.inputs()...)
	inputs = append(inputs, e.blacklistNetworks.inputs()...)
	inputs = append(inputs, e.enabled)
	inputs = append(inputs, e.location.inputs()...)
	inputs = append(inputs, e.operatingSystem.inputs()...)
	inputs = append(inputs, e.browser.inputs()...)
	inputs = append(inputs, e.adminDevice, e.userDevice)

	e.form = form.New(inputs...)

	return e
}

func (e *editor) Form() *form.Form {
	return e.form
}

// buildRules returns the enabled rules keyed by type like the web
// interface.
func (e *editor) buildRules() map[string]*policy.Rule {
	rules := map[string]*policy.Rule{}
	for _, r := range []*ruleInputs{
		e.whitelistNetworks,
		e.blacklistNetworks,
		e.location,
		e.operatingSystem,
		e.browser,
	} {
		if rule := r.rule(); rule != nil {
			rules[r.typ] = rule
		}
	}
	return rules
}

// apply sets the form values on the policy.
func (e *editor) apply(polcy *policy.Policy) {
	polcy.Name = e.name.Value()
	polcy.Comment = e.comment.Value()
	polcy.Disabled = !e.enabled.Value()
	polcy.Roles = e.roles.Values()
	polcy.Rules = e.buildRules()
	polcy.AdminSecondary = e.adminSecondary.id()
	polcy.UserSecondary = e.userSecondary.id()
	polcy.AdminDeviceSecondary = e.adminDevice.Value()
	polcy.UserDeviceSecondary = e.userDevice.Value()
}

// Save applies the form to the current policy the same way as the
// admin policy handler, or inserts a new policy.
func (e *editor) Save(db *database.Database) (err error) {
	var polcy *policy.Policy
	if e.create {
		polcy = &policy.Policy{}
	} else {
		polcy, err = policy.Get(db, e.polcy.Id)
		if err != nil {
			return
		}
	}

	e.apply(polcy)

	errData, err := polcy.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = &errortypes.ParseError{
			errors.New("policies: " + errData.Message),
		}
		return
	}

	if e.create {
		err = polcy.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"policy_id": polcy.Id.Hex(),
		}).Info("tui: Policy created")
	} else {
		err = polcy.CommitFields(db, commitFields)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"policy_id": polcy.Id.Hex(),
		}).Info("tui: Policy settings saved")
	}

	err = event.PublishDispatch(db, "policy.change")
	if err != nil {
		return
	}

	return
}
