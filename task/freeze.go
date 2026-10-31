package task

import (
	"sync"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/guest"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

var freezeRecover = &Task{
	Name:       "freeze_recover",
	Local:      true,
	RunOnStart: true,
	Handler:    freezeRecoverHandler,
}

func freezeRecoverThaw(instId bson.ObjectID) {
	frozen, err := guest.FsFreezeStatus(instId)
	if err != nil {
		return
	}

	if !frozen {
		return
	}

	logrus.WithFields(logrus.Fields{
		"instance_id": instId.Hex(),
	}).Warning("task: Thawing guest filesystems left frozen")

	_, err = guest.FsThaw(instId)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"instance_id": instId.Hex(),
			"error":       err,
		}).Error("task: Failed to thaw guest filesystems")
		return
	}
}

func freezeRecoverHandler(db *database.Database) (err error) {
	if !node.Self.IsHypervisor() {
		return
	}

	insts, err := instance.GetAll(db, &bson.M{
		"node": node.Self.Id,
	})
	if err != nil {
		return
	}

	waiter := &sync.WaitGroup{}
	limiter := make(chan struct{}, 8)

	for _, inst := range insts {
		exists, e := utils.Exists(paths.GetGuestPath(inst.Id))
		if e != nil || !exists {
			continue
		}

		waiter.Add(1)
		limiter <- struct{}{}
		go func(instId bson.ObjectID) {
			defer waiter.Done()
			defer func() {
				<-limiter
			}()

			freezeRecoverThaw(instId)
		}(inst.Id)
	}

	waiter.Wait()

	return
}

func init() {
	register(freezeRecover)
}
