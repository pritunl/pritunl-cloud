package logging

import (
	"bufio"
	"context"
	"os/exec"
	"runtime/debug"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/tools/logger"
)

type File struct {
	path   string
	output chan *types.Entry
	cmd    *exec.Cmd
	ctx    context.Context
	cancel context.CancelFunc
}

func (f *File) GetOutput() (entries []*types.Entry) {
	for {
		select {
		case entry := <-f.output:
			entries = append(entries, entry)
		default:
			return
		}
	}
}

func (f *File) followJournal() (err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			logger.WithFields(logger.Fields{
				"path":  f.path,
				"panic": rec,
			}).Error("agent: File follower panic")
		}
	}()

	f.cmd = exec.CommandContext(f.ctx, "tail",
		"-F", f.path,
	)

	stdout, err := f.cmd.StdoutPipe()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "agent: Error creating stdout pipe"),
		}
		return
	}

	err = f.cmd.Start()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "agent: Error starting tail"),
		}
		return
	}

	scanner := bufio.NewScanner(stdout)
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		select {
		case <-f.ctx.Done():
			return
		default:
		}

		line := scanner.Text()

		timestamp := time.Now()
		level := int32(5)

		if strings.Contains(strings.ToLower(line), "error") {
			level = 3
		}

		select {
		case f.output <- &types.Entry{
			Timestamp: timestamp,
			Level:     level,
			Message:   line,
		}:
		default:
		}
		continue
	}

	err = scanner.Err()
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "agent: Error reading tail"),
		}
		return
	}

	f.cmd.Wait()

	return
}

func (f *File) Open() (err error) {
	f.output = make(chan *types.Entry, 10000)
	f.ctx, f.cancel = context.WithCancel(context.Background())

	go func() {
		defer func() {
			panc := recover()
			if panc != nil {
				logger.WithFields(logger.Fields{
					"trace": string(debug.Stack()),
					"panic": panc,
				}).Error("sync: Panic in tail open")
			}
		}()

		for {
			select {
			case <-f.ctx.Done():
				return
			default:
			}

			e := f.followJournal()

			select {
			case <-f.ctx.Done():
				return
			default:
			}

			if e != nil {
				logger.WithFields(logger.Fields{
					"path":  f.path,
					"error": e,
				}).Error("agent: Journal follower error, restarting")
			} else {
				logger.WithFields(logger.Fields{
					"path": f.path,
				}).Info("agent: Journal follower exited, restarting")
			}

			select {
			case <-time.After(3 * time.Second):
			case <-f.ctx.Done():
				return
			}
		}
	}()

	return
}

func (f *File) Close() (err error) {
	defer func() {
		panc := recover()
		if panc != nil {
			logger.WithFields(logger.Fields{
				"trace": string(debug.Stack()),
				"panic": panc,
			}).Error("sync: Panic in journal close")
		}
	}()

	if f.cancel != nil {
		f.cancel()
	}

	if f.cmd != nil && f.cmd.Process != nil {
		f.cmd.Process.Kill()
	}

	if f.output != nil {
		close(f.output)
	}

	return
}

func NewFile(path string) *File {
	return &File{
		path: path,
	}
}
