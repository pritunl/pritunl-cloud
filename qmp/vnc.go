package qmp

import (
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type vncPasswordArgs struct {
	Password string `json:"password"`
}

func VncPassword(vmId bson.ObjectID, passwd string) (err error) {
	cmd := &Command{
		Execute: "change-vnc-password",
		Arguments: &vncPasswordArgs{
			Password: passwd,
		},
	}

	returnData := &CommandReturn{}
	err = RunCommand(vmId, cmd, returnData)
	if err != nil {
		return
	}

	if returnData.Error != nil {
		err = &errortypes.ApiError{
			errors.Newf("qmp: Return error %s", returnData.Error.Desc),
		}
		return
	}

	time.Sleep(1 * time.Second)

	return
}
