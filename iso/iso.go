package iso

import (
	"io/ioutil"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

var (
	syncLast  time.Time
	syncLock  sync.Mutex
	syncCache []*Iso
)

type Iso struct {
	Name string `bson:"name" json:"name"`
}

func GetIsos(isoDir string) (isos []*Iso, err error) {
	if time.Since(syncLast) < 30*time.Second {
		isos = syncCache
		return
	}

	syncLock.Lock()
	defer syncLock.Unlock()

	err = utils.ExistsMkdir(isoDir, 0755)
	if err != nil {
		return
	}

	isoFiles, err := ioutil.ReadDir(isoDir)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "backup: Failed to read isos directory"),
		}
		return
	}

	for _, item := range isoFiles {
		filename := item.Name()
		filenameFilt := utils.FilterRelPath(filename)

		if filenameFilt != filename {
			logrus.WithFields(logrus.Fields{
				"name": filename,
			}).Error("iso: Invalid ISO filename")
		}

		iso := &Iso{
			Name: filenameFilt,
		}
		isos = append(isos, iso)
	}

	syncCache = isos
	syncLast = time.Now()

	return
}
