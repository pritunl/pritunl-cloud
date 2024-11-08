package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/server/utils"
	"github.com/pritunl/pritunl-cloud/imds/types"
)

var Path = ""
var Config = &ConfigData{}

type ConfigData struct {
	loaded       bool                 `json:"-"`
	ClientIps    []string             `json:"client_ips"`
	Instance     *types.Instance      `json:"instance"`
	Vpc          *types.Vpc           `json:"vpc"`
	Subnet       *types.Subnet        `json:"subnet"`
	Secrets      []*types.Secret      `json:"secrets"`
	Certificates []*types.Certificate `json:"certificates"`
	Services     []*types.Service     `json:"services"`
	Hash         uint32               `json:"hash"`
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
