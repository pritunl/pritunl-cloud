package logging

import (
	"bufio"
	"context"
	"encoding/json"
	"os/exec"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/tools/commander"
	"github.com/pritunl/tools/logger"
)

const journalMaxCapacity = 2 * 1024 * 1024

type journalEntry struct {
	Message   journalMessage `json:"MESSAGE"`
	Priority  string         `json:"PRIORITY"`
	Timestamp string         `json:"__REALTIME_TIMESTAMP"`
}

type journalMessage string

func (m *journalMessage) UnmarshalJSON(data []byte) (err error) {
	var msgStr string
	if json.Unmarshal(data, &msgStr) == nil {
		*m = journalMessage(msgStr)
		return
	}

	var msgBytes []byte
	if json.Unmarshal(data, &msgBytes) == nil {
		*m = journalMessage(msgBytes)
		return
	}

	var msgVals []journalMessage
	if json.Unmarshal(data, &msgVals) == nil {
		vals := make([]string, 0, len(msgVals))
		for _, val := range msgVals {
			vals = append(vals, string(val))
		}

		*m = journalMessage(strings.Join(vals, "\n"))
		return
	}

	*m = ""
	return
}

type Systemd struct {
	unit   string
	output chan *types.Entry
	cmd    *exec.Cmd
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *Systemd) GetOutput() (entries []*types.Entry) {
	for {
		select {
		case entry := <-s.output:
			entries = append(entries, entry)
		default:
			return
		}
	}
}

func (s *Systemd) followJournal() (err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			logger.WithFields(logger.Fields{
				"unit":  s.unit,
				"panic": rec,
			}).Error("agent: Journal follower panic")
		}
	}()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		resp, _ := commander.Exec(&commander.Opt{
			Name: "journalctl",
			Args: []string{
				"-q",
				"-b",
				"-n", "1",
				"-o", "cat",
				"--all",
				"-u", s.unit,
			},
			Timeout: 10 * time.Second,
			PipeOut: true,
		})

		if resp != nil && len(resp.Output) > 0 {
			break
		}

		select {
		case <-time.After(800 * time.Millisecond):
		case <-s.ctx.Done():
			return
		}
	}

	s.cmd = exec.CommandContext(s.ctx, "journalctl",
		"-f",
		"-b",
		"-n", "20",
		"-o", "json",
		"--all",
		"-u", s.unit,
	)

	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "agent: Error creating stdout pipe"),
		}
		return
	}

	err = s.cmd.Start()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "agent: Error starting journalctl"),
		}
		return
	}

	firstLine := make(chan struct{})
	followDone := make(chan struct{})
	defer close(followDone)

	watchCmd := s.cmd
	go func() {
		select {
		case <-firstLine:
		case <-followDone:
		case <-s.ctx.Done():
		case <-time.After(20 * time.Second):
			logger.WithFields(logger.Fields{
				"unit": s.unit,
			}).Warn("agent: Journal follower stalled, restarting")

			if watchCmd.Process != nil {
				watchCmd.Process.Kill()
			}
		}
	}()

	scanner := bufio.NewScanner(stdout)
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, journalMaxCapacity)

	first := true
	for scanner.Scan() {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		if first {
			first = false
			close(firstLine)
		}

		line := scanner.Bytes()

		var entry journalEntry
		e := json.Unmarshal(line, &entry)
		if e != nil {
			continue
		}

		var timestamp time.Time
		ts, e := strconv.ParseInt(entry.Timestamp, 10, 64)
		if e == nil {
			timestamp = time.Unix(0, ts*1000)
		} else {
			timestamp = time.Now()
		}

		level := int32(5)
		if entry.Priority != "" {
			switch entry.Priority {
			case "0":
				level = 1
			case "1", "2":
				level = 2
			case "3":
				level = 3
			case "4":
				level = 4
			case "5", "6":
				level = 5
			case "7":
				level = 6
			}
		}

		select {
		case s.output <- &types.Entry{
			Timestamp: timestamp,
			Level:     level,
			Message:   strings.TrimSuffix(string(entry.Message), "\n"),
		}:
		default:
		}
	}

	err = scanner.Err()
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "agent: Error reading journal"),
		}
		return
	}

	s.cmd.Wait()

	return
}

func (s *Systemd) Open() (err error) {
	s.output = make(chan *types.Entry, 10000)
	s.ctx, s.cancel = context.WithCancel(context.Background())

	go func() {
		defer func() {
			panc := recover()
			if panc != nil {
				logger.WithFields(logger.Fields{
					"trace": string(debug.Stack()),
					"panic": panc,
				}).Error("sync: Panic in journal open")
			}
		}()

		for {
			select {
			case <-s.ctx.Done():
				return
			default:
			}

			e := s.followJournal()

			select {
			case <-s.ctx.Done():
				return
			default:
			}

			if e != nil {
				logger.WithFields(logger.Fields{
					"unit":  s.unit,
					"error": e,
				}).Error("agent: Journal follower error, restarting")
			} else {
				logger.WithFields(logger.Fields{
					"unit": s.unit,
				}).Info("agent: Journal follower exited, restarting")
			}

			select {
			case <-time.After(3 * time.Second):
			case <-s.ctx.Done():
				return
			}
		}
	}()

	return
}

func (s *Systemd) Close() (err error) {
	defer func() {
		panc := recover()
		if panc != nil {
			logger.WithFields(logger.Fields{
				"trace": string(debug.Stack()),
				"panic": panc,
			}).Error("sync: Panic in journal close")
		}
	}()

	if s.cancel != nil {
		s.cancel()
	}

	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
	}

	if s.output != nil {
		close(s.output)
	}

	return
}

func NewSystemd(unit string) *Systemd {
	return &Systemd{
		unit: unit,
	}
}
