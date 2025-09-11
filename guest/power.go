package guest

import (
	"encoding/json"
	"net"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Shutdown(vmId bson.ObjectID) (err error) {
	lockId := socketsLock.Lock(vmId.Hex())
	defer socketsLock.Unlock(vmId.Hex(), lockId)

	sockPath := paths.GetGuestPath(vmId)

	exists, err := utils.Exists(sockPath)
	if err != nil {
		return
	}

	if !exists {
		err = &errortypes.ReadError{
			errors.New("guest: Guest agent socket missing"),
		}
		return
	}

	conn, err := net.DialTimeout(
		"unix",
		sockPath,
		3*time.Second,
	)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "guest: Failed to open socket"),
		}
		return
	}
	defer conn.Close()

	err = conn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "guest: Failed set deadline"),
		}
		return
	}

	cmd := Command{
		Execute: "guest-shutdown",
		Arguments: map[string]interface{}{
			"mode": "powerdown",
		},
	}

	cmdData, err := json.Marshal(cmd)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "guest: Failed to marshal socket data"),
		}
		return
	}

	cmdData = append(cmdData, '\n')

	_, err = conn.Write(cmdData)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "guest: Failed to write socket"),
		}
		return
	}

	buffer := make([]byte, 8192)
	n, err := conn.Read(buffer)
	if err != nil {
		err = nil
		return
	}

	var response Response
	response.RawData = buffer[:n]
	err = json.Unmarshal(response.RawData, &response)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "guest: Failed to parse socket data"),
		}
		return
	}

	if response.Error != nil {
		err = &errortypes.ReadError{
			errors.Newf("guest: Guest returned error %v", response.Error),
		}
		return
	}

	return
}
