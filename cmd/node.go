package cmd

import (
	"github.com/Sirupsen/logrus"
	"github.com/pritunl/pritunl-cloud/config"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/router"
	"github.com/pritunl/pritunl-cloud/setup"
	"github.com/pritunl/pritunl-cloud/sync"
	"github.com/pritunl/pritunl-cloud/task"
	"gopkg.in/mgo.v2/bson"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Node() (err error) {
	nde := &node.Node{
		Id: bson.ObjectIdHex(config.Config.NodeId),
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
