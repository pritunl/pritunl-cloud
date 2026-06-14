package psutil

import (
	"bufio"
	"os"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

func cpuTimes() (total, idle uint64, err error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to open stat"),
		}
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return
	}

	fields := strings.Fields(scanner.Text())
	if len(fields) < 5 || fields[0] != "cpu" {
		return
	}

	for i := 1; i < len(fields); i++ {
		val := parseUint(fields[i])
		total += val
		if i == 4 || i == 5 {
			idle += val
		}
	}

	return
}

func GetProcesses() (count uint64, err error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to read loadavg"),
		}
		return
	}

	fields := strings.Fields(string(data))
	if len(fields) < 4 {
		return
	}

	parts := strings.Split(fields[3], "/")
	if len(parts) != 2 {
		return
	}

	count = parseUint(parts[1])

	return
}
