// Package iface runs the terminal interface, the interface connects
// directly to the database from the admin perspective and is organized
// as tabs of resource lists.
package iface

import (
	"io"

	tea "charm.land/bubbletea/v2"
	_ "charm.land/huh/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/sirupsen/logrus"
)

// Iface runs the terminal interface. True color is forced by default as
// SSH and sudo sessions rarely set COLORTERM and the detected profile
// downsamples the colors, autoColor uses the detected profile instead.
func Iface(autoColor bool) (err error) {
	// Log output to the terminal would corrupt the interface, entries
	// still reach the log file and database through the logger hook
	logrus.SetOutput(io.Discard)

	opts := []tea.ProgramOption{}
	if !autoColor {
		opts = append(opts, tea.WithColorProfile(colorprofile.TrueColor))
	}

	// Alternate screen and mouse mode are set on the view
	prog := tea.NewProgram(NewModel(), opts...)

	_, err = prog.Run()
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "iface: Program run error"),
		}
		return
	}

	return
}
