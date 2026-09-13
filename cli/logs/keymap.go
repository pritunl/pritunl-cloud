package logs

import (
	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
	Clear   key.Binding
	Top     key.Binding
	End     key.Binding
	Refresh key.Binding
}

var keys = KeyMap{
	Clear: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "clear"),
	),
	Top: key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("home", "top"),
	),
	End: key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("end", "end"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("R", "ctrl+r"),
		key.WithHelp("R", "refresh"),
	),
}
