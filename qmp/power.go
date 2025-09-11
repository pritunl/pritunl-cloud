package qmp

import (
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

func Shutdown(vmId bson.ObjectID) (err error) {
	cmd := &Command{
		Execute: "system_powerdown",
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

	time.Sleep(1000 * time.Millisecond)

	return
}
