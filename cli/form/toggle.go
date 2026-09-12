package form

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Toggle is an on off switch.
type Toggle struct {
	Base

	// OnChange runs after the user flips the toggle by key or click.
	OnChange func()

	value   bool
	initial bool
	focused bool
}

func NewToggle(label string, value bool) *Toggle {
	return &Toggle{
		Base:    Base{Label: label},
		value:   value,
		initial: value,
	}
}

func (t *Toggle) Value() bool {
	return t.value
}

func (t *Toggle) Changed() bool {
	return t.value != t.initial
}

func (t *Toggle) Focused() bool {
	return t.focused
}

func (t *Toggle) Focus(reverse bool) tea.Cmd {
	t.focused = true
	return nil
}

func (t *Toggle) Blur() {
	t.focused = false
}

func (t *Toggle) Next(reverse bool) (tea.Cmd, bool) {
	return nil, false
}

// SetValue sets the toggle without running the change hook.
func (t *Toggle) SetValue(value bool) {
	t.value = value
}

func (t *Toggle) flip() {
	t.value = !t.value
	if t.OnChange != nil {
		t.OnChange()
	}
}

func (t *Toggle) Update(msg tea.Msg) (tea.Cmd, bool) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if ok && keyIs(keyMsg, "space", "enter") {
		t.flip()
		return nil, true
	}
	return nil, false
}

func (t *Toggle) Click(x, y int) tea.Cmd {
	t.flip()
	return nil
}

func (t *Toggle) View(width int) string {
	var state string
	if t.value {
		state = toggleOnStyle.Render("ON ")
	} else {
		state = toggleOffStyle.Render("OFF")
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		state,
		renderLabel(t.Label, t.focused),
	)
}
