package qmp

import (
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

func Shutdown(vmId primitive.ObjectID) (err error) {
	cmd := &cmdBase{
		Execute: "system_powerdown",
	}

	returnData := &cmdReturn{}
	err = runCommand(vmId, cmd, returnData)
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
