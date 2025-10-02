package logging

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/agent/constants"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/types"
)

const maxCapacity = 128 * 1024

type Redirect struct {
	file         *os.File
	writer       io.Writer
	origStout    *os.File
	origStderr   *os.File
	output       chan *types.Entry
	stdoutReader *os.File
	stdoutWriter *os.File
	stderrReader *os.File
	stderrWriter *os.File
}

func (r *Redirect) GetOutput() (entries []*types.Entry) {
	for {
		select {
		case entry := <-r.output:
			entries = append(entries, entry)
		default:
			return
		}
	}
}

func (r *Redirect) handleOutput(reader *os.File, level int32) {
	scanner := bufio.NewScanner(reader)
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()
		timestamp := time.Now()

		fmt.Fprintf(
			r.writer, "[%s] %s\n",
			timestamp.Format("2006-01-02 15:04:05"),
			line,
		)

		select {
		case r.output <- &types.Entry{
			Timestamp: timestamp,
			Level:     level,
			Message:   line,
		}:
		default:
		}
	}
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

	r.output = make(chan *types.Entry, 10000)
	r.writer = io.MultiWriter(r.file, os.Stdout)
	r.origStout = os.Stdout
	r.origStderr = os.Stderr

	r.stdoutReader, r.stdoutWriter, err = os.Pipe()
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "agent: Failed to create stdout pipe"),
		}
		return
	}

	r.stderrReader, r.stderrWriter, err = os.Pipe()
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "agent: Failed to create stderr pipe"),
		}
		return
	}

	os.Stdout = r.stdoutWriter
	os.Stderr = r.stderrWriter

	go r.handleOutput(r.stdoutReader, types.Info)
	go r.handleOutput(r.stderrReader, types.Error)

	return
}

func (r *Redirect) Close() (err error) {
	os.Stdout = r.origStout
	os.Stderr = r.origStderr

	if r.stdoutWriter != nil {
		r.stdoutWriter.Close()
	}
	if r.stdoutReader != nil {
		r.stdoutReader.Close()
	}
	if r.stderrWriter != nil {
		r.stderrWriter.Close()
	}
	if r.stderrReader != nil {
		r.stderrReader.Close()
	}

	err = r.file.Close()
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "agent: Failed to close log pipe"),
		}
		return
	}

	return
}
