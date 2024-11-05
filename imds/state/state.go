package state

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
	"github.com/pritunl/pritunl-cloud/paths"
)

type StateData struct {
	Memory    float64   `json:"memory"`
	HugePages float64   `json:"hugepages"`
	Load1     float64   `json:"load1"`
	Load5     float64   `json:"load5"`
	Load15    float64   `json:"load15"`
	Timestamp time.Time `json:"timestamp"`
}

func Read(instId primitive.ObjectID) (data *StateData, err error) {
	file, err := ioutil.ReadFile(paths.GetImdsStatePath(instId))
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			return
		}

		err = &errortypes.ReadError{
			errors.Wrap(err, "state: File read error"),
		}
		return
	}

	file = bytes.TrimSpace(file)
	if len(file) == 0 {
		return
	}

	data = &StateData{}
	err = json.Unmarshal(file, data)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "state: File unmarshal error"),
		}
		return
	}

	return
}
