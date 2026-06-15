package psutil

import (
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/metric"
	"github.com/pritunl/pritunl-cloud/utils"
	"golang.org/x/sys/unix"
)

var diskIgnoreFs = map[string]bool{
	"devfs":     true,
	"fdescfs":   true,
	"procfs":    true,
	"linprocfs": true,
	"linsysfs":  true,
	"nullfs":    true,
	"tmpfs":     true,
	"fusefs":    true,
	"virtiofs":  true,
}

func disksList() (disks []*metric.Mount, err error) {
	n, err := unix.Getfsstat(nil, unix.MNT_NOWAIT)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to get fsstat count"),
		}
		return
	}

	buf := make([]unix.Statfs_t, n)
	n, err = unix.Getfsstat(buf, unix.MNT_NOWAIT)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to get fsstat"),
		}
		return
	}
	if n < len(buf) {
		buf = buf[:n]
	}

	disks = []*metric.Mount{}
	seenMount := map[string]bool{}

	for i := range buf {
		stat := &buf[i]

		fsType := unix.ByteSliceToString(stat.Fstypename[:])
		mount := unix.ByteSliceToString(stat.Mntonname[:])

		if diskIgnoreFs[fsType] {
			continue
		}
		if seenMount[mount] {
			continue
		}
		if stat.Blocks == 0 {
			continue
		}

		bsize := uint64(stat.Bsize)
		size := stat.Blocks * bsize
		used := (stat.Blocks - stat.Bfree) * bsize

		if size < diskMinSize {
			continue
		}

		seenMount[mount] = true

		disks = append(disks, &metric.Mount{
			Mount: mount,
			Used:  utils.ToFixed(float64(used)/float64(size)*100, 2),
			Size:  size,
		})
	}

	return
}
