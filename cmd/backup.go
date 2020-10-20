package cmd

import (
	"flag"

	"github.com/pritunl/pritunl-cloud/errortypes"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/backup"
)

func Backup() (err error) {
	dest := flag.Arg(1)

	if dest == "" {
		err = &errortypes.ParseError{
			errors.New("cmd: Missing backup destination path"),
		}
		return
	}

	back := backup.New(dest)

	err = back.Run()
	if err != nil {
		return
	}

	return
}
