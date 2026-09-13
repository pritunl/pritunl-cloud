package widget

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
)

const (
	// ScrollbarWidth is the columns used by the scrollbar and its gap
	// beside a viewport.
	ScrollbarWidth = 2
)

var (
	scrollbarTrackStyle = lipgloss.NewStyle().
				Foreground(ColorDim)
	scrollbarThumbStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary)
)

// RenderScrollbar renders a vertical scrollbar matching the viewport
// height, the thumb size and position reflect the visible portion. The
// column is blank when the content fits in the viewport.
func RenderScrollbar(view viewport.Model) string {
	height := view.Height()
	total := view.TotalLineCount()

	if height <= 0 {
		return ""
	}

	if total <= height {
		return strings.TrimRight(strings.Repeat(" \n", height), "\n")
	}

	thumb := max(height*height/total, 1)
	track := height - thumb
	pos := 0
	if track > 0 {
		pos = view.YOffset() * track / (total - height)
		if view.AtBottom() {
			pos = track
		}
		pos = min(pos, track)
	}

	lines := make([]string, 0, height)
	for i := 0; i < height; i++ {
		if i >= pos && i < pos+thumb {
			lines = append(lines, scrollbarThumbStyle.Render("┃"))
		} else {
			lines = append(lines, scrollbarTrackStyle.Render("│"))
		}
	}

	return strings.Join(lines, "\n")
}

// RenderScrollView renders the viewport with a scrollbar on the right, the
// viewport width must already leave ScrollbarWidth columns for the bar.
func RenderScrollView(view viewport.Model) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		view.View(),
		lipgloss.NewStyle().MarginLeft(1).Render(RenderScrollbar(view)),
	)
}
