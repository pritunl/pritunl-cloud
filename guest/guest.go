package guest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	socketsLock = utils.NewMultiTimeoutLock(1 * time.Minute)
)

const (
	syncDelimiter = 0xff
)

type Command struct {
	Execute   string                 `json:"execute"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type Response struct {
	Return  map[string]interface{} `json:"return"`
	Error   map[string]interface{} `json:"error,omitempty"`
	RawData []byte                 `json:"-"`
}

type rawResponse struct {
	Return json.RawMessage        `json:"return"`
	Error  map[string]interface{} `json:"error,omitempty"`
}

type syncResponse struct {
	Return int64 `json:"return"`
}

func writeCommand(conn net.Conn, cmd *Command) (err error) {
	cmdData, err := json.Marshal(cmd)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "guest: Failed to marshal socket data"),
		}
		return
	}

	cmdData = append(cmdData, '\n')

	_, err = conn.Write(cmdData)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "guest: Failed to write socket"),
		}
		return
	}

	return
}

func execute(vmId bson.ObjectID, cmd *Command, timeout time.Duration) (
	ret json.RawMessage, err error) {

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

	err = conn.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "guest: Failed set deadline"),
		}
		return
	}

	reader := bufio.NewReader(conn)

	syncId := time.Now().UnixNano()
	err = writeCommand(conn, &Command{
		Execute: "guest-sync-delimited",
		Arguments: map[string]interface{}{
			"id": syncId,
		},
	})
	if err != nil {
		return
	}

	for {
		line, e := reader.ReadBytes('\n')
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "guest: Failed to read from guest agent"),
			}
			return
		}

		idx := bytes.LastIndexByte(line, syncDelimiter)
		if idx < 0 {
			continue
		}

		resp := &syncResponse{}
		e = json.Unmarshal(bytes.TrimSpace(line[idx+1:]), resp)
		if e == nil && resp.Return == syncId {
			break
		}
	}

	err = writeCommand(conn, cmd)
	if err != nil {
		return
	}

	line, e := reader.ReadBytes('\n')
	if e != nil {
		err = &errortypes.ReadError{
			errors.Wrap(e, "guest: Failed to read from guest agent"),
		}
		return
	}

	resp := &rawResponse{}
	err = json.Unmarshal(bytes.TrimSpace(line), resp)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "guest: Failed to parse guest agent response"),
		}
		return
	}

	if resp.Error != nil {
		err = &errortypes.ReadError{
			errors.Newf("guest: Guest returned error %v", resp.Error),
		}
		return
	}

	ret = resp.Return

	return
}
