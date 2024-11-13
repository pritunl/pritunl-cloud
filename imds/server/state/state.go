package state

import (
	"encoding/json"
	"io/ioutil"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/types"
)

var Path = ""
var State = &StateData{}

type StateData struct {
	*types.State
}

func (s *StateData) Save() (err error) {
	data, err := json.MarshalIndent(s, "", "\t")
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
