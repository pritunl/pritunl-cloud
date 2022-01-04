package features

import (
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

const (
	Libvirt = "/usr/libexec/qemu-kvm"
	System  = "/usr/bin/qemu-system-x86_64"
)

func GetQemuPath() (path string, err error) {
	exists, err := utils.Exists(System)
	if err != nil {
		return
	}
	if exists {
		path = System
	} else {
		path = Libvirt
	}

	return
}

func GetQemuVersion() (major, minor, patch int, err error) {
	qemuPath, err := GetQemuPath()
	if err != nil {
		return
	}

	output, _ := utils.ExecCombinedOutputLogged(
		nil,
		qemuPath, "--version",
	)

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 4 || strings.ToLower(fields[2]) != "version" {
			continue
		}

		versions := strings.Split(fields[3], ".")
		if len(versions) != 3 {
			continue
		}

		var e error
		major, e = strconv.Atoi(versions[0])
		if e != nil {
			continue
		}

		minor, e = strconv.Atoi(versions[1])
		if e != nil {
			major = 0
			continue
		}

		patch, e = strconv.Atoi(versions[1])
		if e != nil {
			major = 0
			minor = 0
			continue
		}

		break
	}

	if major == 0 {
		err = &errortypes.ParseError{
			errors.Newf("qemu: Invalid Qemu version '%s'", output),
		}
		return
	}

	return
}

func GetUringSupport() (supported bool, err error) {
	major, minor, _, err := GetQemuVersion()
	if err != nil {
		return
	}

	if major > 6 {
		supported = true
	} else if major == 6 && minor >= 2 {
		supported = true
	}

	return
}
