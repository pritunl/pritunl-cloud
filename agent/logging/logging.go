package logging

import (
	"os"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/agent/constants"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type Redirect struct {
	file *os.File
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

	os.Stdout = r.file
	os.Stderr = r.file

	return
}

func (r *Redirect) Close() {
	r.file.Close()
}
