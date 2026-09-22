package widget

import (
	"fmt"
	"io"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"golang.org/x/sys/unix"
)

// selectCommand prints plain text to the terminal with the interface
// released so the text can be selected and copied with the terminal, the
// terminal wraps long lines itself and copies them as one line. The
// text is printed to the main screen so long text scrolls with the
// terminal scrollback and is erased from the screen after returning.
type selectCommand struct {
	title  string
	text   string
	stdin  io.Reader
	stdout io.Writer
}

func (c *selectCommand) SetStdin(r io.Reader) {
	c.stdin = r
}

func (c *selectCommand) SetStdout(w io.Writer) {
	c.stdout = w
}

func (c *selectCommand) SetStderr(w io.Writer) {
}

// quietInput disables echo and line buffering on the terminal so keys
// such as the arrows are not printed, the returned function restores the
// terminal.
func quietInput(stdin io.Reader) func() {
	file, ok := stdin.(interface{ Fd() uintptr })
	if !ok {
		return func() {}
	}
	fd := int(file.Fd())

	state, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return func() {}
	}

	quiet := *state
	quiet.Lflag &^= unix.ECHO | unix.ICANON
	quiet.Cc[unix.VMIN] = 1
	quiet.Cc[unix.VTIME] = 0

	err = unix.IoctlSetTermios(fd, unix.TCSETS, &quiet)
	if err != nil {
		return func() {}
	}

	return func() {
		_ = unix.IoctlSetTermios(fd, unix.TCSETS, state)
	}
}

// waitReturn blocks until enter, q or esc is pressed, other keys and
// escape sequences such as the arrows are ignored.
func waitReturn(stdin io.Reader) {
	buf := make([]byte, 64)
	for {
		n, err := stdin.Read(buf)
		if err != nil {
			return
		}

		if n == 1 {
			switch buf[0] {
			case '\r', '\n', 'q', 'Q', 0x1b:
				return
			}
		}
	}
}

// clearSequence returns the sequence that removes the printed lines
// after returning so a private key is not left on the screen. Output
// that fits on the screen is erased in place, longer output is erased
// from the screen only and the lines that scrolled into the scrollback
// are left for the user to clear.
func clearSequence(stdout io.Writer, output string) string {
	file, ok := stdout.(interface{ Fd() uintptr })
	if !ok {
		return ""
	}

	size, err := unix.IoctlGetWinsize(int(file.Fd()), unix.TIOCGWINSZ)
	if err != nil || size.Col == 0 || size.Row == 0 {
		return ""
	}
	cols := int(size.Col)

	rows := 0
	for _, line := range strings.Split(
		strings.TrimSuffix(output, "\n"), "\n") {

		rows += max((lipgloss.Width(line)+cols-1)/cols, 1)
	}

	if rows >= int(size.Row) {
		return "\x1b[H\x1b[J"
	}

	return fmt.Sprintf("\r\x1b[%dA\x1b[J", rows)
}

func (c *selectCommand) Run() (err error) {
	if c.stdin == nil || c.stdout == nil {
		return
	}

	output := fmt.Sprintf(
		"\n%s\n\n%s\n\n%s\n",
		TitleStyle.Render(c.title),
		strings.TrimRight(c.text, "\n"),
		MutedStyle.Render("Select the text above to copy, scroll "+
			"with the terminal, press enter to return"),
	)

	_, err = fmt.Fprint(c.stdout, output)
	if err != nil {
		return
	}

	restore := quietInput(c.stdin)
	defer restore()

	waitReturn(c.stdin)

	_, err = fmt.Fprint(c.stdout, clearSequence(c.stdout, output))
	return
}

// SelectText prints the text to the terminal without borders or mouse
// capture for copying with the terminal selection, this works in any
// terminal including over SSH.
func SelectText(title, text string) tea.Cmd {
	return tea.Exec(&selectCommand{
		title: title,
		text:  text,
	}, nil)
}
