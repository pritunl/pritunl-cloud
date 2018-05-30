package ipsec

import (
	"github.com/Sirupsen/logrus"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/link"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"gopkg.in/mgo.v2/bson"
	"path"
	"sync"
	"time"
)

var (
	deployStates = map[bson.ObjectId][]*link.State{}
	curStates    = map[bson.ObjectId][]*link.State{}
	deployLock   = sync.Mutex{}
	ipsecLock    = utils.NewMultiTimeoutLock(2 * time.Minute)
)

func deploy(vpcId bson.ObjectId, states []*link.State) (err error) {
	db := database.GetDatabase()
	defer db.Close()

	lockId := ipsecLock.Lock(vpcId.Hex())
	defer ipsecLock.Unlock(vpcId.Hex(), lockId)

	namespace := vm.GetLinkNamespace(vpcId, 0)

	vc, err := vpc.Get(db, vpcId)
	if err != nil {
		return
	}

	netAddr, err := vc.GetIp(db, vpc.Gateway, vc.Id)
	if err != nil {
		return
	}

	netAddr6 := vc.GetIp6(netAddr)
	if err != nil {
		return
	}

	runPth := path.Join("/", "etc", "netns", namespace, "ipsec.d", "run")
	err = utils.ExistsMkdir(runPth, 0755)
	if err != nil {
		return
	}

	err = writeTemplates(vpcId, states)
	if err != nil {
		return
	}

	err = addRoutes(db, vc, states,
		netAddr.String(), netAddr6.String())
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "netns", "exec", namespace,
		"ipsec", "restart",
	)
	if err != nil {
		return
	}

	return
}

func Deploy(vcId bson.ObjectId, states []*link.State) {
	deployLock.Lock()
	deployStates[vcId] = states
	deployLock.Unlock()
}

func Redeploy(vcId bson.ObjectId) {
	deployLock.Lock()
	if deployStates[vcId] == nil && curStates[vcId] != nil {
		deployStates[vcId] = curStates[vcId]
	}
	deployLock.Unlock()
}

func runDeploy() {
	for {
		deploying := map[bson.ObjectId][]*link.State{}
		deployLock.Lock()
		for vpcId, states := range deployStates {
			if states == nil {
				continue
			}
			deploying[vpcId] = states
		}
		deployStates = map[bson.ObjectId][]*link.State{}
		deployLock.Unlock()

		for vpcId, states := range deploying {
			logrus.WithFields(logrus.Fields{
				"vpc_id": vpcId.Hex(),
			}).Info("state: Deploying IPsec state")

			err := deploy(vpcId, states)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("state: Failed to deploy state")

				time.Sleep(3 * time.Second)

				deployLock.Lock()
				if deployStates[vpcId] == nil {
					deployStates[vpcId] = states
				}
				deployLock.Unlock()
			}

			deployLock.Lock()
			curStates[vpcId] = states
			deployLock.Unlock()
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func init() {
	go func() {
		time.Sleep(6 * time.Second)

		go runDeploy()
	}()
}
