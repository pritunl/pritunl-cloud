package ipsec

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/link"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
)

var (
	ipsecLock = utils.NewMultiTimeoutLock(2 * time.Minute)
)

func deployIpsec(vpcId primitive.ObjectID, states []*link.State) (err error) {
	db := database.GetDatabase()
	defer db.Close()

	lockId := ipsecLock.Lock(vpcId.Hex())
	defer ipsecLock.Unlock(vpcId.Hex(), lockId)

	namespace := vm.GetLinkNamespace(vpcId, 0)

	vc, err := vpc.Get(db, vpcId)
	if err != nil {
		return
	}

	if vc.Subnets == nil || len(vc.Subnets) == 0 {
		err = &errortypes.ReadError{
			errors.New("ipsec: Cannot get VPC default subnet"),
		}
		return
	}
	subId := vc.Subnets[0].Id

	netAddr, _, err := vc.GetIp(db, subId, vc.Id)
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
		"ip", "netns",
		"exec", namespace,
		"ipsec", "restart",
	)
	if err != nil {
		return
	}

	return
}

func destroyIpsec(vcId primitive.ObjectID) (err error) {
	namespace := vm.GetLinkNamespace(vcId, 0)
	namespacePth := fmt.Sprintf("/etc/netns/%s", namespace)

	_, _ = utils.ExecCombinedOutputLogged(
		[]string{
			"No such file or directory",
		},
		"ip", "netns",
		"exec", namespace,
		"ipsec", "stop",
	)

	time.Sleep(1 * time.Second)

	charonPth := filepath.Join(
		namespacePth, "ipsec.d", "run", "charon.pid")
	charonExists, err := utils.Exists(charonPth)
	if err != nil {
		return
	}

	if charonExists {
		pidByt, e := ioutil.ReadFile(charonPth)
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "ipsec: Failed to read pid"),
			}
			return
		}

		pid, e := strconv.Atoi(strings.TrimSpace(string(pidByt)))
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "ipsec: Failed to parse pid"),
			}
			return
		}

		exists, _ := utils.Exists(fmt.Sprintf("/proc/%d/status", pid))
		if exists {
			utils.ExecCombinedOutput("", "kill", "-9", strconv.Itoa(pid))
		}
	}

	starterPth := filepath.Join(
		namespacePth, "ipsec.d", "run", "charon.pid")
	starterExists, err := utils.Exists(starterPth)
	if err != nil {
		return
	}

	if starterExists {
		pidByt, e := ioutil.ReadFile(starterPth)
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "ipsec: Failed to read pid"),
			}
			return
		}

		pid, e := strconv.Atoi(strings.TrimSpace(string(pidByt)))
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "ipsec: Failed to parse pid"),
			}
			return
		}

		exists, _ := utils.Exists(fmt.Sprintf("/proc/%d/status", pid))
		if exists {
			utils.ExecCombinedOutput("", "kill", "-9", strconv.Itoa(pid))
		}
	}

	return
}
