package psutil

import (
	"bufio"
	"os"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

func networkList() (stats []*networkStat, err error) {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to open net dev"),
		}
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		idx := strings.IndexByte(line, ':')
		if idx < 0 {
			continue
		}

		name := strings.TrimSpace(line[:idx])
		if name == "" || name == "lo" ||
			strings.HasPrefix(name, "veth") ||
			strings.HasPrefix(name, "docker") ||
			strings.HasPrefix(name, "br-") {

			continue
		}

		fields := strings.Fields(line[idx+1:])
		if len(fields) < 16 {
			continue
		}

		stats = append(stats, &networkStat{
			Name:        name,
			BytesRecv:   parseUint(fields[0]),
			PacketsRecv: parseUint(fields[1]),
			ErrorsRecv:  parseUint(fields[2]),
			DropsRecv:   parseUint(fields[3]),
			FifoRecv:    parseUint(fields[4]),
			BytesSent:   parseUint(fields[8]),
			PacketsSent: parseUint(fields[9]),
			ErrorsSent:  parseUint(fields[10]),
			DropsSent:   parseUint(fields[11]),
			FifoSent:    parseUint(fields[12]),
		})
	}

	err = scanner.Err()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to read net dev"),
		}
		return
	}

	return
}
