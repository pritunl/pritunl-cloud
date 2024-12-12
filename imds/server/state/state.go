package state

import (
	"encoding/json"
	"io/ioutil"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/types"
)

var Path = ""
var Global = &Store{
	State:  &types.State{},
	output: make(chan *types.Entry, 10000),
}

type Store struct {
	State  *types.State
	output chan *types.Entry
}

func (s *Store) AppendOutput(entry *types.Entry) {
	if len(s.output) > 9000 {
		return
	}
	s.output <- entry
}

func (s *Store) Save() (err error) {
	data, err := json.MarshalIndent(s.State, "", "\t")
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "state: File marshal error"),
		}
		return
	}

	err = ioutil.WriteFile(Path, data, 0600)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "state: File write error"),
		}
		return
	}

	return
}

func Init() (err error) {
	return
}
