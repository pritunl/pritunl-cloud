package permission

import (
	"fmt"
	"os/user"
	"strconv"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
)

func GetUserName(vmId primitive.ObjectID) string {
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
		"--user-group",
		"--no-create-home",
		"--uid", strconv.Itoa(virt.UnixId),
		name,
	)
	if err != nil {
		return
	}

	return
}
