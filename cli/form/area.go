package form

import (
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
)

// Area is a multiline text input, the up and down keys leave the input
// at the first and last line.
type Area struct {
	Base
	model   textarea.Model
	initial string
}

func NewArea(label, placeholder, value string, rows int) *Area {
	m := textarea.New()
	m.Placeholder = placeholder
	m.ShowLineNumbers = false
	m.CharLimit = 0
	m.SetHeight(max(rows, 1))
	m.SetValue(value)

	return &Area{
		Base:    Base{Label: label},
		model:   m,
		initial: value,
	}
}

func (a *Area) Value() string {
	return a.model.Value()
}

func (a *Area) Changed() bool {
	return a.Value() != a.initial
}

func (a *Area) Focused() bool {
	return a.model.Focused()
}

func (a *Area) Focus(reverse bool) tea.Cmd {
	return a.model.Focus()
}

func (a *Area) Blur() {
	a.model.Blur()
}

func (a *Area) Next(reverse bool) (tea.Cmd, bool) {
	return nil, false
}

func (a *Area) Update(msg tea.Msg) (tea.Cmd, bool) {
	keyMsg, isKey := msg.(tea.KeyPressMsg)
	if isKey {
		switch {
		case keyIs(keyMsg, "tab", "shift+tab", "esc"):
			return nil, false
		case keyIs(keyMsg, "up") && a.model.Line() == 0:
			return nil, false
		case keyIs(keyMsg, "down") &&
			a.model.Line() >= a.model.LineCount()-1:
			return nil, false
		}
	}

	var cmd tea.Cmd
	a.model, cmd = a.model.Update(msg)
	return cmd, isKey
}

func (a *Area) Click(x, y int) tea.Cmd {
	return nil
}

func (a *Area) View(width int) string {
	a.model.SetWidth(max(width, 10))
	return renderLabel(a.Label, a.Focused()) + "\n" + a.model.View()
}
