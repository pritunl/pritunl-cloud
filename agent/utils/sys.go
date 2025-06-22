package utils

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

func Read(path string) (data string, err error) {
	dataByt, err := ioutil.ReadFile(path)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrapf(err, "utils: Failed to read '%s'", path),
		}
		return
	}

	data = string(dataByt)
	return
}

func DelayExit(code int, delay time.Duration) {
	time.Sleep(delay)
	os.Exit(code)
}
