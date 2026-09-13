package policies

import (
	"image/color"
	"strings"

	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/policy"
)

// Item is a policy card with the two-factor providers for the editor.
type Item struct {
	polcy     *policy.Policy
	providers []widget.SelectOption
}

func (i *Item) Policy() *policy.Policy {
	return i.polcy
}

func (i *Item) Id() string {
	return i.polcy.Id.Hex()
}

func (i *Item) Name() string {
	return i.polcy.Name
}

func (i *Item) Tag() string {
	return i.polcy.Id.Hex()
}

func enabledColor(disabled bool) color.Color {
	if disabled {
		return widget.ColorRed
	}
	return widget.ColorGreen
}

func (i *Item) Fields() []resource.Field {
	return []resource.Field{
		{
			Label: "State",
			Value: widget.Bool(i.polcy.Disabled, "Disabled", "Enabled"),
			Color: enabledColor(i.polcy.Disabled),
		},
		{
			Label: "Roles",
			Value: strings.Join(i.polcy.Roles, ", "),
		},
		{
			Label: "Comment",
			Value: i.polcy.Comment,
		},
	}
}

// Info mirrors the fields of the detailed policy view.
func (i *Item) Info() []widget.InfoField {
	return []widget.InfoField{
		{"ID", i.polcy.Id.Hex()},
	}
}

// Actions mirrors the delete button of the detailed policy view, the
// admin handler checks the relations before removing.
func (i *Item) Actions() []resource.Action {
	return []resource.Action{
		resource.DeleteAction("policy", i.polcy.Name,
			resource.DeleteRelated("policy", "policy", "policy.change",
				i.polcy.Id, policy.Remove)),
	}
}

func (i *Item) Editor() resource.Editor {
	return newEditor(i.polcy, i.providers)
}
