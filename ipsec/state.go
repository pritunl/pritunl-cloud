package ipsec

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vpc"
)

var (
	syncState      []*vpc.Vpc
	syncStateLock  = sync.Mutex{}
	activeVpcs     set.Set
	activeVpcsLock = sync.Mutex{}
)

func ApplyState(newState []*vpc.Vpc) {
	syncStateLock.Lock()
	syncState = newState
	syncStateLock.Unlock()

	return
}

func InitState() (err error) {
	var items []os.FileInfo
	namespacePth := paths.GetNamespacesPath()
	exists, err := utils.Exists(namespacePth)
	if exists {
		items, err = ioutil.ReadDir(namespacePth)
		if err != nil {
			err = &errortypes.ReadError{
				errors.Wrap(err, "ipsec: Failed to read namespace directory"),
			}
			return
		}
	} else {
		items = []os.FileInfo{}
	}

	newActiveVpcs := set.NewSet()
	for _, item := range items {
		namespace := item.Name()

		if !item.IsDir() || len(namespace) != 14 ||
			!strings.HasPrefix(namespace, "x") {

			continue
		}

		vcPth := filepath.Join("/etc/netns", namespace, "vpc.id")
		vcExists, e := utils.Exists(vcPth)
		if e != nil {
			err = e
			return
		}

		if vcExists {
			vcIdByt, e := ioutil.ReadFile(vcPth)
			if err != nil {
				err = &errortypes.ReadError{
					errors.Wrap(e, "ipsec: Failed to read vpc id"),
				}
				return
			}

			vcId, e := primitive.ObjectIDFromHex(
				strings.TrimSpace(string(vcIdByt)))
			if e != nil {
				err = &errortypes.ParseError{
					errors.Wrap(e, "ipsec: Failed to parse vpc id"),
				}
				return
			}

			newActiveVpcs.Add(vcId)
		}
	}
	activeVpcs = newActiveVpcs

	return
}

func SyncState() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	syncStateLock.Lock()
	curState := syncState
	syncStateLock.Unlock()

	if curState == nil {
		return
	}

	activeVpcsLock.Lock()
	remVpcs := activeVpcs.Copy()
	for _, vc := range curState {
		activeVpcs.Add(vc.Id)
		remVpcs.Remove(vc.Id)
	}
	activeVpcsLock.Unlock()

	for _, vc := range curState {
		held, err := vc.PingLink(db)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("ipsec: Failed to update link timestamp")
			continue
		}

		if !held {
			continue
		}

		go deployVpc(vc)
	}

	for vcIdInf := range remVpcs.Iter() {
		vcId := vcIdInf.(primitive.ObjectID)
		go removeVpc(vcId)
	}

	return
}
