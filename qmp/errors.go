package qmp

import "github.com/dropbox/godropbox/errors"

type DiskNotFound struct {
	errors.DropboxError
}
