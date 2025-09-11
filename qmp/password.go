package qmp

import (
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

const (
	Spice = "spice"
	Vnc   = "vnc"
)

type setPasswordArgs struct {
	Protocol string `json:"protocol"`
	Password string `json:"password"`
}

func SetPassword(vmId bson.ObjectID, proto, passwd string) (err error) {
	cmd := &Command{
		Execute: "set_password",
		Arguments: &setPasswordArgs{
			Protocol: proto,
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

	time.Sleep(50 * time.Millisecond)

	return
}
