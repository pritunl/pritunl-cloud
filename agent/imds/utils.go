package imds

import (
	"encoding/json"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/agent/constants"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

func GetState() string {
	confData, err := utils.Read(constants.ImdsConfPath)
	if err != nil {
		return ""
	}

	conf := &Imds{}

	err = json.Unmarshal([]byte(confData), conf)
	if err != nil {
		return ""
	}

	return conf.State
}

func SetState(state string) (err error) {
	confData, err := utils.Read(constants.ImdsConfPath)
	if err != nil {
		return
	}

	conf := &Imds{}

	err = json.Unmarshal([]byte(confData), conf)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "agent: Failed to unmarshal imds conf"),
		}
		return
	}

	conf.State = state

	dataByt, err := json.Marshal(conf)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "agent: Failed to unmarshal imds conf"),
		}
		return
	}

	err = utils.Write(constants.ImdsConfPath, string(dataByt), 0600)
	if err != nil {
		return
	}

	return
}
