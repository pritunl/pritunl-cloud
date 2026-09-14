package psutil

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

func isPid(name string) bool {
	if name == "" {
		return false
	}

	for _, c := range name {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

func processName(pid string) (name string) {
	exe, err := os.Readlink(filepath.Join("/proc", pid, "exe"))
	if err == nil {
		exe = strings.TrimSuffix(exe, " (deleted)")
		name = filepath.Base(exe)
		if name != "" && name != "." && name != "/" {
			return
		}
	}

	cmdline, err := os.ReadFile(filepath.Join("/proc", pid, "cmdline"))
	if err != nil || len(cmdline) == 0 {
		return ""
	}

	arg := cmdline
	if idx := bytes.IndexByte(cmdline, 0); idx >= 0 {
		arg = cmdline[:idx]
	}
	arg = bytes.TrimSpace(arg)
	if len(arg) == 0 {
		return ""
	}

	name = filepath.Base(string(arg))
	if name == "." || name == "/" {
		return ""
	}

	return
}

func processNames() (names []string, err error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to read proc"),
		}
		return
	}

	seen := map[string]bool{}
	names = []string{}

	for _, entry := range entries {
		if !entry.IsDir() || !isPid(entry.Name()) {
			continue
		}

		name := processName(entry.Name())
		if name == "" || seen[name] {
			continue
		}

		seen[name] = true
		names = append(names, name)
	}

	return
}
