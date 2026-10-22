package qmp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Command struct {
	Execute   string      `json:"execute"`
	Arguments interface{} `json:"arguments,omitempty"`
}

type CommandId struct {
	Id string `json:"id"`
}

type CommandNode struct {
	NodeName string `json:"node-name"`
}

type JobStatus struct {
	Id              string `json:"id"`
	Type            string `json:"type"`
	Status          string `json:"status"`
	CurrentProgress int64  `json:"current-progress"`
	TotalProgress   int64  `json:"total-progress"`
	Error           string `json:"error"`
}

type JobStatusReturn struct {
	Return []*JobStatus  `json:"return"`
	Error  *CommandError `json:"error"`
}

type QmpVersionData struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Micro int `json:"micro"`
}

type QmpVersion struct {
	Qemu QmpVersionData `json:"qemu"`
}

type QmpData struct {
	Version QmpVersion `json:"version"`
}

type QmpCapabilities struct {
	QMP QmpData `json:"QMP"`
}

type QemuInfo struct {
	VersionMajor int
	VersionMinor int
	VersionMicro int
}

type CommandError struct {
	Class string `json:"class"`
	Desc  string `json:"desc"`
}

type CommandReturn struct {
	Return interface{}   `json:"return"`
	Error  *CommandError `json:"error"`
}

type EventCallback func() (resp interface{}, err error)

const (
	readerSize = 64 * 1024
)

var (
	socketsLock = utils.NewMultiTimeoutLock(1 * time.Minute)
)

type Connection struct {
	vmId     bson.ObjectID
	sock     net.Conn
	reader   *bufio.Reader
	lockId   bson.ObjectID
	deadline time.Duration
	logging  bool
	command  interface{}
	response interface{}
}

func (c *Connection) connect() (info *QemuInfo, err error) {
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

	deadline := c.deadline
	if deadline == 0 {
		deadline = 6 * time.Second
	}

	err = c.sock.SetDeadline(time.Now().Add(deadline))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "qmp: Failed set deadline"),
		}
		return
	}

	c.reader = bufio.NewReaderSize(c.sock, readerSize)

	var infoByt []byte
	for {
		line, e := c.readLine()
		if e != nil {
			err = e
			return
		}

		if isGreeting(line) {
			infoByt = line
			break
		}
	}

	connInfo := &QmpCapabilities{}

	err = json.Unmarshal(infoByt, connInfo)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrapf(
				err,
				"qmp: Failed to unmarshal info '%s'",
				string(infoByt),
			),
		}
		return
	}

	info = &QemuInfo{
		VersionMajor: connInfo.QMP.Version.Qemu.Major,
		VersionMinor: connInfo.QMP.Version.Qemu.Minor,
		VersionMicro: connInfo.QMP.Version.Qemu.Micro,
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

func (c *Connection) SetDeadline(deadline time.Duration) {
	c.deadline = deadline
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

	if !constants.Production && c.logging {
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
	for {
		line, e := c.readLine()
		if e != nil {
			err = e
			return
		}

		if isResponse(line) {
			returnData = line
			break
		}
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

func messageKeys(line []byte) (keys map[string]json.RawMessage, ok bool) {
	keys = map[string]json.RawMessage{}

	err := json.Unmarshal(line, &keys)
	if err != nil {
		return
	}

	ok = true

	return
}

func isGreeting(line []byte) bool {
	keys, ok := messageKeys(line)
	if !ok {
		return false
	}

	_, ok = keys["QMP"]

	return ok
}

func isEvent(line []byte) bool {
	keys, ok := messageKeys(line)
	if !ok {
		return false
	}

	_, ok = keys["event"]

	return ok
}

func isResponse(line []byte) bool {
	keys, ok := messageKeys(line)
	if !ok {
		return false
	}

	_, ok = keys["event"]
	if ok {
		return false
	}

	_, hasReturn := keys["return"]
	_, hasError := keys["error"]

	return hasReturn || hasError
}

func (c *Connection) readLine() (line []byte, err error) {
	for {
		data, e := c.reader.ReadBytes('\n')
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "qmp: Failed to read socket"),
			}
			return
		}

		line = bytes.TrimSpace(data)
		if len(line) == 0 {
			continue
		}

		if !constants.Production && c.logging {
			fmt.Println(string(line))
		}

		return
	}
}

func (c *Connection) Event(resp interface{}, callback EventCallback) (
	err error) {

	for {
		line, e := c.readLine()
		if e != nil {
			err = e
			return
		}

		if !isEvent(line) {
			continue
		}

		err = json.Unmarshal(line, resp)
		if err != nil {
			err = &errortypes.ParseError{
				errors.Wrapf(
					err,
					"qmp: Failed to unmarshal return '%s'",
					string(line),
				),
			}
			return
		}

		resp, err = callback()
		if err != nil || resp == nil {
			return
		}
	}
}

func (c *Connection) Connect() (info *QemuInfo, err error) {
	info, err = c.connect()
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

func NewConnection(vmId bson.ObjectID, logging bool) (conn *Connection) {
	conn = &Connection{
		vmId:    vmId,
		logging: logging,
	}

	return
}

func RunCommand(vmId bson.ObjectID, cmd interface{},
	resp interface{}) (err error) {

	conn := NewConnection(vmId, true)
	defer conn.Close()

	_, err = conn.Connect()
	if err != nil {
		return
	}

	err = conn.Send(cmd, resp)
	if err != nil {
		return
	}

	return
}
