package logging

import (
	"github.com/pritunl/pritunl-cloud/imds/types"
)

type Handler interface {
	Open() (err error)
	Close() (err error)
	GetOutput() (entries []*types.Entry)
}
