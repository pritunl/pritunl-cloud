// Package logs provides the logs tab showing the cluster log entries from
// the database in the format of the node log output.
package logs

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/view"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/log"
)

const (
	// limit is the number of most recent entries loaded.
	limit          = 500
	numberMinWidth = 3
)

var (
	numberStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Background(lipgloss.Color("#1F2937"))
)

// loadMsg carries the loaded entries oldest first.
type loadMsg struct {
	seq     int
	entries []*log.Entry
	count   int64
	err     error
}

// clearMsg reports the result of clearing the logs.
type clearMsg struct {
	err error
}

type Logs struct {
	viewport viewport.Model
	entries  []*log.Entry
	count    int64
	follow   bool
	loading  bool
	loaded   bool
	loadErr  error
	width    int
	height   int
	seq      int
	lastSync time.Time
}

func New() *Logs {
	l := &Logs{
		follow:  true,
		entries: []*log.Entry{},
	}
	l.viewport = viewport.New(viewport.WithWidth(10), viewport.WithHeight(1))
	l.viewport.SetContent("Loading...")
	return l
}

func (l *Logs) Title() string {
	return "Logs"
}

func (l *Logs) SetSize(width, height int) {
	l.width = width
	l.height = height
	l.viewport.SetWidth(max(width-widget.ScrollbarWidth, 10))
	l.viewport.SetHeight(max(height, 1))

	// Rewrap the log output for the new width
	if l.loaded {
		l.setContent()
	}
	if l.follow {
		l.viewport.GotoBottom()
	}
}

func loadCmd(seq int) tea.Cmd {
	return func() tea.Msg {
		db := database.GetDatabase()
		if db == nil {
			return loadMsg{
				seq: seq,
				err: &errortypes.DatabaseError{
					errors.New("logs: Database not connected"),
				},
			}
		}
		defer db.Close()

		entries, count, err := log.GetAll(db, &bson.M{}, 0, limit)
		if err != nil {
			return loadMsg{
				seq: seq,
				err: err,
			}
		}

		// Entries are returned newest first, show oldest at the top
		for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
			entries[i], entries[j] = entries[j], entries[i]
		}

		return loadMsg{
			seq:     seq,
			entries: entries,
			count:   count,
		}
	}
}

func clearCmd() tea.Cmd {
	return func() tea.Msg {
		db := database.GetDatabase()
		if db == nil {
			return clearMsg{
				err: &errortypes.DatabaseError{
					errors.New("logs: Database not connected"),
				},
			}
		}
		defer db.Close()

		return clearMsg{
			err: log.Clear(db),
		}
	}
}

func (l *Logs) load() tea.Cmd {
	l.seq++
	l.loading = true
	l.lastSync = time.Now()
	return loadCmd(l.seq)
}

// setContent numbers the log lines and wraps them to the viewport width
// so the page never scrolls horizontally, wrapped lines continue under
// their number.
func (l *Logs) setContent() {
	if len(l.entries) == 0 {
		l.viewport.SetContent("No log entries")
		return
	}

	lines := []string{}
	for _, entry := range l.entries {
		lines = append(lines, strings.Split(RenderEntry(entry), "\n")...)
	}

	numberWidth := max(len(strconv.Itoa(len(lines))), numberMinWidth)

	// Gutter has the number padded by a space on each side followed by a
	// plain space before the text
	indent := numberStyle.Render(strings.Repeat(" ", numberWidth+2)) + " "
	wrapStyle := lipgloss.NewStyle().Width(
		max(l.viewport.Width()-numberWidth-3, 10))

	output := make([]string, 0, len(lines))
	for i, line := range lines {
		number := numberStyle.Render(
			fmt.Sprintf(" %*d ", numberWidth, i+1)) + " "

		for j, part := range strings.Split(wrapStyle.Render(line), "\n") {
			if j == 0 {
				output = append(output, number+part)
			} else {
				output = append(output, indent+part)
			}
		}
	}

	l.viewport.SetContent(strings.Join(output, "\n"))
}

func (l *Logs) setEntries(entries []*log.Entry, count int64) {
	l.loaded = true
	l.entries = entries
	l.count = count

	l.setContent()
	if l.follow {
		l.viewport.GotoBottom()
	}
}

func (l *Logs) Menu() []widget.MenuItem {
	return []widget.MenuItem{
		{Title: "Scroll", Key: "↑/↓"},
		{Title: "Page", Key: "pgup/pgdn"},
		{Title: "Top", Key: "home"},
		{Title: "End", Key: "end"},
		{Title: "Clear", Key: "c"},
		{Title: "Refresh", Key: "R"},
	}
}

func (l *Logs) Status() string {
	if !l.loaded {
		if l.loadErr != nil {
			return "Load failed"
		}
		return "Loading..."
	}

	status := fmt.Sprintf("%d of %d entries", len(l.entries), l.count)
	if l.follow {
		status += "  Following"
	}
	return status
}

func (l *Logs) Editing() bool {
	return false
}

func (l *Logs) clear() tea.Cmd {
	return view.Dialog(widget.NewConfirmDialog(
		"Clear Logs",
		"Clear all log entries? This cannot be undone.",
		"Clear",
		true,
	), func(ret int) tea.Cmd {
		if ret != widget.DialogOk {
			return nil
		}
		return clearCmd()
	})
}

func (l *Logs) Update(msg tea.Msg) (view.View, tea.Cmd) {
	switch msg := msg.(type) {
	case view.ActivateMsg:
		return l, l.load()

	case view.ResizeMsg:
		return l, nil

	case view.TickMsg:
		if !l.loading && time.Since(l.lastSync) >= view.SyncInterval {
			return l, l.load()
		}
		return l, nil

	case loadMsg:
		if msg.seq != l.seq {
			return l, nil
		}
		l.loading = false
		if msg.err != nil {
			if l.loadErr == nil {
				view.LogError("Load Logs", msg.err)
			}
			l.loadErr = msg.err
			return l, view.Status(
				"Load failed: "+widget.ErrorMessage(msg.err), true)
		}
		l.loadErr = nil
		l.setEntries(msg.entries, msg.count)
		return l, nil

	case clearMsg:
		if msg.err != nil {
			return l, view.Error("Clear Logs", msg.err)
		}
		l.follow = true
		return l, tea.Batch(view.Status("Logs cleared", false), l.load())

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keys.Clear):
			return l, l.clear()
		case key.Matches(msg, keys.Refresh):
			return l, l.load()
		case key.Matches(msg, keys.Top):
			l.follow = false
			l.viewport.GotoTop()
			return l, nil
		case key.Matches(msg, keys.End):
			l.follow = true
			l.viewport.GotoBottom()
			return l, nil
		}
	}

	var cmd tea.Cmd
	l.viewport, cmd = l.viewport.Update(msg)
	l.follow = l.viewport.AtBottom()

	return l, cmd
}

func (l *Logs) Render() string {
	return lipgloss.NewStyle().
		Height(l.height).
		MaxHeight(l.height).
		Render(widget.RenderScrollView(l.viewport))
}
