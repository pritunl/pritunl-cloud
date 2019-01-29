package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/config"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/router"
	"github.com/pritunl/pritunl-cloud/setup"
	"github.com/pritunl/pritunl-cloud/sync"
	"github.com/pritunl/pritunl-cloud/task"
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

	err = setup.Iptables()
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

	task.Init()

	go func() {
		err = routr.Run()
		if err != nil {
			panic(err)
		}
	}()

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	constants.Interrupt = true

	logrus.Info("cmd.node: Shutting down")
	go routr.Shutdown()
	if constants.Production {
		time.Sleep(10 * time.Second)
	} else {
		time.Sleep(300 * time.Millisecond)
	}

	return
}
