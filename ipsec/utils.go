package ipsec

import (
	"fmt"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"io/ioutil"
	"strings"
)

func networkStopDhClient(vpcId primitive.ObjectID) (err error) {
	ifaceExternal := vm.GetLinkIfaceExternal(vpcId, 0)
	pidPath := fmt.Sprintf("/var/run/dhclient-%s.pid", ifaceExternal)

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
