package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/backup"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

func Backup() (err error) {
	dest := flag.Arg(1)

	if dest == "" {
		err = &errortypes.ParseError{
			errors.New("cmd: Missing backup destination path"),
		}
		return
	}

	fmt.Println("Feature comming soon")
	os.Exit(1)

	back := backup.New(dest)

	err = back.Run()
	if err != nil {
		return
	}

	return
}
