package iface

import (
	"fmt"
	"math"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/pritunl/pritunl-cloud/cli/authorities"
	"github.com/pritunl/pritunl-cloud/cli/balancers"
	"github.com/pritunl/pritunl-cloud/cli/blocks"
	"github.com/pritunl/pritunl-cloud/cli/certificates"
	"github.com/pritunl/pritunl-cloud/cli/datacenters"
	"github.com/pritunl/pritunl-cloud/cli/disks"
	"github.com/pritunl/pritunl-cloud/cli/domains"
	"github.com/pritunl/pritunl-cloud/cli/firewalls"
	"github.com/pritunl/pritunl-cloud/cli/images"
	"github.com/pritunl/pritunl-cloud/cli/instances"
	"github.com/pritunl/pritunl-cloud/cli/logs"
	"github.com/pritunl/pritunl-cloud/cli/nodes"
	"github.com/pritunl/pritunl-cloud/cli/organizations"
	"github.com/pritunl/pritunl-cloud/cli/plans"
	"github.com/pritunl/pritunl-cloud/cli/policies"
	"github.com/pritunl/pritunl-cloud/cli/pools"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/secrets"
	"github.com/pritunl/pritunl-cloud/cli/shapes"
	"github.com/pritunl/pritunl-cloud/cli/storages"
	"github.com/pritunl/pritunl-cloud/cli/view"
	"github.com/pritunl/pritunl-cloud/cli/vpcs"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/cli/zones"
	"github.com/pritunl/pritunl-cloud/constants"
)

const (
	statusTimeout      = 6 * time.Second
	statusErrorTimeout = 12 * time.Second

	// frameHeight is the tab bar, status line and menu bar around the
	// active tab body.
	frameHeight = 3
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Model is the root of the interface, it owns the tabs, the dialog
// overlay, the status line and the menu bar. Each tab is a view.View,
// resource tabs are resource.List instances with a resource provider.
type Model struct {
	views     []view.View
	active    int
	winWidth  int
	winHeight int
	ready     bool
	activated bool

	statusMsg  string
	statusErr  bool
	statusTime time.Time

	showDialog     bool
	dialog         widget.Dialog
	dialogCallback func(ret int) tea.Cmd
}

// NewModel creates the root model with the resource tabs, new resources
// are added here with a resource provider.
func NewModel() Model {
	return Model{
		views: []view.View{
			resource.NewList(instances.New()),
			resource.NewList(disks.New()),
			resource.NewList(images.New()),
			resource.NewList(firewalls.New()),
			resource.NewList(vpcs.New()),
			resource.NewList(domains.New()),
			resource.NewList(balancers.New()),
			resource.NewList(secrets.New()),
			resource.NewList(authorities.New()),
			resource.NewList(pools.New()),
			resource.NewList(blocks.New()),
			resource.NewList(nodes.New()),
			resource.NewList(shapes.New()),
			resource.NewList(plans.New()),
			resource.NewList(organizations.New()),
			resource.NewList(policies.New()),
			resource.NewList(certificates.New()),
			resource.NewList(storages.New()),
			resource.NewList(datacenters.New()),
			resource.NewList(zones.New()),
			logs.New(),
		},
	}
}

func (m Model) Init() tea.Cmd {
	return tickCmd()
}

func (m Model) brand() string {
	return fmt.Sprintf("Pritunl Cloud v%s", constants.Version)
}

func (m Model) titles() []string {
	titles := make([]string, 0, len(m.views))
	for _, v := range m.views {
		titles = append(titles, v.Title())
	}
	return titles
}

func (m Model) activeView() view.View {
	return m.views[m.active]
}

func (m *Model) setStatus(msg string, isErr bool) {
	m.statusMsg = msg
	m.statusErr = isErr
	m.statusTime = time.Now()
}

func (m *Model) openDialog(d widget.Dialog, callback func(ret int) tea.Cmd) {
	d.SetSize(widget.MaxWidth(m.winWidth), m.winHeight)
	m.dialog = d
	m.dialogCallback = callback
	m.showDialog = true
}

func (m *Model) openError(action string, err error) {
	view.LogError(action, err)
	m.openDialog(widget.NewMessageDialog(
		action+" Failed", widget.ErrorMessage(err)), nil)
}

// update delivers the message to the tab at the index.
func (m *Model) update(index int, msg tea.Msg) tea.Cmd {
	v, cmd := m.views[index].Update(msg)
	m.views[index] = v
	return cmd
}

// updateActive delivers the message to the active tab only.
func (m *Model) updateActive(msg tea.Msg) tea.Cmd {
	return m.update(m.active, msg)
}

// updateAll delivers the message to every tab, data messages are
// delivered to every tab so background results reach the tab that
// requested them.
func (m *Model) updateAll(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	for i := range m.views {
		cmds = append(cmds, m.update(i, msg))
	}
	return tea.Batch(cmds...)
}

// activate switches to the tab at the index and asks it to load.
func (m *Model) activate(index int) tea.Cmd {
	n := len(m.views)
	if n == 0 {
		return nil
	}
	m.active = ((index % n) + n) % n
	m.activated = true
	return m.updateActive(view.ActivateMsg{})
}

func (m Model) updateSize(msg tea.WindowSizeMsg) (Model, tea.Cmd) {
	m.winWidth = msg.Width
	m.winHeight = msg.Height

	bodyHeight := max(m.winHeight-frameHeight, 1)
	for _, v := range m.views {
		v.SetSize(m.winWidth, bodyHeight)
	}

	if m.showDialog {
		m.dialog.SetSize(widget.MaxWidth(m.winWidth), m.winHeight)
	}

	m.ready = true

	if !m.activated {
		return m, m.activate(m.active)
	}
	return m, m.updateActive(view.ResizeMsg{})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.updateSize(msg)

	case tickMsg:
		cmds := []tea.Cmd{tickCmd()}
		if m.activated {
			cmds = append(cmds, m.updateActive(view.TickMsg(msg)))
		}
		return m, tea.Batch(cmds...)

	case view.DialogMsg:
		m.openDialog(msg.Dialog, msg.Callback)
		return m, nil

	case view.StatusMsg:
		m.setStatus(msg.Message, msg.Error)
		return m, nil

	case view.ErrorMsg:
		m.setStatus(msg.Action+" failed", true)
		m.openError(msg.Action, msg.Err)
		return m, nil

	case widget.DialogCloseMsg:
		m.showDialog = false
		callback := m.dialogCallback
		m.dialogCallback = nil
		if callback != nil {
			return m, callback(msg.Return)
		}
		return m, nil

	case tea.MouseMsg:
		return m.updateMouse(msg)

	case tea.KeyPressMsg:
		if m.showDialog {
			var cmd tea.Cmd
			m.dialog, cmd = m.dialog.Update(msg)
			return m, cmd
		}

		// Text inputs receive every key while editing
		if m.activated && m.activeView().Editing() {
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, m.updateActive(msg)
		}

		switch {
		case key.Matches(msg, keys.Quit):
			if backer, ok := m.backer(); ok {
				backer.Back()
				return m, nil
			}
			return m, tea.Quit
		case key.Matches(msg, keys.NextTab):
			return m, m.activate(m.active + 1)
		case key.Matches(msg, keys.PrevTab):
			return m, m.activate(m.active - 1)
		}

		// Number keys jump directly to a tab
		if len(msg.Text) == 1 && msg.Text[0] >= '1' && msg.Text[0] <= '9' {
			index := int(msg.Text[0] - '1')
			if index < len(m.views) {
				return m, m.activate(index)
			}
			return m, nil
		}

		return m, m.updateActive(msg)
	}

	// Data messages from background commands
	return m, m.updateAll(msg)
}

// placeOffset returns the start position of content centered in size,
// matching lipgloss.Place which puts the odd remainder after the content.
func placeOffset(size, content int) int {
	gap := size - content
	if gap <= 0 {
		return 0
	}
	return gap - int(math.Round(float64(gap)*0.5))
}

func (m Model) updateMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	pos := msg.Mouse()

	if m.showDialog {
		var cmd tea.Cmd
		if widget.IsLeftClick(msg) {
			dialogView := m.dialog.View()
			x := pos.X - placeOffset(m.winWidth, lipgloss.Width(dialogView))
			y := pos.Y - placeOffset(m.winHeight, lipgloss.Height(dialogView))
			m.dialog, cmd = m.dialog.Click(x, y)
			return m, cmd
		}

		m.dialog, cmd = m.dialog.Update(msg)
		return m, cmd
	}

	if widget.IsLeftClick(msg) {
		// Tab bar is the first line of the view
		if pos.Y == 0 {
			index, ok := widget.TabBarClick(
				m.winWidth, m.brand(), m.titles(), m.active, pos.X)
			if ok && index != m.active {
				return m, m.activate(index)
			}
			return m, nil
		}

		// Menu bar is the last line of the view
		if pos.Y == m.winHeight-1 {
			keyMsg, ok := widget.MenuBarClick(
				m.winWidth, m.menuItems(), pos.X)
			if ok {
				return m.Update(keyMsg)
			}
			return m, nil
		}

		// Status line has no actions
		if pos.Y == m.winHeight-2 {
			return m, nil
		}
	}

	// Tab body starts below the tab bar
	return m, m.updateActive(widget.OffsetMouse(msg, 0, -1))
}

// backer returns the active tab back action when it has one.
func (m Model) backer() (view.Backer, bool) {
	if !m.activated {
		return nil, false
	}
	backer, ok := m.activeView().(view.Backer)
	if !ok || backer.BackLabel() == "" {
		return nil, false
	}
	return backer, true
}

func (m Model) menuItems() []widget.MenuItem {
	menu := []widget.MenuItem{}
	if m.activated {
		menu = append(menu, m.activeView().Menu()...)
	}

	quit := widget.MenuItem{Title: "Quit", Key: "q"}
	if backer, ok := m.backer(); ok {
		quit.Title = backer.BackLabel()
	}

	return append(menu,
		widget.MenuItem{Title: "Next Tab", Key: "tab"},
		quit,
	)
}

// renderStatus renders the line above the menu bar with the tab status
// on the left and the status message beside it.
func (m Model) renderStatus() string {
	width := m.winWidth - 1
	line := ""

	if m.activated {
		text := m.activeView().Status()
		if text != "" {
			text = widget.Truncate(width, "%s", text)
			line = " " + widget.LabelStyle.Render(text)
			width -= len([]rune(text)) + 2
		}
	}

	timeout := statusTimeout
	if m.statusErr {
		timeout = statusErrorTimeout
	}
	if m.statusMsg != "" && time.Since(m.statusTime) < timeout &&
		width > 4 {

		text := widget.Truncate(width, "%s", m.statusMsg)
		if m.statusErr {
			line += " " + widget.StatusErrorStyle.Render(text)
		} else {
			line += " " + widget.StatusInfoStyle.Render(text)
		}
	}

	return line
}

// View renders the interface in the alternate screen with mouse cell
// motion for the tab, card and button clicks.
func (m Model) View() tea.View {
	view := tea.NewView(m.render())
	view.AltScreen = true
	view.MouseMode = tea.MouseModeCellMotion
	return view
}

func (m Model) render() string {
	if !m.ready {
		return "Initializing..."
	}

	if m.showDialog {
		return lipgloss.Place(
			m.winWidth,
			m.winHeight,
			lipgloss.Center,
			lipgloss.Center,
			m.dialog.View(),
		)
	}

	var body string
	if m.activated {
		body = m.activeView().Render()
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		widget.RenderTabBar(m.winWidth, m.brand(), m.titles(), m.active),
		body,
		m.renderStatus(),
		widget.RenderMenuBar(m.winWidth, m.menuItems()),
	)
}
