package resource

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/cli/view"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

const (
	// cardFrameHeight is the title line and top and bottom border.
	cardFrameHeight = 3
	// splitWidth is the window width from which card fields are shown in
	// two columns.
	splitWidth = 90
	// splitMaxWidth caps the two column layout width on wide windows.
	splitMaxWidth = 160

	cursorNone = -1
	cursorLast = -2
)

var (
	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(widget.ColorBorder).
			Padding(0, 1)
	cardSelectedStyle = cardStyle.
				BorderForeground(widget.ColorFocus)
	cardColStyle = lipgloss.NewStyle().
			Align(lipgloss.Left)
)

// loadMsg carries a loaded page, the sequence drops results of requests
// superseded by a newer request.
type loadMsg struct {
	provider Provider
	seq      int
	result   *Page
	err      error
}

// List is the paged card list of a resource, each page is a database
// query sized to the number of cards that fit in the window. The
// selected card can be expanded in place of the list showing the detail
// fields and action buttons like the expanded row of the web interface.
type List struct {
	provider Provider
	filter   Filter
	summary  string

	items     []Item
	count     int64
	page      int64
	pageCount int64
	cursor    int

	// pendingCursor is applied once the requested page loads
	pendingCursor int

	expanded   bool
	expandedId string
	detail     viewport.Model

	// editor is the settings form of the expanded item, kept while it
	// has changes or focus and rebuilt from the loaded item otherwise
	editor   Editor
	editorId string
	formTop  int
	saving   bool

	// draft is the new resource card shown in place of the list while
	// creating, its editor inserts the resource
	creating bool
	draft    *draftItem

	width   int
	height  int
	split   bool
	resized bool

	loading  bool
	loaded   bool
	loadErr  error
	seq      int
	lastSync time.Time
}

func NewList(provider Provider) *List {
	l := &List{
		provider:      provider,
		filter:        Filter{},
		items:         []Item{},
		pageCount:     1,
		pendingCursor: cursorNone,
	}
	l.detail = viewport.New(viewport.WithWidth(10), viewport.WithHeight(1))
	return l
}

func (l *List) Title() string {
	return l.provider.Title()
}

// bodyRows is the number of field rows in a card body.
func (l *List) bodyRows() int {
	rows := l.provider.FieldCount()
	if l.split {
		rows = (rows + 1) / 2
	}
	return rows
}

func (l *List) cardHeight() int {
	return l.bodyRows() + cardFrameHeight
}

func (l *List) pages() int64 {
	if l.count <= 0 || l.pageCount <= 0 {
		return 0
	}
	return (l.count + l.pageCount - 1) / l.pageCount
}

func (l *List) SetSize(width, height int) {
	l.width = width
	l.height = height
	l.split = width >= splitWidth
	l.refreshDetail()

	pageCount := int64(max(l.height/l.cardHeight(), 1))
	if pageCount == l.pageCount {
		return
	}

	// Keep the selected item visible by moving to the page holding it
	// at the new page size
	index := l.page*l.pageCount + int64(l.cursor)
	l.pageCount = pageCount
	l.page = index / pageCount
	l.cursor = int(index % pageCount)
	l.resized = true
}

// load requests the current page in the background.
func (l *List) load() tea.Cmd {
	l.seq++
	l.loading = true
	l.lastSync = time.Now()

	seq := l.seq
	provider := l.provider
	filter := l.filter.Clone()
	page := l.page
	pageCount := l.pageCount

	return func() tea.Msg {
		db := database.GetDatabase()
		if db == nil {
			return loadMsg{
				provider: provider,
				seq:      seq,
				err: &errortypes.DatabaseError{
					errors.New("resource: Database not connected"),
				},
			}
		}
		defer db.Close()

		result, err := provider.Load(db, filter, page, pageCount)
		return loadMsg{
			provider: provider,
			seq:      seq,
			result:   result,
			err:      err,
		}
	}
}

func (l *List) updateLoad(msg loadMsg) tea.Cmd {
	if msg.provider != l.provider || msg.seq != l.seq {
		return nil
	}
	l.loading = false

	if msg.err != nil {
		if l.loadErr == nil {
			view.LogError("Load "+l.provider.Title(), msg.err)
		}
		l.loadErr = msg.err
		return view.Status(
			"Load failed: "+widget.ErrorMessage(msg.err), true)
	}

	l.loadErr = nil
	l.loaded = true
	l.items = msg.result.Items
	l.count = msg.result.Count

	// Items were removed since the page was requested
	pages := l.pages()
	if pages > 0 && l.page >= pages {
		l.page = pages - 1
		return l.load()
	}

	switch l.pendingCursor {
	case cursorNone:
	case cursorLast:
		l.cursor = len(l.items) - 1
	default:
		l.cursor = l.pendingCursor
	}
	l.pendingCursor = cursorNone

	// Follow the expanded item when the page order changes
	if l.expanded && !l.creating {
		for i, item := range l.items {
			if item.Id() == l.expandedId {
				l.cursor = i
				break
			}
		}
	}

	if l.cursor >= len(l.items) {
		l.cursor = len(l.items) - 1
	}
	if l.cursor < 0 {
		l.cursor = 0
	}

	if l.creating {
		l.refreshDetail()
	} else if l.expanded {
		if item := l.selected(); item != nil {
			l.expandedId = item.Id()
			l.refreshDetail()
		} else {
			l.collapse()
		}
	}

	return nil
}

// moveCursor moves the selection flowing onto the previous or next page
// at the ends of the current page.
func (l *List) moveCursor(dir int) tea.Cmd {
	if len(l.items) == 0 || l.creating {
		return nil
	}

	next := l.cursor + dir
	if next < 0 {
		if l.page > 0 {
			l.page--
			l.pendingCursor = cursorLast
			return l.load()
		}
		return nil
	}

	if next >= len(l.items) {
		if l.page+1 < l.pages() {
			l.page++
			l.pendingCursor = 0
			return l.load()
		}
		return nil
	}

	l.cursor = next
	if l.expanded {
		l.expand()
	}
	return nil
}

func (l *List) gotoPage(page int64) tea.Cmd {
	pages := l.pages()
	if pages == 0 {
		return nil
	}

	page = max(min(page, pages-1), 0)
	if page == l.page {
		return nil
	}

	l.page = page
	if l.expanded {
		// Expanded item is on another page, expand the same position on
		// the new page
		l.expandedId = ""
	}
	return l.load()
}

func (l *List) selected() Item {
	if l.creating {
		return l.draft
	}
	if l.cursor < 0 || l.cursor >= len(l.items) {
		return nil
	}
	return l.items[l.cursor]
}

// creator returns the provider create hook when the resource can be
// created from the list.
func (l *List) creator() (Creator, bool) {
	creator, ok := l.provider.(Creator)
	return creator, ok
}

func (l *List) Menu() []widget.MenuItem {
	if l.editing() {
		menu := []widget.MenuItem{
			{Title: "Done", Key: "esc"},
			{Title: "Next Input", Key: "tab"},
		}
		if l.creating {
			return append(menu, widget.MenuItem{Title: "Create", Key: "^s"})
		}
		return append(menu,
			widget.MenuItem{Title: "Save", Key: "^s"},
			widget.MenuItem{Title: "Cancel", Key: "^z"},
		)
	}

	if l.creating {
		return []widget.MenuItem{
			{Title: "Edit", Key: "e"},
			{Title: "Create", Key: "^s"},
			{Title: "Scroll", Key: "↑/↓"},
		}
	}

	if l.expanded {
		menu := []widget.MenuItem{}
		if item := l.selected(); item != nil {
			for _, action := range item.Actions() {
				menu = append(menu, widget.MenuItem{
					Title: action.Label,
					Key:   action.Key,
				})
			}
		}
		if l.editor != nil {
			menu = append(menu, widget.MenuItem{Title: "Edit", Key: "e"})
			if l.editor.Form().Changed() {
				menu = append(menu,
					widget.MenuItem{Title: "Save", Key: "^s"},
					widget.MenuItem{Title: "Cancel", Key: "^z"},
				)
			}
		}
		return append(menu,
			widget.MenuItem{Title: "Scroll", Key: "↑/↓"},
			widget.MenuItem{Title: "Item", Key: "←/→"},
			widget.MenuItem{Title: "Refresh", Key: "R"},
		)
	}

	menu := []widget.MenuItem{
		{Title: "Expand", Key: "enter"},
	}
	if _, ok := l.creator(); ok {
		menu = append(menu, widget.MenuItem{Title: "New", Key: "n"})
	}
	return append(menu,
		widget.MenuItem{Title: "Filter", Key: "f"},
		widget.MenuItem{Title: "Select", Key: "↑/↓"},
		widget.MenuItem{Title: "Page", Key: "←/→"},
		widget.MenuItem{Title: "Refresh", Key: "R"},
	)
}

func (l *List) Status() string {
	if !l.loaded {
		if l.loadErr != nil {
			return "Load failed"
		}
		return "Loading..."
	}

	status := fmt.Sprintf(
		"Page %d/%d  %d %s",
		l.page+1,
		max(l.pages(), 1),
		l.count,
		strings.ToLower(l.provider.Title()),
	)
	if l.summary != "" {
		status += "  Filter: " + l.summary
	}

	return status
}

func (l *List) Update(msg tea.Msg) (view.View, tea.Cmd) {
	switch msg := msg.(type) {
	case view.ActivateMsg:
		l.resized = false
		return l, l.load()

	case view.ResizeMsg:
		if l.resized && (l.loaded || l.loading) {
			l.resized = false
			return l, l.load()
		}
		return l, nil

	case view.TickMsg:
		if !l.loading && time.Since(l.lastSync) >= view.SyncInterval {
			return l, l.load()
		}
		return l, nil

	case loadMsg:
		return l, l.updateLoad(msg)

	case actionDoneMsg:
		return l, l.updateActionDone(msg)

	case saveDoneMsg:
		return l, l.updateSaveDone(msg)

	case newEditorMsg:
		if msg.provider != l.provider {
			return l, nil
		}
		if msg.err != nil {
			return l, view.Error("New "+l.provider.Singular(), msg.err)
		}
		return l, l.startCreate(msg.editor)

	case filterFieldsMsg:
		if msg.provider != l.provider {
			return l, nil
		}
		if msg.err != nil {
			return l, view.Error("Load Filter", msg.err)
		}
		return l, l.filterDialog(msg.fields)

	case applyFilterMsg:
		if msg.provider != l.provider {
			return l, nil
		}
		l.filter = msg.filter
		l.summary = msg.summary
		l.page = 0
		l.cursor = 0
		l.collapse()
		return l, l.load()

	case tea.MouseMsg:
		if l.expanded {
			return l, l.updateDetailMouse(msg)
		}
		return l, l.updateMouse(msg)

	case tea.KeyPressMsg:
		if l.expanded {
			return l, l.updateDetailKey(msg)
		}

		switch {
		case key.Matches(msg, keys.Up):
			return l, l.moveCursor(-1)
		case key.Matches(msg, keys.Down):
			return l, l.moveCursor(1)
		case key.Matches(msg, keys.PrevPage):
			return l, l.gotoPage(l.page - 1)
		case key.Matches(msg, keys.NextPage):
			return l, l.gotoPage(l.page + 1)
		case key.Matches(msg, keys.First):
			return l, l.gotoPage(0)
		case key.Matches(msg, keys.Last):
			return l, l.gotoPage(l.pages() - 1)
		case key.Matches(msg, keys.Expand):
			l.expand()
			return l, nil
		case key.Matches(msg, keys.Filter):
			return l, filterFieldsCmd(l.provider)
		case key.Matches(msg, keys.New):
			return l, l.newEditorCmd()
		case key.Matches(msg, keys.Refresh):
			return l, l.load()
		}

	default:
		// Cursor blink and other input messages
		if l.editing() {
			cmd, _ := l.editor.Form().Update(msg)
			return l, cmd
		}
	}

	return l, nil
}

// BackLabel returns the collapse action for the quit key while a card is
// expanded and not being edited.
func (l *List) BackLabel() string {
	if l.editing() {
		return ""
	}
	if l.creating {
		return "Cancel"
	}
	if l.expanded {
		return "Collapse"
	}
	return ""
}

func (l *List) Back() {
	l.collapse()
}

// editing returns true while a settings input is focused.
func (l *List) editing() bool {
	return l.expanded && l.editor != nil && l.editor.Form().Editing()
}

func (l *List) Editing() bool {
	return l.editing()
}

// updateDetailKey handles keys while a card is expanded, keys go to the
// settings form while editing and the item action keys open the
// confirm dialog otherwise.
func (l *List) updateDetailKey(msg tea.KeyPressMsg) tea.Cmd {
	if l.editing() {
		cmd, consumed := l.editor.Form().Update(msg)
		if consumed {
			l.refreshForm()
			return cmd
		}

		switch msg.String() {
		case "ctrl+s":
			return l.save()
		case "ctrl+z":
			return l.cancelEdit()
		case "pgup":
			l.detail.PageUp()
		case "pgdown":
			l.detail.PageDown()
		}
		return nil
	}

	switch {
	case key.Matches(msg, keys.Collapse), key.Matches(msg, keys.Expand):
		l.collapse()
		return nil
	case key.Matches(msg, keys.Edit):
		return l.edit()
	case key.Matches(msg, keys.Save):
		return l.save()
	case key.Matches(msg, keys.Cancel):
		return l.cancelEdit()
	case key.Matches(msg, detailKeys.ScrollUp):
		l.detail.ScrollUp(1)
		return nil
	case key.Matches(msg, detailKeys.ScrollDown):
		l.detail.ScrollDown(1)
		return nil
	case key.Matches(msg, detailKeys.PageUp):
		l.detail.PageUp()
		return nil
	case key.Matches(msg, detailKeys.PageDown):
		l.detail.PageDown()
		return nil
	case key.Matches(msg, detailKeys.Prev):
		return l.moveCursor(-1)
	case key.Matches(msg, detailKeys.Next):
		return l.moveCursor(1)
	case key.Matches(msg, keys.First):
		l.detail.GotoTop()
		return nil
	case key.Matches(msg, keys.Last):
		l.detail.GotoBottom()
		return nil
	case key.Matches(msg, keys.Refresh):
		return l.load()
	}

	item := l.selected()
	action, ok := actionByKey(item, msg)
	if ok {
		return l.confirmAction(item, action)
	}

	return nil
}

// updateDetailMouse scrolls the expanded card and handles clicks on the
// button row.
func (l *List) updateDetailMouse(msg tea.MouseMsg) tea.Cmd {
	switch widget.WheelDir(msg) {
	case -1:
		l.detail.ScrollUp(3)
		return nil
	case 1:
		l.detail.ScrollDown(3)
		return nil
	}

	if !widget.IsLeftClick(msg) {
		return nil
	}

	item := l.selected()
	if item == nil {
		return nil
	}

	pos := msg.Mouse()

	// Content starts below the border and title
	top := 2
	if pos.Y >= top && pos.Y < top+l.detail.Height() &&
		l.editor != nil && l.formTop >= 0 {

		contentY := pos.Y - top + l.detail.YOffset()
		if contentY >= l.formTop {
			cmd := l.editor.Form().Click(
				pos.X-detailContentLeft, contentY-l.formTop)
			l.refreshForm()
			return cmd
		}
		return nil
	}

	// Button row is the last row inside the card border
	if pos.Y != l.height-2 {
		return nil
	}

	btn, ok := buttonAt(l.buttons(item), pos.X-detailContentLeft)
	if !ok {
		return nil
	}

	return btn.run()
}

// updateMouse selects the card under the pointer, clicking the selected
// card expands it like opening a row in the web interface.
func (l *List) updateMouse(msg tea.MouseMsg) tea.Cmd {
	if dir := widget.WheelDir(msg); dir != 0 {
		return l.moveCursor(dir)
	}

	pos := msg.Mouse()
	if !widget.IsLeftClick(msg) || pos.Y < 0 {
		return nil
	}

	index := pos.Y / l.cardHeight()
	if index >= len(l.items) {
		return nil
	}

	if index == l.cursor {
		l.expand()
		return nil
	}
	l.cursor = index

	return nil
}

func renderField(field Field, width int) string {
	label := field.Label + ": "
	value := widget.Truncate(
		width-len([]rune(label)),
		"%s",
		widget.Default(field.Value, "-"),
	)
	if field.Color != nil {
		value = lipgloss.NewStyle().Foreground(field.Color).Render(value)
	}

	return widget.LabelStyle.Render(label) + value
}

// renderBody renders the card fields in one or two columns padded to the
// row count so every card has the same height.
func (l *List) renderBody(fields []Field, width int) string {
	rows := make([]string, 0, l.bodyRows())

	if l.split {
		available := min(width, splitMaxWidth)
		rightWidth := available / 2
		leftWidth := available - rightWidth
		leftStyle := cardColStyle.Width(leftWidth)
		rightStyle := cardColStyle.Width(rightWidth)

		for i := 0; i < len(fields); i += 2 {
			left := leftStyle.Render(renderField(fields[i], leftWidth-1))
			right := ""
			if i+1 < len(fields) {
				right = rightStyle.Render(renderField(fields[i+1], rightWidth))
			}
			rows = append(rows, lipgloss.JoinHorizontal(
				lipgloss.Left, left, right))
		}
	} else {
		for _, field := range fields {
			rows = append(rows, renderField(field, width))
		}
	}

	for len(rows) < l.bodyRows() {
		rows = append(rows, "")
	}
	rows = rows[:l.bodyRows()]

	return strings.Join(rows, "\n")
}

func (l *List) renderCard(item Item, selected bool) string {
	style := cardStyle
	if selected {
		style = cardSelectedStyle
	}

	// Style width includes the padding and border
	frameX, _ := style.GetFrameSize()
	contentWidth := max(l.width-frameX, 20)
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
		l.renderBody(item.Fields(), contentWidth),
	)

	return style.Width(cardWidth).Render(content)
}

func (l *List) Render() string {
	var body string

	if len(l.items) == 0 {
		var text string
		if l.loadErr != nil {
			text = "Failed to load " +
				strings.ToLower(l.provider.Title()) + ": " +
				widget.ErrorMessage(l.loadErr)
		} else if !l.loaded {
			text = "Loading " + strings.ToLower(l.provider.Title()) + "..."
		} else {
			text = l.provider.Empty()
		}
		body = widget.EmptyStyle.Width(max(l.width-4, 10)).Render(text)
	} else if l.expanded && l.selected() != nil {
		body = l.renderDetail(l.selected())
	} else {
		cards := make([]string, 0, len(l.items))
		for i, item := range l.items {
			cards = append(cards, l.renderCard(item, i == l.cursor))
		}
		body = lipgloss.JoinVertical(lipgloss.Left, cards...)
	}

	return lipgloss.NewStyle().
		Height(l.height).
		MaxHeight(l.height).
		Render(body)
}
