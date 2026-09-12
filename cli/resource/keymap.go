package resource

import (
	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	PrevPage key.Binding
	NextPage key.Binding
	First    key.Binding
	Last     key.Binding
	Expand   key.Binding
	Collapse key.Binding
	Filter   key.Binding
	Refresh  key.Binding
	New      key.Binding
	Edit     key.Binding
	Save     key.Binding
	Cancel   key.Binding
}

var keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	PrevPage: key.NewBinding(
		key.WithKeys("left", "pgup", "h", "["),
		key.WithHelp("←", "previous page"),
	),
	NextPage: key.NewBinding(
		key.WithKeys("right", "pgdown", "l", "]"),
		key.WithHelp("→", "next page"),
	),
	First: key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("home", "first page"),
	),
	Last: key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("end", "last page"),
	),
	Expand: key.NewBinding(
		key.WithKeys("enter", "i"),
		key.WithHelp("enter", "expand"),
	),
	Collapse: key.NewBinding(
		key.WithKeys("esc", "backspace"),
		key.WithHelp("esc", "collapse"),
	),
	Filter: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "filter"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("R", "ctrl+r"),
		key.WithHelp("R", "refresh"),
	),
	New: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit settings"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("ctrl+z"),
		key.WithHelp("ctrl+z", "cancel changes"),
	),
}

// DetailKeyMap are the keys of the expanded resource view, up and down
// scroll the detail and left and right move between items across pages.
type DetailKeyMap struct {
	ScrollUp   key.Binding
	ScrollDown key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
	Prev       key.Binding
	Next       key.Binding
}

var detailKeys = DetailKeyMap{
	ScrollUp: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑", "scroll up"),
	),
	ScrollDown: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓", "scroll down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("pgup", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("pgdn", "page down"),
	),
	Prev: key.NewBinding(
		key.WithKeys("left", "h", "["),
		key.WithHelp("←", "previous item"),
	),
	Next: key.NewBinding(
		key.WithKeys("right", "l", "]"),
		key.WithHelp("→", "next item"),
	),
}
