package widget

import (
	"charm.land/lipgloss/v2"
)

var (
	ColorPrimary = lipgloss.Color("#3B82F6")
	ColorAccent  = lipgloss.Color("#4a8cf7")
	ColorBorder  = lipgloss.Color("#361da3")
	ColorFocus   = lipgloss.Color("#a1cdff")
	ColorGreen   = lipgloss.Color("#10B981")
	ColorRed     = lipgloss.Color("#EF4444")
	ColorYellow  = lipgloss.Color("#fffb00")
	ColorOrange  = lipgloss.Color("#F59E0B")
	ColorCyan    = lipgloss.Color("#22D3EE")
	ColorWhite   = lipgloss.Color("#FFFFFF")
	ColorLabel   = lipgloss.Color("#9CA3AF")
	ColorMuted   = lipgloss.Color("#6B7280")
	ColorDim     = lipgloss.Color("#374151")
	ColorPanel   = lipgloss.Color("#1F2937")
	ColorGray    = lipgloss.Color("#4B5563")
)

var (
	// BarStyle is the full width title and menu bar background.
	BarStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Background(ColorPrimary)

	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)
	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorLabel)
	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)
	EmptyStyle = lipgloss.NewStyle().
			Foreground(ColorLabel).
			Padding(1, 2)

	StatusInfoStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)
	StatusErrorStyle = lipgloss.NewStyle().
				Foreground(ColorRed)

	GreenStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)
	RedStyle = lipgloss.NewStyle().
			Foreground(ColorRed)
	YellowStyle = lipgloss.NewStyle().
			Foreground(ColorYellow)
)
