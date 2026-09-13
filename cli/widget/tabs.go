package widget

import (
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	tabBrandStyle  = BarStyle.Bold(true).Padding(0, 1)
	tabStyle       = BarStyle.Padding(0, 2)
	tabActiveStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Background(ColorWhite).
			Bold(true).
			Padding(0, 2)
	tabMoreStyle = BarStyle.Padding(0, 1)
)

const (
	tabMoreLeft  = "‹"
	tabMoreRight = "›"
)

// tabPart is a rendered tab and its column position in the bar, the
// index is -1 for the brand and the more markers.
type tabPart struct {
	text  string
	x     int
	width int
	index int
}

// tabBarParts renders the brand followed by a window of the tabs that
// fits in the width and contains the active tab, more markers show
// when tabs are cut on either side. The first part is the brand.
func tabBarParts(width int, brand string, titles []string,
	active int) []tabPart {

	parts := []tabPart{}
	used := 0

	text := tabBrandStyle.Render(brand)
	w := lipgloss.Width(text)
	parts = append(parts, tabPart{text: text, x: 0, width: w, index: -1})
	used += w

	texts := make([]string, len(titles))
	widths := make([]int, len(titles))
	for i, title := range titles {
		if i == active {
			texts[i] = tabActiveStyle.Render(title)
		} else {
			texts[i] = tabStyle.Render(title)
		}
		widths[i] = lipgloss.Width(texts[i])
	}

	moreLeft := tabMoreStyle.Render(tabMoreLeft)
	moreRight := tabMoreStyle.Render(tabMoreRight)
	moreW := lipgloss.Width(moreLeft)

	// Advance the window start until the active tab fits with the
	// markers on the sides
	avail := width - used
	start := 0
	for start < active {
		need := 0
		for i := start; i <= active; i++ {
			need += widths[i]
		}
		if start > 0 {
			need += moreW
		}
		if active < len(titles)-1 {
			need += moreW
		}
		if need <= avail {
			break
		}
		start++
	}

	if start > 0 {
		parts = append(parts, tabPart{
			text:  moreLeft,
			x:     used,
			width: moreW,
			index: -1,
		})
		used += moreW
	}

	end := start
	for end < len(titles) {
		reserve := 0
		if end < len(titles)-1 {
			reserve = moreW
		}
		if used+widths[end]+reserve > width {
			break
		}
		parts = append(parts, tabPart{
			text:  texts[end],
			x:     used,
			width: widths[end],
			index: end,
		})
		used += widths[end]
		end++
	}

	if end < len(titles) && used+moreW <= width {
		parts = append(parts, tabPart{
			text:  moreRight,
			x:     used,
			width: moreW,
			index: -1,
		})
	}

	return parts
}

// RenderTabBar renders the full width top bar with the brand on the left
// followed by the resource tabs, the active tab is highlighted and the
// tabs scroll to keep it visible.
func RenderTabBar(width int, brand string, titles []string,
	active int) string {

	parts := tabBarParts(width, brand, titles, active)

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

// TabBarClick returns the index of the tab at the column.
func TabBarClick(width int, brand string, titles []string, active int,
	x int) (int, bool) {

	parts := tabBarParts(width, brand, titles, active)
	for _, part := range parts {
		if part.index < 0 {
			continue
		}
		if x >= part.x && x < part.x+part.width {
			return part.index, true
		}
	}
	return 0, false
}
