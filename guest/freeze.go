package guest

import (
	"encoding/json"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

const (
	pingTimeout   = 3 * time.Second
	freezeTimeout = 60 * time.Second
)

func FsFreezeStatus(vmId bson.ObjectID) (frozen bool, err error) {
	ret, err := execute(vmId, &Command{
		Execute: "guest-fsfreeze-status",
	}, pingTimeout)
	if err != nil {
		return
	}

	status := ""
	err = json.Unmarshal(ret, &status)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "guest: Failed to parse freeze status"),
		}
		return
	}

	frozen = status == "frozen"

	return
}

func FsFreeze(vmId bson.ObjectID) (count int, err error) {
	ret, err := execute(vmId, &Command{
		Execute: "guest-fsfreeze-freeze",
	}, freezeTimeout)
	if err != nil {
		return
	}

	err = json.Unmarshal(ret, &count)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "guest: Failed to parse freeze count"),
		}
		return
	}

	return
}

func FsThaw(vmId bson.ObjectID) (count int, err error) {
	ret, err := execute(vmId, &Command{
		Execute: "guest-fsfreeze-thaw",
	}, freezeTimeout)
	if err != nil {
		return
	}

	err = json.Unmarshal(ret, &count)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "guest: Failed to parse thaw count"),
		}
		return
	}

	return
}
