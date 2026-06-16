package psutil

import (
	"encoding/binary"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"golang.org/x/sys/unix"
)

const (
	cpuStates = 5
	cpuIdle   = 4
)

func cpuTimes() (total, idle uint64, err error) {
	buf, err := unix.SysctlRaw("kern.cp_time")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to read cp_time"),
		}
		return
	}

	if len(buf) < cpuStates*8 {
		return
	}

	for i := 0; i < cpuStates; i++ {
		val := binary.NativeEndian.Uint64(buf[i*8 : i*8+8])
		total += val
		if i == cpuIdle {
			idle += val
		}
	}

	return
}

func GetProcesses() (count uint64, err error) {
	buf, err := unix.SysctlRaw("kern.proc.proc")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to read proc"),
		}
		return
	}

	if len(buf) < 4 {
		return
	}

	structSize := binary.NativeEndian.Uint32(buf[0:4])
	if structSize == 0 {
		return
	}

	count = uint64(len(buf)) / uint64(structSize)

	return
}
