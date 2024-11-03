package state

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
)

var Path = ""
var tempPath = ""
var State = &StateData{}

type StateData struct {
	SyncTimestamp time.Time `json:"sync_timestamp"`
}

func (s *StateData) Save() (err error) {
	data, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "state: File marshal error"),
		}
		return
	}

	err = ioutil.WriteFile(tempPath, data, 0600)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "state: File write error"),
		}
		return
	}

	err = os.Rename(tempPath, Path)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "state: File rename error"),
		}
		return
	}

	return
}

func Init() (err error) {
	tempPath = Path + ".tmp"

	return
}
