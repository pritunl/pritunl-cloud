package image

import (
	"github.com/dropbox/godropbox/errors"
)

type LostImageError struct {
	errors.DropboxError
}
