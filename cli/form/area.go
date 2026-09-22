package form

import (
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/pritunl/pritunl-cloud/cli/widget"
)

// Area is a multiline text input, the up and down keys leave the input
// at the first and last line. A copy button beside the label shows the
// text for copying with the terminal selection.
type Area struct {
	Base
	model   textarea.Model
	initial string

	// Position of the copy button in the label row of the last view
	copyX     int
	copyWidth int
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

// copyCmd shows the text on a plain screen for selecting and copying
// with the terminal.
func (a *Area) copyCmd() tea.Cmd {
	if a.Value() == "" {
		return nil
	}
	return widget.SelectText(a.Label, a.Value())
}

// copyAt returns true when the position is on the copy button.
func (a *Area) copyAt(x, y int) bool {
	return y == 0 && a.copyWidth > 0 &&
		x >= a.copyX && x < a.copyX+a.copyWidth
}

func (a *Area) Next(reverse bool) (tea.Cmd, bool) {
	return nil, false
}

func (a *Area) Update(msg tea.Msg) (tea.Cmd, bool) {
	keyMsg, isKey := msg.(tea.KeyPressMsg)
	if isKey {
		switch {
		case keyIs(keyMsg, "ctrl+y"):
			return a.copyCmd(), true
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
	if a.copyAt(x, y) {
		return a.copyCmd()
	}
	return nil
}

func (a *Area) copyButton() string {
	switch {
	case a.Value() == "":
		return ""
	case a.Focused():
		return copyButtonStyle.Render("[ctrl+y] Copy")
	}
	return copyButtonStyle.Render("Copy")
}

func (a *Area) View(width int) string {
	a.model.SetWidth(max(width, 10))

	label := renderLabel(a.Label, a.Focused())
	button := a.copyButton()
	a.copyX = lipgloss.Width(label) + 2
	a.copyWidth = lipgloss.Width(button)
	if button != "" {
		label += "  " + button
	}

	return label + "\n" + a.model.View()
}
