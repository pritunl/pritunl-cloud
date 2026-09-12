// Package resource provides the generic paged list view shared by every
// resource tab, a resource only implements Provider and Item.
package resource

import (
	"image/color"

	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
)

// Filter is the active filter terms keyed by FilterField key, empty
// values are removed so an empty filter matches everything.
type Filter map[string]string

func (f Filter) Clone() Filter {
	clone := Filter{}
	for key, val := range f {
		clone[key] = val
	}
	return clone
}

// FilterField describes one input of the resource filter dialog, fields
// with options are shown as a select otherwise as a text input.
type FilterField struct {
	Key         string
	Label       string
	Placeholder string
	Options     []widget.SelectOption
}

// Field is a label and value shown in the resource card body, the color
// is applied to the value when set.
type Field struct {
	Label string
	Value string
	Color color.Color
}

// Item is one resource entry rendered as a card in the list.
type Item interface {
	// Id is the resource id.
	Id() string

	// Name is the card title.
	Name() string

	// Tag is the muted text shown beside the title.
	Tag() string

	// Fields are the card body fields, every item of a provider must
	// return the number of fields given by Provider.FieldCount.
	Fields() []Field

	// Info are the detailed fields shown when the card is expanded.
	Info() []widget.InfoField

	// Actions are the buttons shown when the card is expanded, only the
	// actions valid for the current state are returned.
	Actions() []Action

	// Editor returns the settings form of the item shown below the info
	// when the card is expanded, nil for resources without settings. A
	// new editor is created for every edit session.
	Editor() Editor
}

// Creator is implemented by providers whose resources can be created
// from the list, the editor form is shown as a new card and its Save
// inserts the resource.
type Creator interface {
	// NewEditor returns the settings form of a new resource with the
	// default values, select options are loaded from the database.
	NewEditor(db *database.Database) (Editor, error)
}

// Editor holds the settings form of an expanded item and saves it.
type Editor interface {
	Form() *form.Form

	// Save commits the form values to the database, validation errors are
	// returned with a message for the user.
	Save(db *database.Database) error
}

// Action is a button of an expanded resource such as start or stop, every
// action is confirmed in a dialog before Run is called in the background.
type Action struct {
	// Key is the single key that triggers the action.
	Key string

	// Label is the button and confirm dialog label.
	Label string

	// Status is the status line prefix shown once the action is applied
	// such as Starting.
	Status string

	// Danger renders the button and confirm button red.
	Danger bool

	// Confirm is the confirm dialog message.
	Confirm string

	// Run applies the action.
	Run func(db *database.Database) error
}

// Page is one page of resource items and the total count of items
// matching the filter.
type Page struct {
	Items []Item
	Count int64
}

// Provider loads pages of a resource from the database.
type Provider interface {
	// Title is the tab title.
	Title() string

	// Empty is the text shown when no items match.
	Empty() string

	// Singular is the resource name used in messages such as instance.
	Singular() string

	// FieldCount is the number of card body fields of every item, the
	// list uses it to size the cards before loading.
	FieldCount() int

	// FilterFields returns the filter dialog fields, select options are
	// loaded from the database.
	FilterFields(db *database.Database) ([]FilterField, error)

	// Load returns the page of items matching the filter.
	Load(db *database.Database, filter Filter,
		page, pageCount int64) (*Page, error)
}
