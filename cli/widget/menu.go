package widget

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	menuItemStyle = BarStyle
	menuKeyStyle  = BarStyle.Bold(true)
)

// MenuItem is an action shown in the menu bar with its key.
type MenuItem struct {
	Title string
	Key   string
}

// KeyMsg returns the key press the menu item represents so clicking the
// item runs the same action as the key. Items that only describe
// navigation keys have no click action.
func (i MenuItem) KeyMsg() (tea.KeyPressMsg, bool) {
	switch i.Key {
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}, true
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}, true
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}, true
	case "←/→":
		return tea.KeyPressMsg{Code: tea.KeyRight}, true
	case "pgup/pgdn":
		return tea.KeyPressMsg{Code: tea.KeyPgDown}, true
	case "home":
		return tea.KeyPressMsg{Code: tea.KeyHome}, true
	case "end":
		return tea.KeyPressMsg{Code: tea.KeyEnd}, true
	case "^s":
		return tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}, true
	case "^z":
		return tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl}, true
	}

	runes := []rune(i.Key)
	if len(runes) == 1 {
		return tea.KeyPressMsg{Code: runes[0], Text: string(runes)}, true
	}

	return tea.KeyPressMsg{}, false
}

// menuPart is a rendered menu item and its column position in the bar.
type menuPart struct {
	item  MenuItem
	text  string
	x     int
	width int
}

// menuBarParts renders the items that fit in the bar width.
func menuBarParts(width int, items []MenuItem) []menuPart {
	parts := []menuPart{}
	used := 0

	for _, item := range items {
		text := menuKeyStyle.Render(" ["+item.Key+"]") +
			menuItemStyle.Render(" "+item.Title+" ")
		w := lipgloss.Width(text)
		if used+w > width && len(parts) > 0 {
			break
		}
		parts = append(parts, menuPart{
			item:  item,
			text:  text,
			x:     used,
			width: w,
		})
		used += w
	}

	return parts
}

// RenderMenuBar renders the full width menu bar of key hints.
func RenderMenuBar(width int, items []MenuItem) string {
	parts := menuBarParts(width, items)

	used := 0
	texts := []string{}
	for _, part := range parts {
		texts = append(texts, part.text)
		used += part.width
	}

	bar := strings.Join(texts, "")
	if used < width {
		bar += BarStyle.Render(strings.Repeat(" ", width-used))
	}

	return bar
}

// MenuBarClick returns the key press for the menu item at the column.
func MenuBarClick(width int, items []MenuItem, x int) (tea.KeyPressMsg, bool) {
	for _, part := range menuBarParts(width, items) {
		if x >= part.x && x < part.x+part.width {
			return part.item.KeyMsg()
		}
	}
	return tea.KeyPressMsg{}, false
}

// IsLeftClick returns true for a left mouse button press.
func IsLeftClick(msg tea.MouseMsg) bool {
	click, ok := msg.(tea.MouseClickMsg)
	return ok && click.Button == tea.MouseLeft
}

// WheelDir returns -1 for a wheel up, 1 for a wheel down and 0 for other
// mouse messages.
func WheelDir(msg tea.MouseMsg) int {
	wheel, ok := msg.(tea.MouseWheelMsg)
	if !ok {
		return 0
	}

	switch wheel.Button {
	case tea.MouseWheelUp:
		return -1
	case tea.MouseWheelDown:
		return 1
	}

	return 0
}

// OffsetMouse returns the mouse message moved by the offset, used to
// deliver events relative to a nested view.
func OffsetMouse(msg tea.MouseMsg, dx, dy int) tea.MouseMsg {
	switch m := msg.(type) {
	case tea.MouseClickMsg:
		m.X += dx
		m.Y += dy
		return m
	case tea.MouseReleaseMsg:
		m.X += dx
		m.Y += dy
		return m
	case tea.MouseWheelMsg:
		m.X += dx
		m.Y += dy
		return m
	case tea.MouseMotionMsg:
		m.X += dx
		m.Y += dy
		return m
	}
	return msg
}
