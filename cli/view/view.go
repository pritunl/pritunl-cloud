// Package view defines the interface each top level tab of the interface
// implements and the messages a tab sends to the root model.
package view

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/sirupsen/logrus"
)

const (
	// SyncInterval is how often a tab refreshes its data from the
	// database, matching the web interface sync interval.
	SyncInterval = 3 * time.Second
)

// View is a top level tab such as a resource list or the logs.
type View interface {
	// Title is the tab label.
	Title() string

	// Update handles messages, key and mouse input is only delivered to
	// the active tab while data messages are delivered to every tab.
	Update(msg tea.Msg) (View, tea.Cmd)

	// Render draws the tab body, it must fill the height given by SetSize.
	Render() string

	// SetSize sets the size of the tab body.
	SetSize(width, height int)

	// Menu returns the key hints shown in the menu bar.
	Menu() []widget.MenuItem

	// Status returns the text shown on the left of the status line.
	Status() string

	// Editing returns true while the tab has a focused text input, every
	// key except quit is then delivered to the tab.
	Editing() bool
}

// Backer is implemented by tabs with a back action bound to the quit key
// such as collapsing an expanded resource, the interface only quits when
// there is nothing to go back from.
type Backer interface {
	// BackLabel returns the menu label of the back action or empty when
	// there is nothing to go back from.
	BackLabel() string

	// Back performs the back action.
	Back()
}

// TickMsg is sent to the active tab every second.
type TickMsg time.Time

// ActivateMsg is sent to a tab when it becomes the active tab, the tab
// should load or refresh its data.
type ActivateMsg struct{}

// DialogMsg asks the root model to open a dialog, the callback runs when
// the dialog closes with the return value of the button that closed it.
type DialogMsg struct {
	Dialog   widget.Dialog
	Callback func(ret int) tea.Cmd
}

// StatusMsg sets the status line message.
type StatusMsg struct {
	Message string
	Error   bool
}

// ErrorMsg reports a failed action, the error is logged and shown in a
// dialog.
type ErrorMsg struct {
	Action string
	Err    error
}

// Dialog returns a command opening the dialog.
func Dialog(d widget.Dialog, callback func(ret int) tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		return DialogMsg{
			Dialog:   d,
			Callback: callback,
		}
	}
}

// Status returns a command setting the status line message.
func Status(message string, isErr bool) tea.Cmd {
	return func() tea.Msg {
		return StatusMsg{
			Message: message,
			Error:   isErr,
		}
	}
}

// Error returns a command reporting a failed action.
func Error(action string, err error) tea.Cmd {
	return func() tea.Msg {
		return ErrorMsg{
			Action: action,
			Err:    err,
		}
	}
}

// LogError logs a failed action.
func LogError(action string, err error) {
	logrus.WithFields(logrus.Fields{
		"action": action,
		"error":  err,
	}).Error("tui: Action failed")
}

// ResizeMsg is sent to the active tab after SetSize so it can reload
// when the page size changed.
type ResizeMsg struct{}
