package widget

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	DialogCancel = -1
	DialogOk     = 1
	DialogInfo   = 2
	// DialogCustom is the first return value free for dialog specific
	// buttons such as a clear button.
	DialogCustom = 10

	dialogInfoFrame = 12
	dialogMaxWidth  = 80
)

var (
	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	dialogTitleStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				PaddingBottom(1)

	dialogHelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			PaddingTop(1)

	dialogInfoLabelStyle = lipgloss.NewStyle().
				Foreground(ColorLabel)
)

// InfoField is a read only label and value shown in the dialog info list.
type InfoField struct {
	Label string
	Value string
}

// dialogRegionBack is the region index of the info view back button.
const dialogRegionBack = -1

// dialogRegion is the clickable area of a dialog option relative to the
// dialog content area.
type dialogRegion struct {
	index int
	x     int
	y     int
	w     int
	h     int
}

func (r dialogRegion) contains(x, y int) bool {
	return x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h
}

type DialogKeyMap struct {
	Left     key.Binding
	Right    key.Binding
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Esc      key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Space    key.Binding
	Quit     key.Binding
	Close    key.Binding
	Back     key.Binding
	Top      key.Binding
	End      key.Binding
}

var dialogKeys = DialogKeyMap{
	Left: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "right"),
	),
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Esc: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "close"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "previous"),
	),
	Space: key.NewBinding(
		key.WithKeys("space"),
		key.WithHelp("space", "toggle"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
	),
	Close: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "close"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "enter", "q", "backspace", "i"),
		key.WithHelp("esc", "back"),
	),
	Top: key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("home", "top"),
	),
	End: key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("end", "end"),
	),
}

// DialogCloseMsg is emitted when a dialog closes with the return value of
// the button that closed it.
type DialogCloseMsg struct {
	Return int
}

type Dialog struct {
	title   string
	message string
	width   int
	height  int
	options []Option

	// Info fields replace the options when the info view is shown, the
	// view is opened by a button returning DialogInfo. Info only dialogs
	// open directly on the info view and close when it is dismissed.
	infoTitle string
	info      []InfoField
	infoOnly  bool
	showInfo  bool
	infoView  viewport.Model
}

func NewDialog(title, message string, opts ...Option) Dialog {
	d := Dialog{
		title:   title,
		message: message,
		width:   60,
		options: opts,
	}

	d.init()

	return d
}

// NewInfoDialog creates a dialog that only shows read only fields.
func NewInfoDialog(title string, fields []InfoField) Dialog {
	d := NewDialog(title, "")
	d.infoOnly = true
	d.SetInfo(title, fields)
	d.openInfo()
	return d
}

// NewMessageDialog creates a dialog showing a message with a close button.
func NewMessageDialog(title, message string) Dialog {
	return NewDialog(
		title,
		message,
		&OptionButton{
			Label:  "Close",
			Return: DialogOk,
		},
	)
}

// NewConfirmDialog creates a dialog with cancel and confirm buttons, the
// confirm button is red when danger is set.
func NewConfirmDialog(title, message, confirm string, danger bool) Dialog {
	return NewDialog(
		title,
		message,
		&OptionButton{
			Label:  "Cancel",
			Return: DialogCancel,
		},
		&OptionButton{
			Label:  confirm,
			Return: DialogOk,
			Danger: danger,
		},
	)
}

// MaxWidth returns the dialog width for the window width.
func MaxWidth(winWidth int) int {
	return min(winWidth-4, dialogMaxWidth)
}

// SetInfo sets the read only fields shown when the info view is opened.
func (d *Dialog) SetInfo(title string, fields []InfoField) {
	d.infoTitle = title
	d.info = fields
	d.infoView = viewport.New(
		viewport.WithWidth(d.contentWidth()),
		viewport.WithHeight(1),
	)
	d.infoView.MouseWheelEnabled = true
	d.renderInfo()
}

// renderInfo rebuilds the info viewport content for the current size.
func (d *Dialog) renderInfo() {
	if d.info == nil {
		return
	}

	contentWidth := max(d.contentWidth()-ScrollbarWidth, 10)
	lineStyle := lipgloss.NewStyle().Width(contentWidth)

	lines := []string{}
	for _, field := range d.info {
		value := field.Value
		if value == "" {
			value = "-"
		}
		lines = append(lines, lineStyle.Render(
			dialogInfoLabelStyle.Render(field.Label+":")+" "+value))
	}
	content := strings.Join(lines, "\n")

	// Fill the available height without leaving empty space below
	// short field lists.
	maxHeight := 20
	if d.height > 0 {
		maxHeight = max(d.height-dialogInfoFrame, 3)
	}
	contentHeight := lipgloss.Height(content)

	d.infoView.SetWidth(contentWidth)
	d.infoView.SetHeight(max(min(maxHeight, contentHeight), 1))
	d.infoView.SetContent(content)
}

func (d *Dialog) openInfo() {
	if d.info == nil {
		return
	}
	d.showInfo = true
	d.infoView.GotoTop()
}

// closeInfo returns to the options or closes info only dialogs.
func (d *Dialog) closeInfo() tea.Cmd {
	if d.infoOnly {
		return closeDialog(DialogCancel)
	}
	d.showInfo = false
	return nil
}

func (d *Dialog) init() {
	contentWidth := d.contentWidth()

	for _, opt := range d.options {
		opt.Init(contentWidth)
	}

	if d.activeIndex() == -1 {
		d.focusIndex(0)
	}
}

func (d *Dialog) contentWidth() int {
	frameX, _ := dialogBoxStyle.GetFrameSize()
	return max(d.width-frameX, 10)
}

func (d *Dialog) SetSize(width, height int) {
	if d.width == width && d.height == height {
		return
	}
	d.width = max(width, 20)
	d.height = height

	contentWidth := d.contentWidth()
	for _, opt := range d.options {
		opt.SetWidth(contentWidth)
	}

	d.renderInfo()
}

func (d *Dialog) activeIndex() int {
	for i, opt := range d.options {
		if opt.Focused() {
			return i
		}
	}
	return -1
}

func (d *Dialog) GetActiveOption() Option {
	i := d.activeIndex()
	if i == -1 {
		return nil
	}
	return d.options[i]
}

func (d *Dialog) focusIndex(index int) tea.Cmd {
	if len(d.options) == 0 {
		return nil
	}

	index = ((index % len(d.options)) + len(d.options)) % len(d.options)

	var cmd tea.Cmd
	for i, opt := range d.options {
		if i == index {
			cmd = opt.Focus()
		} else if opt.Focused() {
			opt.Unfocus()
		}
	}

	return cmd
}

func (d *Dialog) focusNext() tea.Cmd {
	return d.focusIndex(d.activeIndex() + 1)
}

func (d *Dialog) focusPrev() tea.Cmd {
	cur := d.activeIndex()
	if cur == -1 {
		cur = len(d.options)
	}
	return d.focusIndex(cur - 1)
}

func (d *Dialog) focusFooter(dir int) tea.Cmd {
	cur := d.activeIndex()
	if cur == -1 || !d.options[cur].Footer() {
		return nil
	}

	for i := cur + dir; i >= 0 && i < len(d.options); i += dir {
		if d.options[i].Footer() {
			return d.focusIndex(i)
		}
	}

	return nil
}

func (d *Dialog) defaultReturn() (int, bool) {
	for i := len(d.options) - 1; i >= 0; i-- {
		if btn, ok := d.options[i].(*OptionButton); ok {
			return btn.Return, true
		}
	}
	return 0, false
}

func closeDialog(ret int) tea.Cmd {
	return func() tea.Msg {
		return DialogCloseMsg{
			Return: ret,
		}
	}
}

func (d Dialog) renderInfoView() (string, []dialogRegion) {
	contentWidth := d.contentWidth()

	title := d.infoTitle
	if title == "" {
		title = d.title
	}

	fields := []string{
		dialogTitleStyle.Render(title),
		RenderScrollView(d.infoView),
	}
	y := lipgloss.Height(fields[0]) + lipgloss.Height(fields[1])

	label := "Back"
	help := "↑↓: scroll  pgup/pgdn: page  home/end: top/end  esc: back"
	if d.infoOnly {
		label = "Close"
		help = "↑↓: scroll  pgup/pgdn: page  home/end: top/end  esc: close"
	}

	backBtn := &OptionButton{
		Label:   label,
		focused: true,
	}
	backView := backBtn.View()
	regions := []dialogRegion{{
		index: dialogRegionBack,
		x:     0,
		y:     y + optionButtonStyle.GetMarginTop(),
		w:     lipgloss.Width(backView) - optionButtonStyle.GetMarginRight(),
		h:     1,
	}}
	fields = append(fields, backView)

	fields = append(fields, dialogHelpStyle.Width(contentWidth).Render(help))

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		fields...,
	)

	return dialogBoxStyle.Width(d.width).Render(content), regions
}

// render returns the dialog view and the clickable regions of the
// options, region positions are relative to the dialog content area.
func (d Dialog) render() (string, []dialogRegion) {
	if d.showInfo {
		return d.renderInfoView()
	}

	contentWidth := d.contentWidth()
	regions := []dialogRegion{}

	fields := []string{
		dialogTitleStyle.Render(d.title),
	}
	y := lipgloss.Height(fields[0])

	if d.message != "" {
		message := lipgloss.NewStyle().Width(contentWidth).Render(d.message)
		fields = append(fields, message, "")
		y += lipgloss.Height(message) + 1
	}

	footerFields := []string{}
	footerIndex := []int{}
	hasToggle := false
	hasSelect := false

	prevCompact := false
	for i, opt := range d.options {
		if opt.Footer() {
			footerFields = append(footerFields, opt.View())
			footerIndex = append(footerIndex, i)
			continue
		}

		// Toggles and selects are single lines and stack without a gap
		compact := false
		switch opt.(type) {
		case *OptionToggle:
			hasToggle = true
			compact = true
		case *OptionSelect:
			hasSelect = true
			compact = true
		}
		if compact && prevCompact {
			fields = fields[:len(fields)-1]
			y -= 1
		}

		view := opt.View()
		height := lipgloss.Height(view)
		regions = append(regions, dialogRegion{
			index: i,
			x:     0,
			y:     y,
			w:     contentWidth,
			h:     height,
		})

		fields = append(fields, view, "")
		y += height + 1
		prevCompact = compact
	}

	if len(footerFields) > 0 {
		x := 0
		for i, view := range footerFields {
			width := lipgloss.Width(view)
			regions = append(regions, dialogRegion{
				index: footerIndex[i],
				x:     x,
				y:     y + optionButtonStyle.GetMarginTop(),
				w:     width - optionButtonStyle.GetMarginRight(),
				h:     1,
			})
			x += width
		}

		fields = append(fields, lipgloss.JoinHorizontal(
			lipgloss.Top, footerFields...))
	}

	helpText := "tab/↑↓: move  enter: select  esc: close"
	if hasSelect {
		helpText = "tab/↑↓: move  ←/→: change  enter: select  esc: close"
	} else if hasToggle {
		helpText = "tab/↑↓: move  space: toggle  enter: select  esc: close"
	}
	fields = append(fields, dialogHelpStyle.Width(contentWidth).Render(
		helpText))

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		fields...,
	)

	return dialogBoxStyle.Width(d.width).Render(content), regions
}

func (d Dialog) View() string {
	view, _ := d.render()
	return view
}

// Click activates the option under the given position, x and y are
// relative to the top left corner of the dialog box.
func (d Dialog) Click(x, y int) (Dialog, tea.Cmd) {
	x -= dialogBoxStyle.GetBorderLeftSize() + dialogBoxStyle.GetPaddingLeft()
	y -= dialogBoxStyle.GetBorderTopSize() + dialogBoxStyle.GetPaddingTop()

	_, regions := d.render()
	for _, region := range regions {
		if !region.contains(x, y) {
			continue
		}

		if region.index == dialogRegionBack {
			return d, d.closeInfo()
		}

		cmd := d.focusIndex(region.index)
		opt := d.options[region.index]

		switch opt.(type) {
		case *OptionToggle, *OptionButton, *OptionSelect:
			// Clicking runs the same action as pressing enter on the
			// focused option
			var enterCmd tea.Cmd
			d, enterCmd = d.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			return d, tea.Batch(cmd, enterCmd)
		}

		return d, cmd
	}

	return d, nil
}

func (d Dialog) updateInfo(msg tea.Msg) (Dialog, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, dialogKeys.Quit):
			return d, tea.Quit
		case key.Matches(keyMsg, dialogKeys.Back):
			return d, d.closeInfo()
		case key.Matches(keyMsg, dialogKeys.Top):
			d.infoView.GotoTop()
			return d, nil
		case key.Matches(keyMsg, dialogKeys.End):
			d.infoView.GotoBottom()
			return d, nil
		}
	}

	var cmd tea.Cmd
	d.infoView, cmd = d.infoView.Update(msg)
	return d, cmd
}

func (d Dialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if d.showInfo {
		return d.updateInfo(msg)
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		active := d.GetActiveOption()
		if active != nil {
			return d, active.Update(msg)
		}
		return d, nil
	}

	active := d.GetActiveOption()

	switch {
	case key.Matches(keyMsg, dialogKeys.Quit):
		return d, tea.Quit
	case key.Matches(keyMsg, dialogKeys.Esc):
		return d, closeDialog(DialogCancel)
	case key.Matches(keyMsg, dialogKeys.Close):
		if _, isText := active.(*OptionText); !isText {
			return d, closeDialog(DialogCancel)
		}
	case key.Matches(keyMsg, dialogKeys.Tab),
		key.Matches(keyMsg, dialogKeys.Down):
		return d, d.focusNext()
	case key.Matches(keyMsg, dialogKeys.ShiftTab),
		key.Matches(keyMsg, dialogKeys.Up):
		return d, d.focusPrev()
	case key.Matches(keyMsg, dialogKeys.Left):
		if active != nil && active.Footer() {
			return d, d.focusFooter(-1)
		}
	case key.Matches(keyMsg, dialogKeys.Right):
		if active != nil && active.Footer() {
			return d, d.focusFooter(1)
		}
	case key.Matches(keyMsg, dialogKeys.Space):
		if active != nil && active.OnSpace() {
			return d, nil
		}
	case key.Matches(keyMsg, dialogKeys.Enter):
		if active != nil {
			ret, close, handled := active.OnEnter()
			if handled {
				if close {
					if ret == DialogInfo && d.info != nil {
						d.openInfo()
						return d, nil
					}
					return d, closeDialog(ret)
				}
				return d, nil
			}
		}

		ret, ok := d.defaultReturn()
		if ok {
			return d, closeDialog(ret)
		}
		return d, closeDialog(DialogOk)
	}

	if active != nil {
		return d, active.Update(msg)
	}

	return d, nil
}
