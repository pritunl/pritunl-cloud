package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/imds/server/constants"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/server/utils"
	"github.com/pritunl/pritunl-cloud/instance"
)

var Config = &ConfigData{}

type ConfigData struct {
	loaded       bool `json:"-"`
	Instance     *instance.Instance
	Certificates []*certificate.Certificate
}

func (c *ConfigData) Save() (err error) {
	if !c.loaded {
		err = &errortypes.WriteError{
			errors.New("config: Config file has not been loaded"),
		}
		return
	}

	data, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "config: File marshal error"),
		}
		return
	}

	err = ioutil.WriteFile(constants.ConfPath, data, 0600)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "config: File write error"),
		}
		return
	}

	return
}

func Load() (err error) {
	data := &ConfigData{}

	exists, err := utils.Exists(constants.ConfPath)
	if err != nil {
		return
	}

	if !exists {
		data.loaded = true
		Config = data
		return
	}

	file, err := ioutil.ReadFile(constants.ConfPath)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "config: File read error"),
		}
		return
	}

	err = json.Unmarshal(file, data)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "config: File unmarshal error"),
		}
		return
	}

	data.loaded = true

	Config = data

	return
}

func Save() (err error) {
	err = Config.Save()
	if err != nil {
		return
	}

	return
}

func Init() (err error) {
	err = Load()
	if err != nil {
		return
	}

	exists, err := utils.Exists(constants.ConfPath)
	if err != nil {
		panic(err)
	}

	if !exists {
		err = Save()
		if err != nil {
			return
		}
	}

	go runSyncConfig()

	return
}
