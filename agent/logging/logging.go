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

type Redirect struct {
	file       *os.File
	writer     io.Writer
	origStout  *os.File
	origStderr *os.File
	output     chan *types.Entry
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

			timestamp := time.Now()

			fmt.Fprint(r.writer, fmt.Sprintf(
				"[%s] %s\n",
				timestamp.Format("2006-01-02 15:04:05"),
				line,
			))

			if len(r.output) < 9000 {
				r.output <- &types.Entry{
					Timestamp: timestamp,
					Message:   line,
				}
			}
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
