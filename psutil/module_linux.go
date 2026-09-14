package psutil

import (
	"bufio"
	"os"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

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

		name := fields[0]
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

	return
}
