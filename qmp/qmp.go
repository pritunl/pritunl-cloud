package qmp

import (
	"bytes"
	"encoding/json"
	"net"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/utils"
)

type cmdBase struct {
	Execute   string      `json:"execute"`
	Arguments interface{} `json:"arguments,omitempty"`
}

type jobStatus struct {
	Id     string `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

type jobStatusReturn struct {
	Return []*jobStatus `json:"return"`
	Error  *cmdError    `json:"error"`
}

type cmdError struct {
	Class string `json:"class"`
	Desc  string `json:"desc"`
}

type cmdReturn struct {
	Return interface{} `json:"return"`
	Error  *cmdError   `json:"error"`
}

var (
	socketsLock = utils.NewMultiTimeoutLock(1 * time.Minute)
)

func runCommand(vmId primitive.ObjectID, cmd *cmdBase,
	cmdReturn interface{}) (err error) {

	// TODO Backward compatibility
	sockPath := paths.GetQmpSockPath(vmId)
	sockPathOld := paths.GetQmpSockPathOld(vmId)

	exists, err := utils.Exists(sockPath)
	if err != nil {
		return
	}

	if !exists {
		sockPath = sockPathOld
	}

	lockId := socketsLock.Lock(vmId.Hex())
	defer func() {
		time.Sleep(100 * time.Millisecond)
		socketsLock.Unlock(vmId.Hex(), lockId)
	}()

	conn, err := net.DialTimeout(
		"unix",
		sockPath,
		10*time.Second,
	)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qmp: Failed to open socket"),
		}
		return
	}
	defer conn.Close()

	err = conn.SetDeadline(time.Now().Add(6 * time.Second))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qmp: Failed set deadline"),
		}
		return
	}

	initCmd := &cmdBase{
		Execute: "qmp_capabilities",
	}

	cmdData, err := json.Marshal(initCmd)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "qmp: Failed to marshal command"),
		}
		return
	}

	_, err = conn.Write([]byte(string(cmdData) + "\n"))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qmp: Failed to write socket"),
		}
		return
	}

	time.Sleep(50 * time.Millisecond)

	cmdData, err = json.Marshal(cmd)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "qmp: Failed to marshal command"),
		}
		return
	}

	_, err = conn.Write([]byte(string(cmdData) + "\n"))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qmp: Failed to write socket"),
		}
		return
	}

	buffer := make([]byte, 100000)
	for {
		buf := make([]byte, 10000)
		n, e := conn.Read(buf)
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "qmp: Failed to read socket"),
			}
			return
		}
		buffer = append(buffer, buf[:n]...)

		if bytes.Count(bytes.TrimSpace(buffer), []byte(`"return"`)) > 1 ||
			bytes.Contains(bytes.TrimSpace(buffer), []byte(`"error"`)) {

			break
		}
	}

	initReturn := false
	returnDataStr := ""
	lines := strings.Split(string(buffer), "\n")
	for _, line := range lines {
		if strings.Contains(line, `"return"`) {
			if initReturn {
				returnDataStr = line
				break
			} else {
				initReturn = true
				continue
			}
		} else if strings.Contains(line, `"return"`) ||
			strings.Contains(line, `"error"`) {

			returnDataStr = line
			break
		}
	}

	err = json.Unmarshal([]byte(returnDataStr), cmdReturn)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrapf(
				err,
				"qmp: Failed to unmarshal return %s",
				returnDataStr,
			),
		}
		return
	}

	return
}
