package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/config"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/defaults"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/relations/definitions"
	"github.com/pritunl/pritunl-cloud/router"
	"github.com/pritunl/pritunl-cloud/setup"
	"github.com/pritunl/pritunl-cloud/sync"
	"github.com/pritunl/pritunl-cloud/task"
	"github.com/pritunl/pritunl-cloud/upgrade"
	"github.com/sirupsen/logrus"
)

func Node() (err error) {
	objId, err := primitive.ObjectIDFromHex(config.Config.NodeId)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "cmd: Failed to parse ObjectId"),
		}
		return
	}

	nde := &node.Node{
		Id: objId,
	}
	err = nde.Init()
	if err != nil {
		return
	}

	definitions.Init()

	err = upgrade.Upgrade()
	if err != nil {
		return
	}

	err = setup.Iptables()
	if err != nil {
		return
	}

	err = defaults.Defaults()
	if err != nil {
		return
	}

	sync.Init()

	logrus.WithFields(logrus.Fields{
		"production": constants.Production,
		"types":      nde.Types,
	}).Info("router: Starting node")

	routr := &router.Router{}
	routr.Init()

	err = task.Init()
	if err != nil {
		return
	}

	go func() {
		err = routr.Run()
		if err != nil && !constants.Shutdown {
			panic(err)
		}
	}()

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	logrus.Info("cmd.node: Shutting down")

	constants.Shutdown = true
	go routr.Shutdown()

	if constants.Production && !constants.FastExit {
		time.Sleep(20 * time.Second)
	}

	constants.Interrupt = true

	if !constants.Production || constants.FastExit {
		time.Sleep(300 * time.Millisecond)
	} else {
		time.Sleep(10 * time.Second)
	}

	return
}
