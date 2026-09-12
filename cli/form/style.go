package form

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/pritunl/pritunl-cloud/cli/widget"
)

var (
	labelStyle      = widget.LabelStyle
	labelFocusStyle = lipgloss.NewStyle().
			Foreground(widget.ColorPrimary).
			Bold(true)

	// Row cells have a panel background in place of a border
	cellStyle = lipgloss.NewStyle().
			Background(widget.ColorPanel)
	dividerStyle = lipgloss.NewStyle().
			Foreground(widget.ColorDim)
	readOnlyStyle = lipgloss.NewStyle().
			Foreground(widget.ColorMuted)

	toggleOffStyle = lipgloss.NewStyle().
			Foreground(widget.ColorMuted).
			Background(lipgloss.Color("#E5E7EB")).
			Padding(0, 1).
			MarginRight(1)
	toggleOnStyle = lipgloss.NewStyle().
			Foreground(widget.ColorWhite).
			Background(widget.ColorPrimary).
			Padding(0, 1).
			MarginRight(1)

	selectValueStyle = lipgloss.NewStyle().
				Foreground(widget.ColorWhite).
				Background(widget.ColorGray).
				Padding(0, 1)
	selectValueFocusStyle = lipgloss.NewStyle().
				Foreground(widget.ColorWhite).
				Background(widget.ColorPrimary).
				Padding(0, 1)
	selectArrowStyle = lipgloss.NewStyle().
				Foreground(widget.ColorMuted)
	selectArrowFocusStyle = lipgloss.NewStyle().
				Foreground(widget.ColorPrimary).
				Bold(true)

	tokenStyle = lipgloss.NewStyle().
			Foreground(widget.ColorWhite).
			Background(widget.ColorGray).
			Padding(0, 1).
			MarginRight(1)
	tokenSelectedStyle = tokenStyle.
				Background(widget.ColorPrimary)

	addButtonStyle = lipgloss.NewStyle().
			Foreground(widget.ColorWhite).
			Background(widget.ColorGreen).
			Padding(0, 1)
	removeButtonStyle = lipgloss.NewStyle().
				Foreground(widget.ColorWhite).
				Background(widget.ColorRed).
				Padding(0, 1)
	buttonFocusStyle = lipgloss.NewStyle().
				Foreground(widget.ColorPrimary).
				Background(widget.ColorWhite).
				Bold(true).
				Padding(0, 1)
)

func renderLabel(label string, focused bool) string {
	if focused {
		return labelFocusStyle.Render(label)
	}
	return labelStyle.Render(label)
}

// keyIs returns true when the key press is one of the keys.
func keyIs(msg tea.KeyPressMsg, keys ...string) bool {
	name := msg.String()
	for _, key := range keys {
		if name == key {
			return true
		}
	}
	return false
}

// navigationKey returns true for keys the form uses to move between
// inputs, inputs never consume them.
func navigationKey(msg tea.KeyPressMsg) bool {
	return keyIs(msg, "tab", "shift+tab", "up", "down", "esc")
}
