package psutil

import (
	"bufio"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

func normalizeModuleName(name string) string {
	return strings.ReplaceAll(strings.TrimSpace(name), "-", "_")
}

func sysfsModuleNames() (names []string, err error) {
	names = []string{}

	entries, err := os.ReadDir("/sys/module")
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			return
		}

		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to read sysfs modules"),
		}
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := normalizeModuleName(entry.Name())
		if name == "" {
			continue
		}

		names = append(names, name)
	}

	return
}

func builtinModinfoNames(modulesDir string) (names []string, err error) {
	names = []string{}

	data, err := os.ReadFile(filepath.Join(
		modulesDir, "modules.builtin.modinfo"))
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			return
		}

		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to read builtin modinfo"),
		}
		return
	}

	seen := map[string]bool{}
	for _, entry := range strings.Split(string(data), "\x00") {
		idx := strings.Index(entry, ".")
		if idx <= 0 {
			continue
		}

		name := normalizeModuleName(entry[:idx])
		if name == "" || seen[name] {
			continue
		}

		seen[name] = true
		names = append(names, name)
	}

	return
}

func builtinModuleNames() (names []string, err error) {
	names = []string{}

	release, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to read kernel release"),
		}
		return
	}

	modulesDir := filepath.Join("/lib/modules",
		strings.TrimSpace(string(release)))

	names, err = builtinModinfoNames(modulesDir)
	if err != nil {
		return
	}

	file, err := os.Open(filepath.Join(modulesDir, "modules.builtin"))
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			return
		}

		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to open builtin modules"),
		}
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		name := path.Base(strings.TrimSpace(scanner.Text()))
		if name == "" || name == "." || name == "/" {
			continue
		}

		if idx := strings.Index(name, ".ko"); idx > 0 {
			name = name[:idx]
		}

		name = normalizeModuleName(name)
		if name == "" {
			continue
		}

		names = append(names, name)
	}

	err = scanner.Err()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to read builtin modules"),
		}
		return
	}

	return
}

func moduleNames() (names []string, err error) {
	file, err := os.Open("/proc/modules")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to open modules"),
		}
		return
	}
	defer file.Close()

	seen := map[string]bool{}
	names = []string{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}

		name := normalizeModuleName(fields[0])
		if name == "" || seen[name] {
			continue
		}

		seen[name] = true
		names = append(names, name)
	}

	err = scanner.Err()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to read modules"),
		}
		return
	}

	builtin, err := builtinModuleNames()
	if err != nil {
		return
	}

	sysfs, err := sysfsModuleNames()
	if err != nil {
		return
	}

	for _, name := range append(builtin, sysfs...) {
		if seen[name] {
			continue
		}

		seen[name] = true
		names = append(names, name)
	}

	return
}
