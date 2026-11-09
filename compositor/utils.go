package compositor

import (
	"context"
	stderrors "errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

func readProc(pidStr, name string) (data []byte, err error) {
	data, err = os.ReadFile(filepath.Join(ProcDir, pidStr, name))
	if err != nil {
		if os.IsNotExist(err) || stderrors.Is(err, syscall.ESRCH) {
			data = nil
			err = nil
			return
		}

		err = &errortypes.ReadError{
			errors.Wrapf(
				err,
				"compositor: Failed to read proc '%s' %s",
				pidStr,
				name,
			),
		}
		return
	}

	return
}

func matchUid(status []byte, uid uint32) bool {
	for _, line := range strings.Split(string(status), "\n") {
		if !strings.HasPrefix(line, "Uid:") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			return false
		}

		for _, field := range fields[1:3] {
			fieldUid, err := strconv.ParseUint(field, 10, 32)
			if err != nil || uint32(fieldUid) != uid {
				return false
			}
		}

		return true
	}

	return false
}

func parseEnviron(data []byte) map[string]string {
	env := map[string]string{}

	for _, entry := range strings.Split(string(data), "\x00") {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}

		switch parts[0] {
		case "DISPLAY", "WAYLAND_DISPLAY", "XAUTHORITY", "XDG_RUNTIME_DIR",
			"XDG_SESSION_ID":

			env[parts[0]] = parts[1]
		}
	}

	return env
}

func socketOwner(path string) (uid uint32, ok bool) {
	info, err := os.Stat(path)
	if err != nil || info.Mode()&os.ModeSocket == 0 {
		return
	}

	stat, isStat := info.Sys().(*syscall.Stat_t)
	if !isStat {
		return
	}

	uid = stat.Uid
	ok = true
	return
}

func x11SocketPath(display string) string {
	if !strings.HasPrefix(display, ":") {
		return ""
	}

	num := display[1:]
	if i := strings.IndexByte(num, '.'); i >= 0 {
		num = num[:i]
	}

	if _, err := strconv.ParseUint(num, 10, 32); err != nil {
		return ""
	}

	return filepath.Join(X11SocketDir, "X"+num)
}

func validSessionId(session string) bool {
	if session == "" {
		return false
	}

	for _, c := range session {
		if (c < '0' || c > '9') && (c < 'a' || c > 'z') &&
			(c < 'A' || c > 'Z') {

			return false
		}
	}

	return true
}

func getActiveSession(uid string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	output, err := exec.CommandContext(ctx, "loginctl", "show-user", uid,
		"--property=Display", "--value", "--no-pager").Output()
	if err != nil {
		return ""
	}

	session := strings.TrimSpace(string(output))
	if !validSessionId(session) {
		return ""
	}

	output, err = exec.CommandContext(ctx, "loginctl", "show-session",
		session, "--property=Active", "--property=Type",
		"--no-pager").Output()
	if err != nil {
		return ""
	}

	props := map[string]string{}
	for _, line := range strings.Split(string(output), "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			props[parts[0]] = strings.TrimSpace(parts[1])
		}
	}

	if props["Active"] != "yes" ||
		(props["Type"] != "wayland" && props["Type"] != "x11") {

		return ""
	}

	return session
}
