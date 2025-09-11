package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/requires"
	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	Config            = &ConfigData{}
	StaticRoot        = ""
	StaticTestingRoot = ""
	DefaultMongoUri   = "mongodb://localhost:27017/pritunl-cloud"
)

type ConfigData struct {
	path     string `json:"-"`
	loaded   bool   `json:"-"`
	MongoUri string `json:"mongo_uri"`
	NodeId   string `json:"node_id"`
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

	err = utils.ExistsMkdir(filepath.Dir(constants.ConfPath), 0755)
	if err != nil {
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

	_, err = os.Stat("/cloud/pritunl-cloud.json")
	if err == nil {
		constants.ConfPath = "/cloud/pritunl-cloud.json"
		constants.DefaultRoot = "/cloud"
		constants.DefaultCache = "/cloud/cache"
	}

	_, err = os.Stat(constants.ConfPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			data.loaded = true
			Config = data
		} else {
			err = &errortypes.ReadError{
				errors.Wrap(err, "config: File stat error"),
			}
		}
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

func GetModTime() (mod time.Time, err error) {
	stat, err := os.Stat(constants.ConfPath)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "config: Failed to stat conf file"),
		}
		return
	}

	mod = stat.ModTime()

	return
}

func init() {
	module := requires.New("config")

	module.Handler = func() (err error) {
		for _, pth := range constants.StaticRoot {
			exists, _ := utils.ExistsDir(pth)
			if exists {
				StaticRoot = pth
			}
		}
		if StaticRoot == "" {
			StaticRoot = constants.StaticRoot[len(constants.StaticRoot)-1]
		}

		for _, pth := range constants.StaticTestingRoot {
			exists, _ := utils.ExistsDir(pth)
			if exists {
				StaticTestingRoot = pth
			}
		}
		if StaticTestingRoot == "" {
			StaticTestingRoot = constants.StaticTestingRoot[len(
				constants.StaticTestingRoot)-1]
		}

		err = Load()
		if err != nil {
			return
		}

		save := false

		if Config.NodeId == "" {
			save = true
			Config.NodeId = bson.NewObjectID().Hex()
		}

		if Config.MongoUri == "" {
			save = true

			data, err := utils.ReadExists("/var/lib/mongo/credentials.txt")
			if err != nil {
				err = nil
			} else {
				lines := strings.Split(string(data), "\n")
				for _, line := range lines {
					if strings.HasPrefix(strings.TrimSpace(line),
						"mongodb://pritunl-cloud") {

						Config.MongoUri = strings.TrimSpace(line)
						break
					}
				}
			}

			if Config.MongoUri == "" {
				Config.MongoUri = DefaultMongoUri
			}
		}

		if save {
			err = Save()
			if err != nil {
				return
			}
		}

		return
	}
}
