package widget

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	optionButtonStyle = lipgloss.NewStyle().
				Foreground(ColorWhite).
				Background(ColorPrimary).
				Padding(0, 3).
				MarginTop(1).
				MarginRight(2)
	optionButtonActiveStyle = optionButtonStyle.
				Foreground(ColorPrimary).
				Background(ColorWhite).
				Underline(true)

	// Destructive buttons are red
	optionButtonDangerStyle = optionButtonStyle.
				Background(ColorRed)
	optionButtonDangerActiveStyle = optionButtonActiveStyle.
					Foreground(ColorRed)

	optionLabelStyle = lipgloss.NewStyle().
				Foreground(ColorLabel)
	optionLabelActiveStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)

	toggleOffStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Background(lipgloss.Color("#E5E7EB")).
			Padding(0, 1).
			MarginRight(1)
	toggleOnStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Background(ColorPrimary).
			Padding(0, 1).
			MarginRight(1)

	selectValueStyle = lipgloss.NewStyle().
				Foreground(ColorWhite).
				Background(ColorGray).
				Padding(0, 1)
	selectValueActiveStyle = lipgloss.NewStyle().
				Foreground(ColorWhite).
				Background(ColorPrimary).
				Padding(0, 1)
	selectArrowStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)
	selectArrowActiveStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)
)

// Option is a single field or button in a Dialog.
type Option interface {
	// Init prepares the option for display in a dialog of the given
	// content width.
	Init(width int)

	// SetWidth resizes the option to the dialog content width.
	SetWidth(width int)

	// Footer options are rendered horizontally at the bottom of the dialog.
	Footer() bool
	Update(tea.Msg) tea.Cmd
	Focused() bool
	Focus() tea.Cmd
	Unfocus()

	// OnEnter returns the dialog return value, whether the dialog should
	// close with that value and whether the key was handled. Unhandled
	// enter presses activate the default button.
	OnEnter() (ret int, close bool, handled bool)

	// OnSpace returns true when the key was consumed by the option.
	OnSpace() bool
	View() string
}

type OptionText struct {
	Label       string
	Placeholder string
	Value       string
	Password    bool
	model       textinput.Model
}

func (o *OptionText) Init(width int) {
	o.model = textinput.New()
	o.model.Placeholder = o.Placeholder
	o.model.CharLimit = 2048
	o.model.SetWidth(max(width-4, 10))
	o.model.Prompt = "> "
	if o.Password {
		o.model.EchoMode = textinput.EchoPassword
		o.model.EchoCharacter = '•'
	}
	if o.Value != "" {
		o.model.SetValue(o.Value)
	}
}

func (o *OptionText) SetWidth(width int) {
	o.model.SetWidth(max(width-4, 10))
}

func (o *OptionText) Footer() bool {
	return false
}

func (o *OptionText) Update(msg tea.Msg) (cmd tea.Cmd) {
	o.model, cmd = o.model.Update(msg)
	return
}

func (o *OptionText) Focused() bool {
	return o.model.Focused()
}

func (o *OptionText) Focus() (cmd tea.Cmd) {
	cmd = o.model.Focus()
	return
}

func (o *OptionText) Unfocus() {
	o.model.Blur()
}

func (o *OptionText) OnEnter() (int, bool, bool) {
	return 0, false, false
}

func (o *OptionText) OnSpace() bool {
	return false
}

func (o *OptionText) View() string {
	var label string
	if o.Focused() {
		label = optionLabelActiveStyle.Render(o.Label)
	} else {
		label = optionLabelStyle.Render(o.Label)
	}
	return lipgloss.JoinVertical(lipgloss.Left, label, o.model.View())
}

func (o *OptionText) GetValue() string {
	return strings.TrimSpace(o.model.Value())
}

type OptionToggle struct {
	Label   string
	Value   bool
	focused bool
}

func (o *OptionToggle) Init(width int) {
}

func (o *OptionToggle) SetWidth(width int) {
}

func (o *OptionToggle) Footer() bool {
	return false
}

func (o *OptionToggle) Update(msg tea.Msg) (cmd tea.Cmd) {
	return
}

func (o *OptionToggle) Focused() bool {
	return o.focused
}

func (o *OptionToggle) Focus() (cmd tea.Cmd) {
	o.focused = true
	return
}

func (o *OptionToggle) Unfocus() {
	o.focused = false
}

func (o *OptionToggle) Toggle() {
	o.Value = !o.Value
}

func (o *OptionToggle) OnEnter() (int, bool, bool) {
	o.Toggle()
	return 0, false, true
}

func (o *OptionToggle) OnSpace() bool {
	o.Toggle()
	return true
}

func (o *OptionToggle) GetValue() bool {
	return o.Value
}

func (o *OptionToggle) View() string {
	var state string
	if o.Value {
		state = toggleOnStyle.Render("ON ")
	} else {
		state = toggleOffStyle.Render("OFF")
	}

	var label string
	if o.focused {
		label = optionLabelActiveStyle.Render(o.Label)
	} else {
		label = optionLabelStyle.Render(o.Label)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, state, label)
}

// SelectOption is a choice in an OptionSelect.
type SelectOption struct {
	Label string
	Value string
}

var selectKeys = struct {
	Prev key.Binding
	Next key.Binding
}{
	Prev: key.NewBinding(
		key.WithKeys("left", "h"),
	),
	Next: key.NewBinding(
		key.WithKeys("right", "l"),
	),
}

// OptionSelect cycles through a fixed list of choices with the left and
// right keys, it replaces the select inputs of the web interface.
type OptionSelect struct {
	Label   string
	Options []SelectOption
	Index   int
	focused bool
	width   int
}

func (o *OptionSelect) Init(width int) {
	o.width = width
	if o.Index < 0 || o.Index >= len(o.Options) {
		o.Index = 0
	}
}

func (o *OptionSelect) SetWidth(width int) {
	o.width = width
}

func (o *OptionSelect) Footer() bool {
	return false
}

func (o *OptionSelect) cycle(dir int) {
	n := len(o.Options)
	if n == 0 {
		return
	}
	o.Index = ((o.Index+dir)%n + n) % n
}

func (o *OptionSelect) Update(msg tea.Msg) (cmd tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return
	}

	switch {
	case key.Matches(keyMsg, selectKeys.Prev):
		o.cycle(-1)
	case key.Matches(keyMsg, selectKeys.Next):
		o.cycle(1)
	}

	return
}

func (o *OptionSelect) Focused() bool {
	return o.focused
}

func (o *OptionSelect) Focus() (cmd tea.Cmd) {
	o.focused = true
	return
}

func (o *OptionSelect) Unfocus() {
	o.focused = false
}

func (o *OptionSelect) OnEnter() (int, bool, bool) {
	o.cycle(1)
	return 0, false, true
}

func (o *OptionSelect) OnSpace() bool {
	o.cycle(1)
	return true
}

// SetValue selects the option with the value, the first option is
// selected when no option matches.
func (o *OptionSelect) SetValue(value string) {
	o.Index = 0
	for i, opt := range o.Options {
		if opt.Value == value {
			o.Index = i
			return
		}
	}
}

func (o *OptionSelect) GetValue() string {
	if o.Index < 0 || o.Index >= len(o.Options) {
		return ""
	}
	return o.Options[o.Index].Value
}

func (o *OptionSelect) GetLabel() string {
	if o.Index < 0 || o.Index >= len(o.Options) {
		return ""
	}
	return o.Options[o.Index].Label
}

func (o *OptionSelect) View() string {
	// Label column keeps the values aligned across stacked selects
	labelWidth := 24
	valueWidth := max(o.width-labelWidth-6, 10)
	labelText := Truncate(labelWidth-1, "%s", o.Label)
	valueText := Truncate(valueWidth, "%s", o.GetLabel())

	var label, value, left, right string
	if o.focused {
		label = optionLabelActiveStyle.Render(labelText)
		value = selectValueActiveStyle.Render(valueText)
		left = selectArrowActiveStyle.Render("◀ ")
		right = selectArrowActiveStyle.Render(" ▶")
	} else {
		label = optionLabelStyle.Render(labelText)
		value = selectValueStyle.Render(valueText)
		left = selectArrowStyle.Render("◀ ")
		right = selectArrowStyle.Render(" ▶")
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(labelWidth).Render(label),
		left,
		value,
		right,
	)
}

type OptionButton struct {
	Label  string
	Return int

	// Danger renders the button red, buttons that remove or clear are
	// always red.
	Danger  bool
	focused bool
}

func (o *OptionButton) Init(width int) {
}

func (o *OptionButton) SetWidth(width int) {
}

func (o *OptionButton) Footer() bool {
	return true
}

func (o *OptionButton) Update(msg tea.Msg) (cmd tea.Cmd) {
	return
}

func (o *OptionButton) Focused() bool {
	return o.focused
}

func (o *OptionButton) Focus() (cmd tea.Cmd) {
	o.focused = true
	return
}

func (o *OptionButton) Unfocus() {
	o.focused = false
}

func (o *OptionButton) OnEnter() (int, bool, bool) {
	return o.Return, true, true
}

func (o *OptionButton) OnSpace() bool {
	return false
}

// isDanger returns true for red buttons.
func (o *OptionButton) isDanger() bool {
	if o.Danger {
		return true
	}
	label := strings.ToLower(o.Label)
	return strings.Contains(label, "remove") ||
		strings.Contains(label, "delete") ||
		strings.Contains(label, "clear")
}

func (o *OptionButton) View() string {
	if o.isDanger() {
		if o.focused {
			return optionButtonDangerActiveStyle.Render(o.Label)
		}
		return optionButtonDangerStyle.Render(o.Label)
	}

	if o.focused {
		return optionButtonActiveStyle.Render(o.Label)
	}
	return optionButtonStyle.Render(o.Label)
}
