package ipsec

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/link"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
)

var (
	vpcLock = utils.NewMultiTimeoutLock(1 * time.Minute)
)

func deployVpc(vc *vpc.Vpc) {
	accquired, lockId := vpcLock.LockOpen(vc.Id.Hex())
	if !accquired {
		return
	}
	defer vpcLock.Unlock(vc.Id.Hex(), lockId)

	db := database.GetDatabase()
	defer db.Close()

	vcNet, err := vc.GetNetwork()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"vpc_id": vc.Id.Hex(),
			"error":  err,
		}).Error("ipsec: Failed to get ipsec link network")
		return
	}

	netAddr, err := vc.GetIp(db, vpc.Gateway, vc.Id)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"vpc_id": vc.Id.Hex(),
			"error":  err,
		}).Error("ipsec: Failed to get ipsec link local IPv4 address")
		return
	}

	netAddr6 := vc.GetIp6(netAddr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"vpc_id": vc.Id.Hex(),
			"error":  err,
		}).Error("ipsec: Failed to get ipsec link local IPv6 address")
		return
	}

	netCidr, _ := vcNet.Mask.Size()

	err = networkConf(db, vc, netAddr.String(), netAddr6.String(), netCidr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"vpc_id":        vc.Id.Hex(),
			"local_address": netAddr.String(),
			"error":         err,
		}).Error("ipsec: Failed to configure ipsec link network")
		_ = networkConfClear(vc.Id)
		return
	}

	pubAddr, pubAddr6, err := getAddr(vc)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"vpc_id":        vc.Id.Hex(),
			"local_address": netAddr.String(),
			"error":         err,
		}).Error("ipsec: Failed get ipsec public address")
		return
	}

	if pubAddr == "" && pubAddr6 == "" {
		logrus.WithFields(logrus.Fields{
			"vpc_id":          vc.Id.Hex(),
			"local_address":   netAddr.String(),
			"public_address":  pubAddr,
			"public_address6": pubAddr6,
		}).Error("ipsec: Failed to get IP address for ipsec link")
		return
	}

	states := link.GetStates(vc.Id, vc.LinkUris,
		netAddr.String(), pubAddr, pubAddr6)
	hsh := md5.New()

	names := set.NewSet()
	for _, stat := range states {
		for i := range stat.Links {
			names.Add(fmt.Sprintf("%s-%d", stat.Id, i))
		}
		io.WriteString(hsh, stat.Hash)
	}

	newHash := hex.EncodeToString(hsh.Sum(nil))

	link.HashesLock.Lock()
	curHash := link.Hashes[vc.Id]
	link.HashesLock.Unlock()

	if newHash != curHash {
		logrus.WithFields(logrus.Fields{
			"vpc_id": vc.Id.Hex(),
		}).Info("ipsec: Deploying ipsec state")

		err = deployIpsec(vc.Id, states)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"vpc_id": vc.Id.Hex(),
				"error":  err,
			}).Error("ipsec: Failed to deploy state")
			return
		}

		link.HashesLock.Lock()
		link.Hashes[vc.Id] = newHash
		link.HashesLock.Unlock()
	}

	resetLinks, err := link.Update(vc.Id, names)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"vpc_id":          vc.Id.Hex(),
			"local_address":   netAddr.String(),
			"public_address":  pubAddr,
			"public_address6": pubAddr6,
		}).Info("ipsec: Failed to get status")
	}

	if resetLinks != nil && len(resetLinks) != 0 {
		logrus.WithFields(logrus.Fields{
			"vpc_id": vc.Id.Hex(),
		}).Warning("ipsec: Disconnected timeout restarting")

		err = deployIpsec(vc.Id, states)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"vpc_id": vc.Id.Hex(),
				"error":  err,
			}).Error("ipsec: Failed to deploy state")
			return
		}
	}
}

func removeVpc(vcId primitive.ObjectID) {
	accquired, lockId := vpcLock.LockOpen(vcId.Hex())
	if !accquired {
		return
	}
	defer vpcLock.Unlock(vcId.Hex(), lockId)

	link.HashesLock.Lock()
	delete(link.Hashes, vcId)
	link.HashesLock.Unlock()

	err := destroyIpsec(vcId)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"vpc_id": vcId.Hex(),
			"error":  err,
		}).Error("ipsec: Failed to stop ipsec")
		return
	}

	err = networkConfClear(vcId)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"vpc_id": vcId.Hex(),
			"error":  err,
		}).Error("ipsec: Failed to clear network state")
		return
	}

	namespace := vm.GetLinkNamespace(vcId, 0)
	namespacePth := fmt.Sprintf("/etc/netns/%s", namespace)
	os.RemoveAll(namespacePth)

	activeVpcsLock.Lock()
	activeVpcs.Remove(vcId)
	activeVpcsLock.Unlock()
}
