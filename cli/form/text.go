package form

import (
	"strconv"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/cli/widget"
)

// Text is a single line text input, compact inputs have no label and
// are used as row cells.
type Text struct {
	Base
	Compact bool
	model   textinput.Model
	initial string
}

func newTextModel(placeholder, value, prompt string) textinput.Model {
	m := textinput.New()
	m.Placeholder = placeholder
	m.CharLimit = 4096
	m.Prompt = prompt
	m.SetValue(value)
	return m
}

func NewText(label, placeholder, value string) *Text {
	return &Text{
		Base:    Base{Label: label},
		model:   newTextModel(placeholder, value, "> "),
		initial: value,
	}
}

// NewCell creates a compact text input for a row cell.
func NewCell(placeholder, value string) *Text {
	return &Text{
		Compact: true,
		model:   newTextModel(placeholder, value, ""),
		initial: value,
	}
}

func (t *Text) Value() string {
	return strings.TrimSpace(t.model.Value())
}

func (t *Text) Changed() bool {
	return t.Value() != strings.TrimSpace(t.initial)
}

func (t *Text) Focused() bool {
	return t.model.Focused()
}

func (t *Text) Focus(reverse bool) tea.Cmd {
	return t.model.Focus()
}

func (t *Text) Blur() {
	t.model.Blur()
}

func (t *Text) Next(reverse bool) (tea.Cmd, bool) {
	return nil, false
}

func (t *Text) Update(msg tea.Msg) (tea.Cmd, bool) {
	keyMsg, isKey := msg.(tea.KeyPressMsg)
	if isKey && (navigationKey(keyMsg) || keyIs(keyMsg, "enter")) {
		return nil, false
	}

	var cmd tea.Cmd
	t.model, cmd = t.model.Update(msg)
	return cmd, isKey
}

func (t *Text) Click(x, y int) tea.Cmd {
	return nil
}

// SetValue replaces the text keeping the initial value for change
// detection.
func (t *Text) SetValue(value string) {
	t.model.SetValue(value)
	t.model.CursorEnd()
}

func (t *Text) View(width int) string {
	if t.ReadOnly() {
		// Fixed values render as muted text without a prompt
		field := readOnlyStyle.Render(widget.Truncate(width, "%s",
			widget.Default(t.model.Value(), "-")))
		if t.Compact {
			return field
		}
		return renderLabel(t.Label, false) + "\n" + field
	}

	t.model.SetWidth(max(width-lipgloss.Width(t.model.Prompt)-1, 5))
	field := t.model.View()
	if t.Compact {
		return field
	}
	return renderLabel(t.Label, t.Focused()) + "\n" + field
}

// Number is a text input accepting digits only, plus and minus step the
// value.
type Number struct {
	Text
	Min     int
	Max     int
	Step    int
	initial int
}

func numberText(value int) string {
	if value == 0 {
		return ""
	}
	return strconv.Itoa(value)
}

func validateDigits(value string) error {
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return errors.New("form: Digits only")
		}
	}
	return nil
}

func NewNumber(label, placeholder string, value, min, max,
	step int) *Number {

	n := &Number{
		Text:    *NewText(label, placeholder, numberText(value)),
		Min:     min,
		Max:     max,
		Step:    step,
		initial: value,
	}
	n.model.Validate = validateDigits
	return n
}

// NewNumberCell creates a compact number input for a row cell.
func NewNumberCell(placeholder string, value int) *Number {
	n := &Number{
		Text:    *NewCell(placeholder, numberText(value)),
		initial: value,
	}
	n.model.Validate = validateDigits
	return n
}

func (n *Number) Int() int {
	value, err := strconv.Atoi(n.Value())
	if err != nil {
		return 0
	}
	return value
}

func (n *Number) Changed() bool {
	return n.Int() != n.initial
}

func (n *Number) set(value int) {
	if n.Max > 0 {
		value = min(value, n.Max)
	}
	value = max(value, n.Min)
	n.model.SetValue(numberText(value))
	n.model.CursorEnd()
}

// Set replaces the number clamped to the range.
func (n *Number) Set(value int) {
	n.set(value)
}

func (n *Number) Update(msg tea.Msg) (tea.Cmd, bool) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok && n.Step > 0 {
		switch {
		case keyIs(keyMsg, "+", "="):
			n.set(n.Int() + n.Step)
			return nil, true
		case keyIs(keyMsg, "-", "_"):
			n.set(n.Int() - n.Step)
			return nil, true
		}
	}
	return n.Text.Update(msg)
}
