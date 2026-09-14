package psutil

import (
	"strconv"
	"strings"
	"time"

	"github.com/pritunl/tools/commander"
)

func listenerList() (listeners []*Listener, err error) {
	resp, err := commander.Exec(&commander.Opt{
		Name: "sockstat",
		Args: []string{
			"-46lqw",
			"-P", "tcp,udp",
		},
		Timeout: 10 * time.Second,
		PipeOut: true,
		PipeErr: true,
	})
	if err != nil {
		return
	}

	seen := map[string]bool{}
	listeners = []*Listener{}

	for _, line := range strings.Split(string(resp.Output), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		name := fields[1]
		if name == "?" {
			name = ""
		}

		pid, e := strconv.Atoi(fields[2])
		if e != nil {
			pid = 0
		}

		proto := fields[4]
		local := fields[5]

		var protocol string
		switch {
		case strings.HasPrefix(proto, "tcp"):
			protocol = "tcp"
		case strings.HasPrefix(proto, "udp"):
			protocol = "udp"
		default:
			continue
		}

		var addr string
		var portStr string
		if strings.HasPrefix(local, "[") {
			end := strings.Index(local, "]:")
			if end < 0 {
				continue
			}

			addr = local[1:end]
			portStr = local[end+2:]
		} else {
			sep := strings.LastIndex(local, ":")
			if sep < 0 {
				continue
			}

			addr = local[:sep]
			portStr = local[sep+1:]
		}

		portNum, e := strconv.ParseUint(portStr, 10, 16)
		if e != nil {
			continue
		}

		if addr == "*" {
			if strings.HasSuffix(proto, "6") {
				addr = "::"
			} else {
				addr = "0.0.0.0"
			}
		}

		key := protocol + "|" + addr + "|" + portStr + "|" +
			strconv.Itoa(pid)
		if seen[key] {
			continue
		}

		seen[key] = true
		listeners = append(listeners, &Listener{
			Protocol: protocol,
			Address:  addr,
			Port:     uint16(portNum),
			Pid:      pid,
			Name:     name,
		})
	}

	return
}
