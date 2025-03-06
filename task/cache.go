package task

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

var cacheClean = &Task{
	Name:    "cache_clean",
	Version: 1,
	Hours: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
	Minutes: []int{0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55},
	Handler: cacheCleanHandler,
}

func cacheCleanHandler(db *database.Database) (err error) {
	cacheDir := node.Self.GetCachePath()

	imageKeys, err := image.GetAllKeys(db)
	if err != nil {
		return
	}

	exists, err := utils.ExistsDir(cacheDir)
	if !exists {
		return
	}

	items, err := ioutil.ReadDir(cacheDir)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "task: Failed to read cache directory"),
		}
		return
	}

	for _, item := range items {
		name := item.Name()
		pth := filepath.Join(cacheDir, name)

		if strings.HasPrefix(name, "image-") {
			keys := strings.Split(name, "-")
			if len(keys) != 3 {
				logrus.WithFields(logrus.Fields{
					"path": pth,
				}).Warning("task: Removing unknown image cache")
				os.Remove(pth)
				continue
			}
			key := fmt.Sprintf("%s-%s", keys[1], keys[2])

			if !imageKeys.Contains(key) {
				if time.Since(item.ModTime()) > 5*time.Minute {
					logrus.WithFields(logrus.Fields{
						"key":  key,
						"path": pth,
					}).Info("task: Removing old image cache")
					os.Remove(pth)
					continue
				}
			}
		}
	}

	return
}

func init() {
	register(cacheClean)
}
