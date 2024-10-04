package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/server/utils"
	"github.com/pritunl/pritunl-cloud/instance"
)

var Path = ""
var Config = &ConfigData{}

type ConfigData struct {
	loaded       bool                       `json:"-"`
	Instance     *instance.Instance         `json:"instance"`
	Certificates []*certificate.Certificate `json:"certificates"`
}

func Load() (err error) {
	data := &ConfigData{}

	exists, err := utils.Exists(Path)
	if err != nil {
		return
	}

	if !exists {
		data.loaded = true
		Config = data
		return
	}

	file, err := ioutil.ReadFile(Path)
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

func Init() (err error) {
	err = Load()
	if err != nil {
		return
	}

	go runSyncConfig()

	return
}
