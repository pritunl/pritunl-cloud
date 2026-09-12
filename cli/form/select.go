package form

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/pritunl/pritunl-cloud/cli/widget"
)

const (
	selectLabelWidth = 26
)

// Select cycles through choices with the left and right keys, Dynamic
// options are computed on each use for choices that depend on another
// input such as subnets on the VPC.
type Select struct {
	Base
	Options []widget.SelectOption
	Dynamic func() []widget.SelectOption
	Compact bool

	// OnChange runs after the value changes by key or click.
	OnChange func()

	value      string
	initial    string
	focused    bool
	labelWidth int
}

func NewSelect(label string, options []widget.SelectOption,
	value string) *Select {

	return &Select{
		Base:    Base{Label: label},
		Options: options,
		value:   value,
		initial: value,
	}
}

func (s *Select) options() []widget.SelectOption {
	if s.Dynamic != nil {
		return s.Dynamic()
	}
	return s.Options
}

// index returns the index of the current value or -1.
func (s *Select) index(opts []widget.SelectOption) int {
	for i, opt := range opts {
		if opt.Value == s.value {
			return i
		}
	}
	return -1
}

// resolve returns the value when it is a choice and the first choice
// otherwise.
func (s *Select) resolve(value string) string {
	opts := s.options()
	for _, opt := range opts {
		if opt.Value == value {
			return value
		}
	}
	if len(opts) > 0 {
		return opts[0].Value
	}
	return ""
}

// Value returns the selected value, the first choice when the current
// value is not a choice.
func (s *Select) Value() string {
	return s.resolve(s.value)
}

func (s *Select) SelectedLabel() string {
	opts := s.options()
	i := s.index(opts)
	if i == -1 && len(opts) > 0 {
		i = 0
	}
	if i == -1 {
		return ""
	}
	return opts[i].Label
}

// Changed compares against the resolved initial value so a stored value
// that is not a choice does not count as a change.
func (s *Select) Changed() bool {
	return s.Value() != s.resolve(s.initial)
}

func (s *Select) cycle(dir int) {
	opts := s.options()
	n := len(opts)
	if n == 0 {
		return
	}

	i := s.index(opts)
	if i == -1 {
		i = 0
	} else {
		i = ((i+dir)%n + n) % n
	}

	changed := s.value != opts[i].Value
	s.value = opts[i].Value
	if changed && s.OnChange != nil {
		s.OnChange()
	}
}

// SetValue sets the value without running the change hook.
func (s *Select) SetValue(value string) {
	s.value = value
}

func (s *Select) Focused() bool {
	return s.focused
}

func (s *Select) Focus(reverse bool) tea.Cmd {
	s.focused = true
	return nil
}

func (s *Select) Blur() {
	s.focused = false
}

func (s *Select) Next(reverse bool) (tea.Cmd, bool) {
	return nil, false
}

func (s *Select) Update(msg tea.Msg) (tea.Cmd, bool) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil, false
	}

	switch {
	case keyIs(keyMsg, "left"):
		s.cycle(-1)
		return nil, true
	case keyIs(keyMsg, "right", "space", "enter"):
		s.cycle(1)
		return nil, true
	}

	return nil, false
}

// Click cycles backward on the left arrow and forward elsewhere on the
// value.
func (s *Select) Click(x, y int) tea.Cmd {
	if x < s.labelWidth {
		return nil
	}
	if x < s.labelWidth+2 {
		s.cycle(-1)
	} else {
		s.cycle(1)
	}
	return nil
}

func (s *Select) View(width int) string {
	var left, right, value string
	if s.focused {
		left = selectArrowFocusStyle.Render("◀ ")
		right = selectArrowFocusStyle.Render(" ▶")
	} else {
		left = selectArrowStyle.Render("◀ ")
		right = selectArrowStyle.Render(" ▶")
	}

	if s.Compact {
		s.labelWidth = 0
		text := widget.Truncate(max(width-6, 3), "%s", s.SelectedLabel())
		if s.focused {
			value = selectValueFocusStyle.Render(text)
		} else {
			value = selectValueStyle.Render(text)
		}
		return lipgloss.JoinHorizontal(lipgloss.Top, left, value, right)
	}

	s.labelWidth = min(selectLabelWidth, max(width/2, 8))
	text := widget.Truncate(max(width-s.labelWidth-6, 5), "%s",
		s.SelectedLabel())
	if s.focused {
		value = selectValueFocusStyle.Render(text)
	} else {
		value = selectValueStyle.Render(text)
	}

	label := renderLabel(
		widget.Truncate(s.labelWidth-1, "%s", s.Label), s.focused)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(s.labelWidth).Render(label),
		left,
		value,
		right,
	)
}
