package logging

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/agent/constants"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type Redirect struct {
	file       *os.File
	writer     io.Writer
	origStout  *os.File
	origStderr *os.File
}

func (r *Redirect) Open() (err error) {
	r.file, err = os.OpenFile(
		constants.ImdsLogPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0600,
	)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "agent: Failed to create log file"),
		}
		return
	}

	r.writer = io.MultiWriter(r.file, os.Stdout)
	r.origStout = os.Stdout
	r.origStderr = os.Stderr

	reader, writer, err := os.Pipe()
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "agent: Failed to create log pipe"),
		}
		return
	}

	os.Stdout = writer
	os.Stderr = writer

	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()

			fmt.Fprintln(r.writer, line)
		}
	}()

	return
}

func (r *Redirect) Close() (err error) {
	os.Stdout = r.origStout
	os.Stderr = r.origStderr

	err = r.file.Close()
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "agent: Failed to close log pipe"),
		}
		return
	}

	return
}
