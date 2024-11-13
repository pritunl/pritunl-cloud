package state

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/paths"
)

func Read(instId primitive.ObjectID) (data *types.State, err error) {
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

	data = &types.State{}
	err = json.Unmarshal(file, data)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "state: File unmarshal error"),
		}
		return
	}

	return
}
