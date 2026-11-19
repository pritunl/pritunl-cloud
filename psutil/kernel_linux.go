package psutil

import (
	"os"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

func kernelRelease() (release string, err error) {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to read kernel release"),
		}
		return
	}

	release = strings.TrimSpace(string(data))

	return
}
