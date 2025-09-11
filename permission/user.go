package permission

import (
	"fmt"
	"os/user"
	"path"
	"strconv"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
)

func GetUserName(vmId bson.ObjectID) string {
	return fmt.Sprintf("pritunl-%s", vmId.Hex())
}

func UserAdd(virt *vm.VirtualMachine) (err error) {
	name := GetUserName(virt.Id)

	usr, e := user.LookupId(strconv.Itoa(virt.UnixId))
	if usr != nil && e == nil {
		return
	}

	if virt.UnixId == 0 {
		err = &errortypes.ParseError{
			errors.New("permission: Virt missing unix id"),
		}
		return
	}

	_, err = utils.ExecCombinedOutputLogged(nil,
		"useradd",
		"--no-user-group",
		"--no-create-home",
		"--uid", strconv.Itoa(virt.UnixId),
		name,
	)
	if err != nil {
		return
	}

	mailPath := path.Join("/var/mail", name)
	_ = utils.RemoveAll(mailPath)

	return
}

func UserDelete(virt *vm.VirtualMachine) (err error) {
	name := GetUserName(virt.Id)

	_, _ = utils.ExecCombinedOutput("",
		"userdel",
		name,
	)

	_, _ = utils.ExecCombinedOutput("",
		"groupdel",
		name,
	)

	return
}

func UserGroupAdd(virtId bson.ObjectID, group string) (err error) {
	name := GetUserName(virtId)

	_, err = utils.ExecCombinedOutputLogged(
		[]string{
			"does not exist",
		},
		"gpasswd",
		"-a", name,
		group,
	)

	return
}

func UserGroupDelete(virtId bson.ObjectID, group string) (err error) {
	name := GetUserName(virtId)

	_, err = utils.ExecCombinedOutputLogged(
		[]string{
			"not a member",
			"does not exist",
		},
		"gpasswd",
		"-d", name,
		group,
	)

	return
}
