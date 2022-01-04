package qmp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Command struct {
	Execute   string      `json:"execute"`
	Arguments interface{} `json:"arguments,omitempty"`
}

type JobStatus struct {
	Id     string `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

type JobStatusReturn struct {
	Return []*JobStatus  `json:"return"`
	Error  *CommandError `json:"error"`
}

type CommandError struct {
	Class string `json:"class"`
	Desc  string `json:"desc"`
}

type CommandReturn struct {
	Return interface{}   `json:"return"`
	Error  *CommandError `json:"error"`
}

var (
	socketsLock = utils.NewMultiTimeoutLock(1 * time.Minute)
)

type Connection struct {
	vmId     primitive.ObjectID
	sock     net.Conn
	lockId   primitive.ObjectID
	command  interface{}
	response interface{}
}

func (c *Connection) connect() (err error) {
	// TODO Backward compatibility
	sockPath := paths.GetQmpSockPath(c.vmId)
	sockPathOld := paths.GetQmpSockPathOld(c.vmId)

	exists, err := utils.Exists(sockPath)
	if err != nil {
		return
	}

	if !exists {
		sockPath = sockPathOld
	}

	c.lockId = socketsLock.Lock(c.vmId.Hex())

	c.sock, err = net.DialTimeout(
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

	err = c.sock.SetDeadline(time.Now().Add(6 * time.Second))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qmp: Failed set deadline"),
		}
		return
	}

	var info []byte
	for {
		buffer := make([]byte, 5000000)
		n, e := c.sock.Read(buffer)
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "qmp: Failed to read socket"),
			}
			return
		}
		buffer = buffer[:n]

		lines := bytes.Split(buffer, []byte("\n"))
		for _, line := range lines {
			if !constants.Production {
				fmt.Println(string(line))
			}

			if bytes.Contains(line, []byte(`"QMP"`)) {
				info = line
				break
			}
		}

		if info != nil {
			break
		}
	}

	if info == nil {
		err = &errortypes.ReadError{
			errors.New("qmp: No info message from socket"),
		}
		return
	}

	return
}

func (c *Connection) Close() {
	sock := c.sock
	if sock != nil {
		_ = sock.Close()
	}

	socketsLock.Unlock(c.vmId.Hex(), c.lockId)
}

func (c *Connection) Send(command interface{}, resp interface{}) (
	err error) {

	cmdData, err := json.Marshal(command)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "qmp: Failed to marshal command"),
		}
		return
	}

	if !constants.Production {
		fmt.Println(string(cmdData))
	}

	cmdData = append(cmdData, '\n')

	_, err = c.sock.Write(cmdData)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qmp: Failed to write socket"),
		}
		return
	}

	var returnData []byte
	returnWait := make(chan bool, 2)

	go func() {
		defer func() {
			returnWait <- true
		}()

		for {
			buffer := make([]byte, 5000000)
			n, e := c.sock.Read(buffer)
			if e != nil {
				err = &errortypes.ReadError{
					errors.Wrap(e, "qmp: Failed to read socket"),
				}
				return
			}
			buffer = buffer[:n]

			lines := bytes.Split(buffer, []byte("\n"))
			for _, line := range lines {
				if !constants.Production {
					fmt.Println(string(line))
				}

				if bytes.Contains(line, []byte(`"return"`)) ||
					bytes.Contains(line, []byte(`"error"`)) {

					returnData = line
					returnWait <- true

					return
				}
			}
		}
	}()

	<-returnWait
	if err != nil {
		return
	}

	if returnData == nil {
		err = &errortypes.ReadError{
			errors.New("qmp: No data from socket"),
		}
		return
	}

	err = json.Unmarshal(returnData, resp)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrapf(
				err,
				"qmp: Failed to unmarshal return '%s'",
				string(returnData),
			),
		}
		return
	}

	return
}

func (c *Connection) Connect() (err error) {
	err = c.connect()
	if err != nil {
		return
	}

	initCmd := &Command{
		Execute: "qmp_capabilities",
	}

	initResp := &CommandReturn{}
	err = c.Send(initCmd, initResp)
	if err != nil {
		return
	}

	if initResp.Error != nil {
		err = &errortypes.ApiError{
			errors.Newf("qmp: Return error '%s'", initResp.Error.Desc),
		}
		return
	}

	return
}

func NewConnection(vmId primitive.ObjectID) (conn *Connection) {
	conn = &Connection{
		vmId: vmId,
	}

	return
}

func RunCommand(vmId primitive.ObjectID, cmd interface{},
	resp interface{}) (err error) {

	conn := NewConnection(vmId)
	defer conn.Close()

	err = conn.Connect()
	if err != nil {
		return
	}

	err = conn.Send(cmd, resp)
	if err != nil {
		return
	}

	return
}
