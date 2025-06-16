package compositor

import (
	"fmt"
	"io/ioutil"
	"os/user"
	"path"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
)

const (
	ProcDir = "/proc"
)

func GetEnv(username, driPath string, driPrime bool) (
	envData string, err error) {

	desktopEnv := settings.Hypervisor.DesktopEnv

	files, err := ioutil.ReadDir(ProcDir)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "compositor: Failed to read proc directory"),
		}
		return
	}

	unixUser, err := user.Lookup(username)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "compositor: Failed to find GUI user"),
		}
		return
	}

	for _, file := range files {
		pidStr := file.Name()

		_, e := strconv.Atoi(pidStr)
		if e != nil {
			continue
		}

		cmdlinePath := path.Join(ProcDir, pidStr, "cmdline")
		environPath := path.Join(ProcDir, pidStr, "environ")
		loginuidPath := path.Join(ProcDir, pidStr, "loginuid")

		exists, e := utils.Exists(cmdlinePath)
		if e != nil {
			err = e
			return
		}
		if !exists {
			continue
		}

		exists, err = utils.Exists(environPath)
		if err != nil {
			return
		}
		if !exists {
			continue
		}

		exists, err = utils.Exists(loginuidPath)
		if err != nil {
			return
		}
		if !exists {
			continue
		}

		procUid, e := ioutil.ReadFile(loginuidPath)
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrapf(
					e,
					"compositor: Failed to read proc '%s' loginuid",
					pidStr,
				),
			}
			return
		}

		if strings.TrimSpace(string(procUid)) != unixUser.Uid {
			continue
		}

		procCmd, e := ioutil.ReadFile(cmdlinePath)
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrapf(
					e,
					"compositor: Failed to read proc '%s' cmdline",
					pidStr,
				),
			}
			return
		}

		if !strings.Contains(string(procCmd), desktopEnv) &&
			!strings.Contains(string(procCmd), "xdg") {
			continue
		}

		procEnv, e := ioutil.ReadFile(environPath)
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrapf(
					e,
					"compositor: Failed to read proc '%s' environ",
					pidStr,
				),
			}
			return
		}

		displayEnv := ""
		waylandDisplayEnv := ""
		xauthEnv := ""

		environ := strings.ReplaceAll(string(procEnv), "\000", "\n")
		for _, env := range strings.Split(environ, "\n") {
			envSpl := strings.SplitN(env, "=", 2)
			if len(envSpl) != 2 {
				continue
			}

			if strings.ToLower(envSpl[0]) == "display" {
				displayEnv = envSpl[1]
			} else if strings.ToLower(envSpl[0]) == "wayland_display" {
				waylandDisplayEnv = envSpl[1]
			} else if strings.ToLower(envSpl[0]) == "xauthority" {
				xauthEnv = envSpl[1]
			}
		}

		if displayEnv == "" || xauthEnv == "" {
			continue
		}

		envData += fmt.Sprintf("\nEnvironment=\"DISPLAY=%s\"", displayEnv)
		if waylandDisplayEnv != "" {
			envData += fmt.Sprintf(
				"\nEnvironment=\"WAYLAND_DISPLAY=%s\"", waylandDisplayEnv)
		}

		envData += fmt.Sprintf("\nEnvironment=\"XAUTHORITY=%s\"", xauthEnv)

		if driPath != "" {
			envData += fmt.Sprintf(
				"\nEnvironment=\"DRI_RENDER_DEVICE=%s\"",
				driPath,
			)
		}
		if driPrime {
			envData += "\nEnvironment=\"DRI_PRIME=1\""
		}

		return
	}

	err = &errortypes.ReadError{
		errors.New("compositor: Failed to find X environment"),
	}
	return
}
