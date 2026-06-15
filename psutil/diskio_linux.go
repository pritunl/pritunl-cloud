package psutil

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

var diskIoNameReg = regexp.MustCompile(
	`^(sd[a-z]+|vd[a-z]+|xvd[a-z]+|nvme\d+n\d+|dm-\d+|mmcblk\d+)$`)

func diskIoList() (stats []*diskIoStat, err error) {
	file, err := os.Open("/proc/diskstats")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to open diskstats"),
		}
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 14 {
			continue
		}

		name := fields[2]
		if !diskIoNameReg.MatchString(name) {
			continue
		}

		stats = append(stats, &diskIoStat{
			Name:       name,
			CountRead:  parseUint(fields[3]),
			BytesRead:  parseUint(fields[5]) * 512,
			TimeRead:   parseUint(fields[6]),
			CountWrite: parseUint(fields[7]),
			BytesWrite: parseUint(fields[9]) * 512,
			TimeWrite:  parseUint(fields[10]),
			TimeIo:     parseUint(fields[12]),
		})
	}

	err = scanner.Err()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to read diskstats"),
		}
		return
	}

	return
}
