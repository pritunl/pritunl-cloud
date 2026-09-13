package widget

import (
	"fmt"
	"strings"
)

// Truncate formats the text and cuts it to the width with an ellipsis so a
// single line never wraps.
func Truncate(width int, format string, args ...interface{}) string {
	data := []rune(fmt.Sprintf(format, args...))
	if width <= 0 {
		return ""
	}
	if len(data) <= width {
		return string(data)
	}
	if width < 4 {
		return string(data[:width])
	}
	return string(data[:width-3]) + "..."
}

// ErrorMessage returns the first line of an error, dropping the wrapped
// stack trace from dropbox errors.
func ErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()

	if i := strings.Index(msg, "\n"); i != -1 {
		msg = msg[:i]
	}

	return msg
}

// Default returns the fallback when the value is empty.
func Default(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

// JoinDefault joins the values or returns the fallback when empty.
func JoinDefault(values []string, sep, fallback string) string {
	if len(values) == 0 {
		return fallback
	}
	return strings.Join(values, sep)
}

// Bool formats a boolean as a label pair such as Enabled and Disabled.
func Bool(value bool, on, off string) string {
	if value {
		return on
	}
	return off
}
