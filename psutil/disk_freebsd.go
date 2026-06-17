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

		if size < diskMinSize {
			continue
		}

		usedBlocks := stat.Blocks - stat.Bfree
		availBlocks := uint64(0)
		if stat.Bavail > 0 {
			availBlocks = uint64(stat.Bavail)
		}
		totalBlocks := usedBlocks + availBlocks

		usedPercent := 0.0
		if totalBlocks > 0 {
			usedPercent = utils.ToFixed(
				float64(usedBlocks)/float64(totalBlocks)*100, 2)
		}

		seenMount[mount] = true

		disks = append(disks, &metric.Mount{
			Mount: mount,
			Used:  usedPercent,
			Size:  size,
		})
	}

	return
}
