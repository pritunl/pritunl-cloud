package psutil

import (
	"strconv"
	"strings"
	"time"

	"github.com/pritunl/tools/commander"
)

func moduleNames() (names []string, err error) {
	resp, err := commander.Exec(&commander.Opt{
		Name:    "kldstat",
		Args:    []string{"-v"},
		Timeout: 10 * time.Second,
		PipeOut: true,
		PipeErr: true,
	})
	if err != nil {
		return
	}

	seen := map[string]bool{}
	names = []string{}

	add := func(name string) {
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}

		name = strings.TrimSpace(name)
		if name == "" || name == "kernel" || seen[name] {
			return
		}

		seen[name] = true
		names = append(names, name)
	}

	for _, line := range strings.Split(string(resp.Output), "\n") {
		fields := strings.Fields(line)

		switch {
		case len(fields) >= 5 && fields[0] != "Id":
			add(strings.TrimSuffix(fields[4], ".ko"))
		case len(fields) == 2 && fields[0] != "Id" &&
			fields[0] != "Contains":

			if _, e := strconv.Atoi(fields[0]); e == nil {
				add(fields[1])
			}
		}
	}

	return
}
