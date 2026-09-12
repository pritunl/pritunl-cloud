package iface

import (
	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
	NextTab key.Binding
	PrevTab key.Binding
	Quit    key.Binding
}

var keys = KeyMap{
	NextTab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
	),
	PrevTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "previous tab"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}
