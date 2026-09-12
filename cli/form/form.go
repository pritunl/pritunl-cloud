// Package form provides the settings inputs shown below the information
// of an expanded resource, the inputs mirror the page inputs of the web
// interface and are laid out vertically with keyboard and mouse focus.
package form

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Input is a form field, composite inputs such as Rows manage the focus
// of their own cells and report through Next when the focus leaves them.
type Input interface {
	// Hidden inputs are skipped in the layout and focus order.
	Hidden() bool

	Focused() bool

	// Focus enters the input, reverse enters from the last cell.
	Focus(reverse bool) tea.Cmd
	Blur()

	// Next moves the focus within the input and returns false when the
	// focus leaves the input.
	Next(reverse bool) (tea.Cmd, bool)

	// Update handles a message while focused and returns whether the
	// message was consumed, navigation keys are left to the form.
	Update(msg tea.Msg) (tea.Cmd, bool)

	// Click handles a mouse click at the position relative to the input.
	Click(x, y int) tea.Cmd

	// ReadOnly inputs are rendered but skipped in the focus order and
	// ignore clicks, used for values fixed by another input such as a
	// shape.
	ReadOnly() bool

	View(width int) string

	// Changed returns true when the value differs from the initial value.
	Changed() bool
}

// Base holds the label, hide and lock conditions shared by inputs.
type Base struct {
	Label string

	// Hide returns true when the input is hidden, used for inputs that
	// depend on another input such as secure boot on UEFI.
	Hide func() bool

	// Lock returns true when the input is read only, used for values
	// fixed by another input such as the memory of a shape.
	Lock func() bool
}

func (b *Base) Hidden() bool {
	return b.Hide != nil && b.Hide()
}

func (b *Base) ReadOnly() bool {
	return b.Lock != nil && b.Lock()
}

// span is the rendered position of an input in the form view.
type span struct {
	index int
	y     int
	h     int
}

type Form struct {
	inputs []Input
	focus  int
	spans  []span
}

func New(inputs ...Input) *Form {
	return &Form{
		inputs: inputs,
		focus:  -1,
	}
}

// Editing returns true while an input is focused.
func (f *Form) Editing() bool {
	return f.focus >= 0 && f.focus < len(f.inputs)
}

func (f *Form) Changed() bool {
	for _, in := range f.inputs {
		if in.Changed() {
			return true
		}
	}
	return false
}

// nextVisible returns the next visible input index in the direction or
// -1 when there is none.
func (f *Form) nextVisible(from int, reverse bool) int {
	step := 1
	if reverse {
		step = -1
	}
	for i := from + step; i >= 0 && i < len(f.inputs); i += step {
		if !f.inputs[i].Hidden() && !f.inputs[i].ReadOnly() {
			return i
		}
	}
	return -1
}

// Focus enters the form on the first input.
func (f *Form) Focus() tea.Cmd {
	if f.Editing() {
		return nil
	}
	index := f.nextVisible(-1, false)
	if index == -1 {
		return nil
	}
	f.focus = index
	return f.inputs[index].Focus(false)
}

// Blur leaves the form.
func (f *Form) Blur() {
	if f.Editing() {
		f.inputs[f.focus].Blur()
	}
	f.focus = -1
}

func (f *Form) focusInput(index int, reverse bool) tea.Cmd {
	if f.Editing() {
		f.inputs[f.focus].Blur()
	}
	f.focus = index
	return f.inputs[index].Focus(reverse)
}

// moveFocus moves within the focused input or to the next input.
func (f *Form) moveFocus(reverse bool) tea.Cmd {
	if !f.Editing() {
		return f.Focus()
	}

	cmd, moved := f.inputs[f.focus].Next(reverse)
	if moved {
		return cmd
	}

	index := f.nextVisible(f.focus, reverse)
	if index == -1 {
		return nil
	}

	return f.focusInput(index, reverse)
}

// Update delivers the message to the focused input, unconsumed
// navigation keys move the focus and esc leaves the form.
func (f *Form) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !f.Editing() {
		return nil, false
	}

	in := f.inputs[f.focus]
	if in.Hidden() || in.ReadOnly() {
		return f.moveFocus(false), true
	}

	cmd, consumed := in.Update(msg)
	if consumed {
		return cmd, true
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return cmd, false
	}

	switch keyMsg.String() {
	case "tab", "down":
		return f.moveFocus(false), true
	case "shift+tab", "up":
		return f.moveFocus(true), true
	case "esc":
		f.Blur()
		return nil, true
	}

	return cmd, false
}

// stacked inputs are single line and rendered without a gap between
// consecutive stacked inputs.
func stacked(in Input) bool {
	switch in.(type) {
	case *Toggle, *Select:
		return true
	}
	return false
}

// View renders the visible inputs and records their positions for
// mouse clicks and scrolling.
func (f *Form) View(width int) string {
	f.spans = []span{}
	parts := []string{}
	y := 0
	prevStacked := false

	for i, in := range f.inputs {
		if in.Hidden() {
			continue
		}

		if len(parts) > 0 && !(prevStacked && stacked(in)) {
			parts = append(parts, "")
			y += 1
		}

		view := in.View(width)
		h := lipgloss.Height(view)
		f.spans = append(f.spans, span{
			index: i,
			y:     y,
			h:     h,
		})
		parts = append(parts, view)
		y += h
		prevStacked = stacked(in)
	}

	return strings.Join(parts, "\n")
}

// FocusRange returns the position of the focused input in the last view.
func (f *Form) FocusRange() (y, h int, ok bool) {
	if !f.Editing() {
		return 0, 0, false
	}
	for _, sp := range f.spans {
		if sp.index == f.focus {
			return sp.y, sp.h, true
		}
	}
	return 0, 0, false
}

// Click focuses and clicks the input at the position relative to the
// last view.
func (f *Form) Click(x, y int) tea.Cmd {
	for _, sp := range f.spans {
		if y < sp.y || y >= sp.y+sp.h {
			continue
		}
		if f.inputs[sp.index].ReadOnly() {
			return nil
		}

		cmds := []tea.Cmd{}
		if f.focus != sp.index {
			cmds = append(cmds, f.focusInput(sp.index, false))
		}
		cmds = append(cmds, f.inputs[sp.index].Click(x, y-sp.y))
		return tea.Batch(cmds...)
	}

	return nil
}
