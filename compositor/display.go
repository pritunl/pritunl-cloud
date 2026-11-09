package compositor

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

const (
	ProcDir      = "/proc"
	RunUserDir   = "/run/user"
	X11SocketDir = "/tmp/.X11-unix"
)

type Display struct {
	Display        string
	WaylandDisplay string
	Xauthority     string
	XdgRuntime     string
}

func (d *Display) better(other *Display) bool {
	if other == nil {
		return true
	}
	if (d.WaylandDisplay != "") != (other.WaylandDisplay != "") {
		return d.WaylandDisplay != ""
	}
	return d.Xauthority != "" && other.Xauthority == ""
}

func GetDisplay(username string) (dsp *Display, err error) {
	unixUser, err := user.Lookup(username)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "compositor: Failed to find GUI user"),
		}
		return
	}

	uid64, err := strconv.ParseUint(unixUser.Uid, 10, 32)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "compositor: Invalid GUI user UID"),
		}
		return
	}
	uid := uint32(uid64)

	files, err := os.ReadDir(ProcDir)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "compositor: Failed to read proc directory"),
		}
		return
	}

	activeSession := getActiveSession(unixUser.Uid)
	runtimeDir := filepath.Join(RunUserDir, unixUser.Uid)
	xauthPath := filepath.Join(unixUser.HomeDir, ".Xauthority")

	var active, other *Display
	var readErr error

	for _, file := range files {
		pidStr := file.Name()
		pid, e := strconv.Atoi(pidStr)
		if e != nil || pid <= 0 || !file.IsDir() {
			continue
		}

		status, e := readProc(pidStr, "status")
		if e != nil {
			if readErr == nil {
				readErr = e
			}
			continue
		}
		if len(status) == 0 || !matchUid(status, uid) {
			continue
		}

		procEnv, e := readProc(pidStr, "environ")
		if e != nil {
			if readErr == nil {
				readErr = e
			}
			continue
		}
		if len(procEnv) == 0 {
			continue
		}
		env := parseEnviron(procEnv)

		sessionId := env["XDG_SESSION_ID"]
		if activeSession != "" && sessionId != "" &&
			sessionId != activeSession {

			continue
		}

		cand := parseDisplay(env, uid, runtimeDir, xauthPath)
		if cand == nil {
			continue
		}

		if activeSession != "" && sessionId == activeSession {
			if cand.better(active) {
				active = cand
			}
		} else if cand.better(other) {
			other = cand
		}
	}

	dsp = active
	if dsp == nil {
		dsp = other
	}
	if dsp != nil {
		return
	}

	if readErr != nil {
		err = readErr
		return
	}

	err = &errortypes.ReadError{
		errors.Newf(
			"compositor: Failed to find graphical environment for '%s'",
			username,
		),
	}
	return
}

func parseDisplay(env map[string]string, uid uint32,
	runtimeDir, xauthPath string) *Display {

	if dir := env["XDG_RUNTIME_DIR"]; dir != "" {
		runtimeDir = dir
	}

	wayland := env["WAYLAND_DISPLAY"]
	if wayland != "" {
		socketPath := wayland
		if !filepath.IsAbs(socketPath) {
			socketPath = filepath.Join(runtimeDir, socketPath)
		}
		owner, ok := socketOwner(socketPath)
		if !ok || owner != uid {
			return nil
		}
	}

	display := env["DISPLAY"]
	if display != "" {
		socketPath := x11SocketPath(display)
		if socketPath == "" {
			display = ""
		} else if _, ok := socketOwner(socketPath); !ok {
			display = ""
		}
	}

	xauth := ""
	if display != "" {
		xauth = env["XAUTHORITY"]
		if xauth == "" {
			if _, err := os.Stat(xauthPath); err == nil {
				xauth = xauthPath
			}
		} else if _, err := os.Stat(xauth); err != nil {
			return nil
		}
	}

	if wayland == "" && display == "" {
		return nil
	}

	return &Display{
		Display:        display,
		WaylandDisplay: wayland,
		Xauthority:     xauth,
		XdgRuntime:     runtimeDir,
	}
}
