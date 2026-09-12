package resource

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/cli/view"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

const (
	// detailFrameHeight is the title line, the blank line above the
	// buttons and the button row inside the expanded card border.
	detailFrameHeight = 3
	// detailSplitWidth is the window width from which the expanded fields
	// are shown in two columns.
	detailSplitWidth = 70
	// detailContentLeft is the column of the card content, the border
	// and padding.
	detailContentLeft = 2
)

var (
	buttonStyle = lipgloss.NewStyle().
			Foreground(widget.ColorWhite).
			Background(widget.ColorPrimary).
			Padding(0, 1).
			MarginRight(1)
	buttonSuccessStyle = buttonStyle.
				Background(widget.ColorGreen)
	buttonDangerStyle = buttonStyle.
				Background(widget.ColorRed)
	buttonMutedStyle = buttonStyle.
				Background(widget.ColorGray)

	sectionStyle = lipgloss.NewStyle().
			Foreground(widget.ColorAccent).
			Bold(true)
)

// actionDoneMsg reports the result of a resource action.
type actionDoneMsg struct {
	provider Provider
	action   Action
	name     string
	err      error
}

// saveDoneMsg reports the result of saving the settings form.
type saveDoneMsg struct {
	provider Provider
	name     string
	created  bool
	err      error
}

// newEditorMsg carries the settings form of a new resource.
type newEditorMsg struct {
	provider Provider
	editor   Editor
	err      error
}

// draftItem is the card of a resource being created, it has no info
// fields or actions and its editor inserts the resource.
type draftItem struct {
	title  string
	editor Editor
}

func (d *draftItem) Id() string               { return "new" }
func (d *draftItem) Name() string             { return d.title }
func (d *draftItem) Tag() string              { return "" }
func (d *draftItem) Fields() []Field          { return nil }
func (d *draftItem) Info() []widget.InfoField { return nil }
func (d *draftItem) Actions() []Action        { return nil }
func (d *draftItem) Editor() Editor           { return d.editor }

// newEditorCmd loads the settings form of a new resource in the
// background.
func (l *List) newEditorCmd() tea.Cmd {
	creator, ok := l.creator()
	if !ok {
		return nil
	}
	provider := l.provider

	return func() tea.Msg {
		db := database.GetDatabase()
		if db == nil {
			return newEditorMsg{
				provider: provider,
				err: &errortypes.DatabaseError{
					errors.New("resource: Database not connected"),
				},
			}
		}
		defer db.Close()

		editor, err := creator.NewEditor(db)
		return newEditorMsg{
			provider: provider,
			editor:   editor,
			err:      err,
		}
	}
}

// startCreate shows the new resource card with its form focused.
func (l *List) startCreate(editor Editor) tea.Cmd {
	l.draft = &draftItem{
		title:  "New " + l.provider.Singular(),
		editor: editor,
	}
	l.creating = true
	l.expanded = true
	l.expandedId = l.draft.Id()
	l.editor = editor
	l.editorId = l.draft.Id()
	l.detail.SetYOffset(0)

	cmd := editor.Form().Focus()
	l.refreshForm()
	return cmd
}

// button is a rendered card button and its column position relative to
// the card content area.
type button struct {
	text  string
	x     int
	width int
	run   func() tea.Cmd
}

// buttons returns the expanded card buttons, the item actions, the
// settings edit buttons and the collapse button. Buttons that do not fit
// in the card width are dropped.
func (l *List) buttons(item Item) []button {
	buttons := []button{}
	x := 0
	maxWidth := l.detailContentWidth()

	add := func(style lipgloss.Style, label string, run func() tea.Cmd) {
		text := style.Render(label)
		width := lipgloss.Width(text)
		if x+width-style.GetMarginRight() > maxWidth {
			return
		}
		buttons = append(buttons, button{
			text:  text,
			x:     x,
			width: width - style.GetMarginRight(),
			run:   run,
		})
		x += width
	}

	for _, action := range item.Actions() {
		style := buttonSuccessStyle
		if action.Danger {
			style = buttonDangerStyle
		}
		action := action
		add(style, fmt.Sprintf("[%s] %s", action.Key, action.Label),
			func() tea.Cmd {
				return l.confirmAction(item, action)
			})
	}

	if l.editor != nil {
		frm := l.editor.Form()
		if frm.Editing() {
			add(buttonMutedStyle, "[esc] Done", func() tea.Cmd {
				frm.Blur()
				l.refreshForm()
				return nil
			})
		} else {
			add(buttonStyle, "[e] Edit", l.edit)
		}
		if l.creating {
			add(buttonSuccessStyle, "[^s] Create", l.save)
		} else if frm.Changed() {
			add(buttonSuccessStyle, "[^s] Save", l.save)
			add(buttonDangerStyle, "[^z] Cancel", l.cancelEdit)
		}
	}

	label := "[q] Collapse"
	if l.creating {
		label = "[q] Cancel"
	}
	if l.editing() {
		label = "Collapse"
		if l.creating {
			label = "Cancel"
		}
	}
	add(buttonMutedStyle, label, func() tea.Cmd {
		l.collapse()
		return nil
	})

	return buttons
}

func renderButtons(buttons []button) string {
	texts := make([]string, 0, len(buttons))
	for _, btn := range buttons {
		texts = append(texts, btn.text)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, texts...)
}

// buttonAt returns the button at the column relative to the card content
// area.
func buttonAt(buttons []button, x int) (button, bool) {
	for _, btn := range buttons {
		if x >= btn.x && x < btn.x+btn.width {
			return btn, true
		}
	}
	return button{}, false
}

// renderInfoColumn renders the fields wrapped to the column width.
func renderInfoColumn(fields []widget.InfoField, width int) string {
	style := lipgloss.NewStyle().Width(width)
	lines := make([]string, 0, len(fields))
	for _, field := range fields {
		lines = append(lines, style.Render(
			widget.LabelStyle.Render(field.Label+":")+" "+
				widget.Default(field.Value, "-")))
	}
	return strings.Join(lines, "\n")
}

// renderInfo renders the fields in two columns on wide windows like the
// expanded resource of the web interface.
func renderInfo(fields []widget.InfoField, width int) string {
	if width < detailSplitWidth || len(fields) < 2 {
		return renderInfoColumn(fields, width)
	}

	// Two columns with a gap, the left column gets the odd field
	gap := 2
	rightWidth := (width - gap) / 2
	leftWidth := width - gap - rightWidth
	split := (len(fields) + 1) / 2

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderInfoColumn(fields[:split], leftWidth),
		strings.Repeat(" ", gap),
		renderInfoColumn(fields[split:], rightWidth),
	)
}

// detailContentWidth is the width inside the expanded card border and
// padding.
func (l *List) detailContentWidth() int {
	frameX, _ := cardSelectedStyle.GetFrameSize()
	return max(l.width-frameX, 20)
}

// ensureEditor keeps the settings form while it is focused or has
// changes and rebuilds it from the loaded item otherwise.
func (l *List) ensureEditor(item Item) {
	if l.editor != nil && l.editorId == item.Id() {
		frm := l.editor.Form()
		if frm.Editing() || frm.Changed() || l.saving {
			return
		}
	}

	l.editor = item.Editor()
	l.editorId = item.Id()
}

// refreshDetail rebuilds the expanded card viewport for the selected
// item and the current size.
func (l *List) refreshDetail() {
	item := l.selected()
	if !l.expanded || item == nil {
		return
	}

	l.ensureEditor(item)

	width := l.detailContentWidth()
	content := ""
	l.formTop = -1

	// New resource cards have no info fields
	parts := []string{}
	if len(item.Info()) > 0 {
		parts = append(parts, renderInfo(item.Info(), width), "")
	}

	if l.editor != nil {
		parts = append(parts, sectionStyle.Render("Settings"))
		l.formTop = 0
		for _, part := range parts {
			l.formTop += lipgloss.Height(part)
		}
		parts = append(parts, l.editor.Form().View(width))
	}
	if len(parts) > 0 {
		content = lipgloss.JoinVertical(lipgloss.Left, parts...)
	}

	_, frameY := cardSelectedStyle.GetFrameSize()
	l.detail.SetWidth(width)
	l.detail.SetHeight(max(l.height-frameY-detailFrameHeight, 1))
	l.detail.SetContent(content)
}

// refreshForm rebuilds the expanded card after input to the settings
// form and keeps the focused input visible. Background reloads use
// refreshDetail so they never move the scroll position.
func (l *List) refreshForm() {
	l.refreshDetail()
	l.scrollToFocus()
}

// scrollToFocus scrolls the focused settings input into view.
func (l *List) scrollToFocus() {
	if !l.editing() || l.formTop < 0 {
		return
	}

	y, h, ok := l.editor.Form().FocusRange()
	if !ok {
		return
	}
	y += l.formTop

	if y-1 < l.detail.YOffset() {
		l.detail.SetYOffset(max(y-1, 0))
	} else if y+h > l.detail.YOffset()+l.detail.Height() {
		l.detail.SetYOffset(y + h - l.detail.Height())
	}
}

// expand opens the selected card in place of the list.
func (l *List) expand() {
	item := l.selected()
	if item == nil {
		return
	}
	l.expanded = true
	l.expandedId = item.Id()
	l.editor = nil
	l.detail.SetYOffset(0)
	l.refreshDetail()
}

func (l *List) collapse() {
	l.expanded = false
	l.expandedId = ""
	l.editor = nil
	l.creating = false
	l.draft = nil
}

// edit focuses the first settings input.
func (l *List) edit() tea.Cmd {
	if l.editor == nil {
		return nil
	}
	cmd := l.editor.Form().Focus()
	l.refreshForm()
	return cmd
}

// cancelEdit discards the settings changes.
func (l *List) cancelEdit() tea.Cmd {
	if l.editor == nil || l.saving || l.creating {
		return nil
	}
	changed := l.editor.Form().Changed()
	l.editor = nil
	l.refreshDetail()
	if changed {
		return view.Status("Changes discarded", false)
	}
	return nil
}

// save commits the settings form in the background.
func (l *List) save() tea.Cmd {
	if l.editor == nil || l.saving {
		return nil
	}
	if !l.creating && !l.editor.Form().Changed() {
		return view.Status("No changes to save", true)
	}

	item := l.selected()
	if item == nil {
		return nil
	}

	l.saving = true
	editor := l.editor
	provider := l.provider
	name := item.Name()
	created := l.creating

	return func() tea.Msg {
		db := database.GetDatabase()
		if db == nil {
			return saveDoneMsg{
				provider: provider,
				name:     name,
				err: &errortypes.DatabaseError{
					errors.New("resource: Database not connected"),
				},
			}
		}
		defer db.Close()

		return saveDoneMsg{
			provider: provider,
			name:     name,
			created:  created,
			err:      editor.Save(db),
		}
	}
}

func (l *List) updateSaveDone(msg saveDoneMsg) tea.Cmd {
	if msg.provider != l.provider {
		return nil
	}
	l.saving = false

	if msg.err != nil {
		if msg.created {
			return view.Error("Create "+l.provider.Singular(), msg.err)
		}
		return view.Error("Save "+msg.name, msg.err)
	}

	// Rebuild the form from the saved item on the next load, a created
	// resource returns to the list
	if l.editor != nil {
		l.editor.Form().Blur()
	}
	l.editor = nil

	status := "Saved " + msg.name
	if msg.created {
		l.collapse()
		status = "Created " + l.provider.Singular()
	}

	return tea.Batch(
		view.Status(status, false),
		l.load(),
	)
}

// renderDetail renders the expanded card filling the list body with the
// info fields, the settings form and the buttons.
func (l *List) renderDetail(item Item) string {
	// Style width includes the padding and border
	style := cardSelectedStyle
	contentWidth := l.detailContentWidth()
	cardWidth := contentWidth + style.GetHorizontalPadding() +
		style.GetBorderLeftSize() + style.GetBorderRightSize()

	name := widget.Truncate(contentWidth, "%s", item.Name())
	title := widget.TitleStyle.Render(name)
	tag := item.Tag()
	if tag != "" {
		tagWidth := contentWidth - len([]rune(name)) - 2
		if tagWidth > 3 {
			title += widget.MutedStyle.Render(
				"  " + widget.Truncate(tagWidth, "%s", tag))
		}
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		l.detail.View(),
		"",
		renderButtons(l.buttons(item)),
	)

	return style.Width(cardWidth).Render(content)
}

// confirmAction opens the confirm dialog of the action and runs it when
// confirmed.
func (l *List) confirmAction(item Item, action Action) tea.Cmd {
	provider := l.provider
	name := item.Name()

	message := action.Confirm
	if message == "" {
		message = fmt.Sprintf("%s %s %s?",
			action.Label, provider.Singular(), name)
	}

	return view.Dialog(widget.NewConfirmDialog(
		fmt.Sprintf("%s %s", action.Label, name),
		message,
		action.Label,
		action.Danger,
	), func(ret int) tea.Cmd {
		if ret != widget.DialogOk {
			return nil
		}

		return func() tea.Msg {
			db := database.GetDatabase()
			if db == nil {
				return actionDoneMsg{
					provider: provider,
					action:   action,
					name:     name,
					err: &errortypes.DatabaseError{
						errors.New("resource: Database not connected"),
					},
				}
			}
			defer db.Close()

			return actionDoneMsg{
				provider: provider,
				action:   action,
				name:     name,
				err:      action.Run(db),
			}
		}
	})
}

// actionByKey returns the action of the expanded item bound to the key.
func actionByKey(item Item, keyMsg tea.KeyPressMsg) (Action, bool) {
	if item == nil || keyMsg.Text == "" {
		return Action{}, false
	}
	for _, action := range item.Actions() {
		if action.Key == keyMsg.String() {
			return action, true
		}
	}
	return Action{}, false
}

func (l *List) updateActionDone(msg actionDoneMsg) tea.Cmd {
	if msg.provider != l.provider {
		return nil
	}

	if msg.err != nil {
		return view.Error(msg.action.Label+" "+msg.name, msg.err)
	}

	status := msg.action.Status
	if status == "" {
		status = msg.action.Label
	}

	return tea.Batch(
		view.Status(status+" "+msg.name, false),
		l.load(),
	)
}
