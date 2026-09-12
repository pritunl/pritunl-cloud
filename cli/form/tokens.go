package form

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/pritunl/pritunl-cloud/cli/widget"
)

// tokenHit is the rendered position of a token for mouse clicks.
type tokenHit struct {
	index int
	y     int
	x     int
	w     int
}

// Tokens is a list of removable text tokens with an input to add new
// tokens, used for roles. Enter adds the input text, backspace on an
// empty input removes the last token and left selects tokens to remove
// with delete.
type Tokens struct {
	Base
	Placeholder string

	// Options replaces the text input with a select of choices added
	// with enter, the token labels come from the options.
	Options []widget.SelectOption

	values    []string
	initial   []string
	model     textinput.Model
	selected  int
	optIndex  int
	hits      []tokenHit
	inputLine int
}

func NewTokens(label, placeholder string, values []string) *Tokens {
	return &Tokens{
		Base:        Base{Label: label},
		Placeholder: placeholder,
		values:      append([]string{}, values...),
		initial:     append([]string{}, values...),
		model:       newTextModel(placeholder, "", "> "),
		selected:    -1,
	}
}

func (t *Tokens) Values() []string {
	return append([]string{}, t.values...)
}

func (t *Tokens) Changed() bool {
	if len(t.values) != len(t.initial) {
		return true
	}
	for i := range t.values {
		if t.values[i] != t.initial[i] {
			return true
		}
	}
	return false
}

func (t *Tokens) Focused() bool {
	return t.model.Focused()
}

func (t *Tokens) Focus(reverse bool) tea.Cmd {
	t.selected = -1
	return t.model.Focus()
}

func (t *Tokens) Blur() {
	t.selected = -1
	t.model.Blur()
}

func (t *Tokens) Next(reverse bool) (tea.Cmd, bool) {
	return nil, false
}

// NewTokensSelect creates a token list where tokens are chosen from
// the options.
func NewTokensSelect(label string, options []widget.SelectOption,
	values []string) *Tokens {

	t := NewTokens(label, "", values)
	t.Options = options
	return t
}

// tokenLabel returns the option label of a token value.
func (t *Tokens) tokenLabel(value string) string {
	for _, opt := range t.Options {
		if opt.Value == value {
			return opt.Label
		}
	}
	return value
}

func (t *Tokens) cycle(dir int) {
	n := len(t.Options)
	if n == 0 {
		return
	}
	t.optIndex = ((t.optIndex+dir)%n + n) % n
}

func (t *Tokens) add() {
	var value string
	if t.Options != nil {
		if t.optIndex < 0 || t.optIndex >= len(t.Options) {
			return
		}
		value = t.Options[t.optIndex].Value
	} else {
		value = strings.TrimSpace(t.model.Value())
	}
	if value == "" {
		return
	}

	for _, cur := range t.values {
		if cur == value {
			t.model.SetValue("")
			return
		}
	}

	t.values = append(t.values, value)
	t.model.SetValue("")
}

func (t *Tokens) remove(index int) {
	if index < 0 || index >= len(t.values) {
		return
	}
	t.values = append(t.values[:index], t.values[index+1:]...)
	if t.selected >= len(t.values) {
		t.selected = len(t.values) - 1
	}
}

func (t *Tokens) Update(msg tea.Msg) (tea.Cmd, bool) {
	keyMsg, isKey := msg.(tea.KeyPressMsg)
	if !isKey {
		var cmd tea.Cmd
		t.model, cmd = t.model.Update(msg)
		return cmd, false
	}

	if navigationKey(keyMsg) {
		return nil, false
	}

	if t.Options != nil {
		return nil, t.updateSelect(keyMsg)
	}

	empty := t.model.Value() == ""

	switch {
	case keyIs(keyMsg, "enter"):
		t.add()
		return nil, true
	case keyIs(keyMsg, "backspace") && empty:
		if t.selected >= 0 {
			t.remove(t.selected)
		} else {
			t.remove(len(t.values) - 1)
		}
		return nil, true
	case keyIs(keyMsg, "delete") && t.selected >= 0:
		t.remove(t.selected)
		return nil, true
	case keyIs(keyMsg, "left") && empty:
		if t.selected == -1 {
			t.selected = len(t.values) - 1
		} else {
			t.selected = max(t.selected-1, 0)
		}
		return nil, true
	case keyIs(keyMsg, "right") && empty && t.selected >= 0:
		t.selected++
		if t.selected >= len(t.values) {
			t.selected = -1
		}
		return nil, true
	}

	t.selected = -1
	var cmd tea.Cmd
	t.model, cmd = t.model.Update(msg)
	return cmd, true
}

// updateSelect handles keys in the select mode, left and right cycle
// the choice, enter adds it and backspace removes the last token.
func (t *Tokens) updateSelect(keyMsg tea.KeyPressMsg) bool {
	switch {
	case keyIs(keyMsg, "enter", "space"):
		t.add()
	case keyIs(keyMsg, "left"):
		t.cycle(-1)
	case keyIs(keyMsg, "right"):
		t.cycle(1)
	case keyIs(keyMsg, "backspace", "delete"):
		if t.selected >= 0 {
			t.remove(t.selected)
		} else {
			t.remove(len(t.values) - 1)
		}
	default:
		return false
	}
	return true
}

// selectView renders the choice line of the select mode.
func (t *Tokens) selectView(width int) string {
	if len(t.Options) == 0 {
		return widget.MutedStyle.Render(
			widget.Default(t.Placeholder, "No options"))
	}
	if t.optIndex < 0 || t.optIndex >= len(t.Options) {
		t.optIndex = 0
	}

	focused := t.Focused()
	text := widget.Truncate(max(width-16, 5), "%s",
		t.Options[t.optIndex].Label)

	var left, right, value, hint string
	if focused {
		left = selectArrowFocusStyle.Render("◀ ")
		right = selectArrowFocusStyle.Render(" ▶")
		value = selectValueFocusStyle.Render(text)
		hint = widget.MutedStyle.Render("  enter: add")
	} else {
		left = selectArrowStyle.Render("◀ ")
		right = selectArrowStyle.Render(" ▶")
		value = selectValueStyle.Render(text)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, left, value, right, hint)
}

// Click removes a token when its close mark is clicked, selects a token
// otherwise and returns to the input on the input line.
func (t *Tokens) Click(x, y int) tea.Cmd {
	for _, hit := range t.hits {
		if y != hit.y || x < hit.x || x >= hit.x+hit.w {
			continue
		}
		if x >= hit.x+hit.w-2 {
			t.remove(hit.index)
			t.selected = -1
		} else {
			t.selected = hit.index
		}
		return nil
	}

	if y == t.inputLine {
		t.selected = -1
		if t.Options != nil {
			if x < 2 {
				t.cycle(-1)
			} else {
				t.cycle(1)
			}
		}
	}
	return nil
}

func (t *Tokens) View(width int) string {
	t.hits = []tokenHit{}
	focused := t.Focused()

	// Row cells have no label line
	lines := []string{}
	if t.Label != "" {
		lines = append(lines, renderLabel(t.Label, focused))
	}

	// Tokens flow onto new lines at the width
	line := ""
	x := 0
	for i, value := range t.values {
		style := tokenStyle
		if focused && i == t.selected {
			style = tokenSelectedStyle
		}
		text := style.Render(t.tokenLabel(value) + " ×")
		w := lipgloss.Width(text)

		if x > 0 && x+w > width {
			lines = append(lines, line)
			line = ""
			x = 0
		}

		t.hits = append(t.hits, tokenHit{
			index: i,
			y:     len(lines),
			x:     x,
			w:     w - style.GetMarginRight(),
		})
		line += text
		x += w
	}
	if line != "" {
		lines = append(lines, line)
	}

	t.inputLine = len(lines)
	if t.Options != nil {
		lines = append(lines, t.selectView(width))
	} else {
		t.model.SetWidth(max(width-lipgloss.Width(t.model.Prompt)-1, 5))
		lines = append(lines, t.model.View())
	}

	return strings.Join(lines, "\n")
}
