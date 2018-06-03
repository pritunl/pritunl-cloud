package ipsec

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/link"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	syncLock = utils.NewMultiTimeoutLock(1 * time.Minute)
)

func syncStates(vc *vpc.Vpc) {
	if syncLock.Locked(vc.Id.Hex()) {
		return
	}

	lockId := syncLock.Lock(vc.Id.Hex())
	defer syncLock.Unlock(vc.Id.Hex(), lockId)

	db := database.GetDatabase()
	defer db.Close()

	vcNet, err := vc.GetNetwork()
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

	netCidr, _ := vcNet.Mask.Size()

	err = networkConf(vc, netAddr.String(), netAddr6.String(), netCidr)
	if err != nil {
		return
	}

	pubAddr := ""
	pubAddr6 := ""
	for i := 0; i < 3; i++ {
		pubAddr, pubAddr6, err = syncAddr(vc)
		if err != nil {
			return
		}

		if pubAddr6 != "" {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	if pubAddr == "" && pubAddr6 == "" {
		logrus.WithFields(logrus.Fields{
			"vpc_id":          vc.Id.Hex(),
			"local_address":   netAddr.String(),
			"public_address":  pubAddr,
			"public_address6": pubAddr6,
		}).Error("ipsec: Failed to get IPv6 address for ipsec link")
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
		Deploy(vc.Id, states)
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
		Redeploy(vc.Id)
	}
}

func SyncStates(vpcs []*vpc.Vpc) {
	if settings.Local.BridgeName == "" {
		return
	}

	db := database.GetDatabase()
	defer db.Close()

	curLinks := set.NewSet()
	curNamespaces := set.NewSet()
	curVirtIfaces := set.NewSet()
	curInternalIfaces := set.NewSet()

	for _, vc := range vpcs {
		if vc.LinkUris == nil || len(vc.LinkUris) == 0 {
			continue
		}

		if vc.LinkNode != node.Self.Id &&
			time.Since(vc.LinkTimestamp) < time.Duration(
				settings.Ipsec.LinkTimeout)*time.Second {

			continue
		}

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

		curLinks.Add(vc.Id)
		curNamespaces.Add(vm.GetLinkNamespace(vc.Id, 0))
		curVirtIfaces.Add(vm.GetLinkIfaceVirt(vc.Id, 0))
		curInternalIfaces.Add(vm.GetLinkIfaceInternal(vc.Id, 0))

		go syncStates(vc)
	}

	output, err := utils.ExecOutputLogged(
		nil, "ip", "-o", "link", "show",
	)
	if err != nil {
		return
	}

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || len(fields[1]) < 2 {
			continue
		}
		iface := strings.Split(fields[1][:len(fields[1])-1], "@")[0]

		if len(iface) != 14 || !strings.HasPrefix(iface, "y") {
			continue
		}

		if !curVirtIfaces.Contains(iface) {
			utils.ExecCombinedOutputLogged(
				nil,
				"ip", "link", "del", iface,
			)
		}
	}

	output, err = utils.ExecOutputLogged(
		nil,
		"ip", "netns", "list",
	)
	if err != nil {
		return
	}

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		namespace := fields[0]
		if len(namespace) != 14 || !strings.HasPrefix(namespace, "x") {
			continue
		}

		if !curNamespaces.Contains(namespace) {
			_, err = utils.ExecCombinedOutputLogged(
				nil,
				"ip", "netns", "del", namespace,
			)
			if err != nil {
				return
			}
		}
	}

	items, err := ioutil.ReadDir("/etc/netns")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "deploy: Failed to read run directory"),
		}
		return
	}

	for _, item := range items {
		namespace := item.Name()

		if !item.IsDir() || len(namespace) != 14 ||
			!strings.HasPrefix(namespace, "x") {

			continue
		}

		if !curNamespaces.Contains(namespace) {
			os.RemoveAll(filepath.Join("/etc/netns", namespace))
		}
	}

	items, err = ioutil.ReadDir("/var/run")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "deploy: Failed to read run directory"),
		}
		return
	}

	for _, item := range items {
		name := item.Name()

		if item.IsDir() || len(name) != 27 ||
			!strings.HasPrefix(name, "dhclient-z") {

			continue
		}

		iface := name[9:23]

		if !curInternalIfaces.Contains(iface) {
			pth := filepath.Join("/var/run", item.Name())

			pidByt, e := ioutil.ReadFile(pth)
			if e != nil {
				err = &errortypes.ReadError{
					errors.Wrap(e, "ipsec: Failed to read dhclient pid"),
				}
				return
			}

			pid, e := strconv.Atoi(strings.TrimSpace(string(pidByt)))
			if e != nil {
				err = &errortypes.ParseError{
					errors.Wrap(e, "ipsec: Failed to parse dhclient pid"),
				}
				return
			}

			exists, _ := utils.Exists(fmt.Sprintf("/proc/%d/status", pid))
			if exists {
				utils.ExecCombinedOutput("", "kill", "-9", strconv.Itoa(pid))
			} else {
				os.Remove(pth)
			}
		}
	}

	return
}
