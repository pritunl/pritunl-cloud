package state

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
)

var Path = ""
var State = &StateData{}

type StateData struct {
	Memory    float64   `json:"memory"`
	HugePages float64   `json:"hugepages"`
	Load1     float64   `json:"load1"`
	Load5     float64   `json:"load5"`
	Load15    float64   `json:"load15"`
	Timestamp time.Time `json:"timestamp"`
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
