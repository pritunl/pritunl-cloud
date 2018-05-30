package ipsec

import (
	"fmt"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"strings"
)

func networkStopDhClient(vpcId bson.ObjectId) (err error) {
	ifaceInternal := vm.GetLinkIfaceInternal(vpcId, 0)
	pidPath := fmt.Sprintf("/var/run/dhclient-%s.pid", ifaceInternal)

	pid := ""
	pidData, _ := ioutil.ReadFile(pidPath)
	if pidData != nil {
		pid = strings.TrimSpace(string(pidData))
	}

	if pid != "" {
		utils.ExecCombinedOutput("", "kill", pid)
	}

	utils.RemoveAll(pidPath)

	return
}
