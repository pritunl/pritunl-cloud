package features

import (
	"io/ioutil"
	"strconv"
	"strings"
	"syscall"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

const (
	Libexec = "/usr/libexec/qemu-kvm"
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
		path = Libexec
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

func GetKernelVersion() (major, minor, patch int, err error) {
	uname := &syscall.Utsname{}

	err = syscall.Uname(uname)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "qemu: Failed to get syscall uname"),
		}
		return
	}

	version := utils.Int8Str(uname.Release[:])

	versions := strings.Split(version, "-")
	if len(versions) < 2 {
		err = &errortypes.ParseError{
			errors.Newf(
				"qemu: Failed to parse uname version 1 '%s'",
				version,
			),
		}
		return
	}

	versions = strings.Split(versions[0], ".")
	if len(versions) < 3 {
		err = &errortypes.ParseError{
			errors.Newf(
				"qemu: Failed to parse uname version 2 '%s'",
				version,
			),
		}
		return
	}

	major, err = strconv.Atoi(versions[0])
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrapf(
				err,
				"qemu: Failed to parse uname version 3 '%s'",
				version,
			),
		}
		return
	}

	minor, err = strconv.Atoi(versions[1])
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrapf(
				err,
				"qemu: Failed to parse uname version 4 '%s'",
				version,
			),
		}
		return
	}

	patch, err = strconv.Atoi(versions[2])
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrapf(
				err,
				"qemu: Failed to parse uname version 5 '%s'",
				version,
			),
		}
		return
	}

	return
}

func GetUringSupport() (supported bool, err error) {
	kallsyms, err := ioutil.ReadFile("/proc/kallsyms")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrapf(err, "features: Failed to read /proc/kallsyms"),
		}
		return
	}

	if !strings.Contains(string(kallsyms), "io_uring_init") {
		return
	}

	major, minor, _, err := GetKernelVersion()
	if err != nil {
		return
	}

	if major < 5 {
		return
	} else if major == 5 && minor < 2 {
		return
	}

	major, minor, _, err = GetQemuVersion()
	if err != nil {
		return
	}

	if major < 6 {
		return
	} else if major == 6 && minor < 2 {
		return
	}

	supported = true
	return
}

func GetMemoryBackendSupport() (supported bool, err error) {
	major, _, _, err := GetQemuVersion()
	if err != nil {
		return
	}

	if major >= 6 {
		supported = true
	}

	return
}

func GetRunWithSupport() (supported bool, err error) {
	major, _, _, err := GetQemuVersion()
	if err != nil {
		return
	}

	if major >= 9 {
		supported = true
	}

	return
}
