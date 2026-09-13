package logs

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/log"
)

const (
	timeFormat = "[2006-01-02 15:04:05]"
)

var (
	timeStyle = lipgloss.NewStyle().
			Bold(true)
	levelStyle = lipgloss.NewStyle().
			Foreground(widget.ColorWhite).
			Bold(true)
	arrowStyle = lipgloss.NewStyle().
			Foreground(widget.ColorPrimary).
			Bold(true)
	diamondStyle = lipgloss.NewStyle().
			Foreground(widget.ColorWhite).
			Bold(true)
	keyStyle = lipgloss.NewStyle().
			Foreground(widget.ColorCyan).
			Bold(true)
	valueStyle = lipgloss.NewStyle().
			Foreground(widget.ColorGreen).
			Bold(true)
	stackStyle = lipgloss.NewStyle().
			Foreground(widget.ColorRed)
)

// FormatLevel returns the level tag matching the node log output.
func FormatLevel(level string) string {
	switch level {
	case log.Debug:
		return "[DEBG]"
	case log.Info:
		return "[INFO]"
	case log.Warning:
		return "[WARN]"
	case log.Error:
		return "[ERRO]"
	case log.Fatal:
		return "[FATL]"
	case log.Panic:
		return "[PANC]"
	}
	return "[UNKN]"
}

func levelBackground(level string) color.Color {
	switch level {
	case log.Info:
		return widget.ColorCyan
	case log.Warning:
		return widget.ColorOrange
	case log.Error, log.Fatal, log.Panic:
		return widget.ColorRed
	}
	return widget.ColorGray
}

// fieldKeys returns the entry field keys sorted for stable output.
func fieldKeys(entry *log.Entry) []string {
	keys := make([]string, 0, len(entry.Fields))
	for key := range entry.Fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// FormatEntry formats an entry as plain text in the format:
//
//	[Timestamp][Level] ▶ Message ◆ key=value ◆ key=value
//	Stack
func FormatEntry(entry *log.Entry) string {
	msg := fmt.Sprintf("%s%s ▶ %s",
		entry.Timestamp.Local().Format(timeFormat),
		FormatLevel(entry.Level),
		entry.Message,
	)

	for _, key := range fieldKeys(entry) {
		msg += fmt.Sprintf(" ◆ %s=%#v", key, entry.Fields[key])
	}

	stack := strings.TrimRight(entry.Stack, "\n")
	if stack != "" {
		msg += "\n" + stack
	}

	return msg
}

// RenderEntry formats an entry with the colors of the node log output.
func RenderEntry(entry *log.Entry) string {
	msg := timeStyle.Render(entry.Timestamp.Local().Format(timeFormat)) +
		levelStyle.Background(levelBackground(entry.Level)).Render(
			FormatLevel(entry.Level)) +
		" " + arrowStyle.Render("▶") + " " + entry.Message

	for _, key := range fieldKeys(entry) {
		msg += " " + diamondStyle.Render("◆") + " " +
			keyStyle.Render(key) + "=" +
			valueStyle.Render(fmt.Sprintf("%#v", entry.Fields[key]))
	}

	stack := strings.TrimRight(entry.Stack, "\n")
	if stack != "" {
		lines := strings.Split(stack, "\n")
		for i, line := range lines {
			lines[i] = stackStyle.Render(line)
		}
		msg += "\n" + strings.Join(lines, "\n")
	}

	return msg
}
