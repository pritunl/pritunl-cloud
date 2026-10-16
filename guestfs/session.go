package guestfs

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

const (
	Qcow2 = "qcow2"
	Raw   = "raw"

	closeTimeout   = 10 * time.Minute
	acquireTimeout = 30 * time.Minute
	outputLimit    = 4 * 1024 * 1024
)

var (
	commandRe = regexp.MustCompile(`^[A-Za-z0-9-]+$`)
	argRe     = regexp.MustCompile(`^[A-Za-z0-9_./:+-]+$`)
	quotedRe  = regexp.MustCompile(`^"(\\x[0-9a-f]{2})*"$`)

	environKeys = set.NewSet(
		"LIBGUESTFS_BACKEND",
		"LIBGUESTFS_ATTACH_METHOD",
		"LIBGUESTFS_DEBUG",
		"LIBGUESTFS_TRACE",
		"LIBGUESTFS_TMPDIR",
		"LIBGUESTFS_CACHEDIR",
		"LIBGUESTFS_MEMSIZE",
		"TMPDIR",
		"LANG",
		"LC_ALL",
		"LANGUAGE",
	)

	sessionsLock = sync.Mutex{}
	sessionsSem  chan struct{}
)

type Session struct {
	path     string
	format   string
	readOnly bool
	device   string
	token    string
	timeout  time.Duration

	cmd     *exec.Cmd
	stdin   io.WriteCloser
	lines   chan string
	exited  chan struct{}
	waitErr error

	lock     sync.Mutex
	lineNo   int
	dead     bool
	closed   bool
	released bool
}

func sessionsLimit() (sem chan struct{}) {
	sessionsLock.Lock()
	defer sessionsLock.Unlock()

	if sessionsSem == nil {
		limit := settings.Hypervisor.GuestfsSessions
		if limit < 1 {
			limit = 1
		}
		sessionsSem = make(chan struct{}, limit)
	}

	sem = sessionsSem
	return
}

func acquire() (err error) {
	sem := sessionsLimit()

	select {
	case sem <- struct{}{}:
		return
	case <-time.After(acquireTimeout):
		err = &errortypes.TimeoutError{
			errors.New("guestfs: Timeout waiting for session slot"),
		}
		return
	}
}

func release() {
	sem := sessionsLimit()

	select {
	case <-sem:
	default:
	}
}

func environ() (env []string) {
	tempPath := paths.GetTempPath()

	for _, item := range os.Environ() {
		key := strings.SplitN(item, "=", 2)[0]
		if environKeys.Contains(key) {
			continue
		}
		env = append(env, item)
	}

	env = append(env,
		"LIBGUESTFS_BACKEND=direct",
		"LIBGUESTFS_TMPDIR="+tempPath,
		"TMPDIR="+tempPath,
		"LC_ALL=C",
		"LANG=C",
	)

	if settings.Hypervisor.GuestfsCachePath != "" {
		env = append(env, "LIBGUESTFS_CACHEDIR="+
			settings.Hypervisor.GuestfsCachePath)
	}

	if settings.Hypervisor.GuestfsMemory > 0 {
		env = append(env, fmt.Sprintf("LIBGUESTFS_MEMSIZE=%d",
			settings.Hypervisor.GuestfsMemory))
	}

	return
}

func newToken() (token string, err error) {
	buf := make([]byte, 12)

	_, err = rand.Read(buf)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "guestfs: Failed to generate session token"),
		}
		return
	}

	token = hex.EncodeToString(buf)

	return
}

func Open(pth, format string, readOnly bool) (sess *Session, err error) {
	if format != Qcow2 && format != Raw {
		err = &errortypes.ParseError{
			errors.Newf("guestfs: Invalid disk format %s", format),
		}
		return
	}

	timeout := time.Duration(settings.Hypervisor.GuestfsTimeout) *
		time.Second
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}

	token, err := newToken()
	if err != nil {
		return
	}

	err = utils.ExistsMkdir(paths.GetTempPath(), 0755)
	if err != nil {
		return
	}

	err = acquire()
	if err != nil {
		return
	}

	sess = &Session{
		path:     pth,
		format:   format,
		readOnly: readOnly,
		token:    token,
		timeout:  timeout,
		lines:    make(chan string, 256),
		exited:   make(chan struct{}),
	}

	err = sess.start()
	if err != nil {
		sess.release()
		sess = nil
		return
	}

	logrus.WithFields(logrus.Fields{
		"path":      pth,
		"format":    format,
		"read_only": readOnly,
	}).Info("guestfs: Launching appliance")

	start := time.Now()

	_, err = sess.Run("run")
	if err != nil {
		_ = sess.Close()
		sess = nil
		return
	}

	devices, err := sess.ListDevices()
	if err != nil {
		_ = sess.Close()
		sess = nil
		return
	}

	if len(devices) != 1 {
		err = &errortypes.ReadError{
			errors.Newf("guestfs: Expected one device found %d",
				len(devices)),
		}
		_ = sess.Close()
		sess = nil
		return
	}
	sess.device = devices[0]

	logrus.WithFields(logrus.Fields{
		"path":      pth,
		"device":    sess.device,
		"read_only": readOnly,
		"launch":    time.Since(start).Round(time.Millisecond).String(),
	}).Info("guestfs: Appliance ready")

	return
}

func (s *Session) start() (err error) {
	mode := "--rw"
	if s.readOnly {
		mode = "--ro"
	}

	cmd := exec.Command(
		"guestfish",
		mode,
		"--format="+s.format,
		"-a", s.path,
	)
	cmd.Env = environ()
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "guestfs: Failed to open guestfish stdin"),
		}
		return
	}

	pipeRead, pipeWrite, err := os.Pipe()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "guestfs: Failed to create guestfish pipe"),
		}
		return
	}

	cmd.Stdout = pipeWrite
	cmd.Stderr = pipeWrite

	err = cmd.Start()
	pipeWrite.Close()
	if err != nil {
		pipeRead.Close()
		err = &errortypes.ExecError{
			errors.Wrap(err, "guestfs: Failed to start guestfish"),
		}
		return
	}

	s.cmd = cmd
	s.stdin = stdin

	go s.reader(pipeRead)

	return
}

func (s *Session) reader(pipeRead *os.File) {
	defer pipeRead.Close()

	reader := bufio.NewReaderSize(pipeRead, 64*1024)
	for {
		line, err := reader.ReadString('\n')
		if line != "" {
			s.lines <- strings.TrimRight(line, "\r\n")
		}
		if err != nil {
			break
		}
	}

	s.waitErr = s.cmd.Wait()
	close(s.exited)
	close(s.lines)
}

func (s *Session) release() {
	if s.released {
		return
	}
	s.released = true
	release()
}

func (s *Session) Device() string {
	return s.device
}

func (s *Session) write(line string) (lineNo int, err error) {
	s.lineNo += 1
	lineNo = s.lineNo

	_, err = io.WriteString(s.stdin, line+"\n")
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "guestfs: Failed to write guestfish command"),
		}
		return
	}

	return
}

func (s *Session) kill() {
	if s.cmd == nil || s.cmd.Process == nil {
		return
	}

	pgid, err := syscall.Getpgid(s.cmd.Process.Pid)
	if err == nil && pgid == s.cmd.Process.Pid {
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
	} else {
		_ = s.cmd.Process.Kill()
	}
}

func (s *Session) drain(timeout time.Duration) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case _, ok := <-s.lines:
			if !ok {
				return
			}
		case <-timer.C:
			s.kill()
			timer.Reset(10 * time.Second)
		}
	}
}

func validateArgs(args []string) (err error) {
	if len(args) == 0 {
		err = &errortypes.ParseError{
			errors.New("guestfs: Missing command"),
		}
		return
	}

	if !commandRe.MatchString(args[0]) {
		err = &errortypes.ParseError{
			errors.Newf("guestfs: Invalid command '%s'", args[0]),
		}
		return
	}

	for _, arg := range args[1:] {
		if !argRe.MatchString(arg) && !quotedRe.MatchString(arg) {
			err = &errortypes.ParseError{
				errors.Newf("guestfs: Invalid command argument '%s'", arg),
			}
			return
		}
	}

	return
}

func (s *Session) run(optional bool, args []string) (
	output string, ok bool, err error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	err = validateArgs(args)
	if err != nil {
		return
	}

	if s.closed || s.dead {
		err = &errortypes.ExecError{
			errors.Newf("guestfs: Session closed before '%s'", args[0]),
		}
		return
	}

	line := strings.Join(args, " ")
	if optional {
		line = "-" + line
	}

	lineNo, err := s.write(line)
	if err != nil {
		s.dead = true
		s.kill()
		return
	}

	sentinel := fmt.Sprintf("%s-%d", s.token, lineNo)

	_, err = s.write("echo " + sentinel)
	if err != nil {
		s.dead = true
		s.kill()
		return
	}

	timer := time.NewTimer(s.timeout)
	defer timer.Stop()

	buf := strings.Builder{}
	failed := false
	errPrefix := fmt.Sprintf("*stdin*:%d:", lineNo)

	for {
		select {
		case text, alive := <-s.lines:
			if !alive {
				s.dead = true
				err = &errortypes.ExecError{
					errors.Newf("guestfs: Command '%s' failed: %s",
						args[0], strings.TrimSpace(buf.String())),
				}
				s.logError(args, buf.String())
				return
			}

			if text == sentinel {
				if failed {
					ok = false
					output = ""
					return
				}

				ok = true
				output = buf.String()
				return
			}

			if strings.HasPrefix(text, errPrefix) {
				failed = true
			}

			if buf.Len() < outputLimit {
				buf.WriteString(text)
				buf.WriteString("\n")
			}
		case <-timer.C:
			s.dead = true
			s.kill()
			err = &errortypes.TimeoutError{
				errors.Newf("guestfs: Command '%s' timed out", args[0]),
			}
			s.logError(args, buf.String())
			return
		}
	}
}

func (s *Session) logError(args []string, output string) {
	logrus.WithFields(logrus.Fields{
		"path":    s.path,
		"command": strings.Join(args, " "),
		"output":  strings.TrimSpace(output),
	}).Error("guestfs: Command failed")
}

func (s *Session) Run(args ...string) (output string, err error) {
	output, _, err = s.run(false, args)
	return
}

func (s *Session) RunOptional(args ...string) (
	output string, ok bool, err error) {

	return s.run(true, args)
}

func (s *Session) Close() (err error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.closed {
		return
	}
	s.closed = true
	defer s.release()

	if s.cmd == nil {
		return
	}

	if !s.dead {
		_, e := s.write("exit")
		if e != nil {
			s.kill()
		}
	}

	_ = s.stdin.Close()

	s.drain(closeTimeout)

	<-s.exited

	if !s.dead && s.waitErr != nil {
		err = &errortypes.ExecError{
			errors.Wrap(s.waitErr, "guestfs: Appliance shutdown failed"),
		}

		logrus.WithFields(logrus.Fields{
			"path":  s.path,
			"error": err,
		}).Error("guestfs: Appliance shutdown failed")
		return
	}

	return
}

func parseInt(val string) (num int64, err error) {
	num, err = strconv.ParseInt(strings.TrimSpace(val), 10, 64)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrapf(err, "guestfs: Failed to parse number '%s'",
				strings.TrimSpace(val)),
		}
		return
	}

	return
}
