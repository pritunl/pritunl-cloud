package psutil

import (
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"golang.org/x/sys/unix"
)

func kernelRelease() (release string, err error) {
	release, err = unix.Sysctl("kern.osrelease")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to read kernel release"),
		}
		return
	}

	release = strings.TrimSpace(release)

	return
}
