package psutil

import (
	"encoding/binary"
	"regexp"
	"strconv"
	"unsafe"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"golang.org/x/sys/unix"
)

var diskIoNameReg = regexp.MustCompile(
	`^(da|ada|nvd|nda|vtbd|mmcsd)[0-9]+$`)

type devstatBintime struct {
	Sec  int64
	Frac uint64
}

type devstatHead struct {
	Sequence0    uint32
	Allocated    int32
	StartCount   uint32
	EndCount     uint32
	BusyFrom     devstatBintime
	DevLinks     uint64
	DeviceNumber uint32
	DeviceName   [16]byte
	UnitNumber   int32
	Bytes        [4]uint64
	Operations   [4]uint64
	Duration     [4]devstatBintime
	BusyTime     devstatBintime
}

const (
	devstatGenLen = 8

	devstatRead  = 1
	devstatWrite = 2

	dsHeadLen  = int(unsafe.Sizeof(devstatHead{}))
	dsNameOff  = int(unsafe.Offsetof(devstatHead{}.DeviceName))
	dsUnitOff  = int(unsafe.Offsetof(devstatHead{}.UnitNumber))
	dsBytesOff = int(unsafe.Offsetof(devstatHead{}.Bytes))
	dsOpsOff   = int(unsafe.Offsetof(devstatHead{}.Operations))
	dsDurOff   = int(unsafe.Offsetof(devstatHead{}.Duration))
	dsBusyOff  = int(unsafe.Offsetof(devstatHead{}.BusyTime))
)

func dsU64(rec []byte, off int) uint64 {
	return binary.NativeEndian.Uint64(rec[off : off+8])
}

func dsBintimeMs(rec []byte, off int) uint64 {
	sec := int64(binary.NativeEndian.Uint64(rec[off : off+8]))
	frac := binary.NativeEndian.Uint64(rec[off+8 : off+16])
	return uint64((float64(sec) + float64(frac)/float64(1<<64)) * 1000)
}

func diskIoList() (stats []*diskIoStat, err error) {
	buf, err := unix.SysctlRaw("kern.devstat.all")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to read devstat"),
		}
		return
	}

	numDevs, err := unix.SysctlUint32("kern.devstat.numdevs")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to read devstat numdevs"),
		}
		return
	}

	if numDevs == 0 || len(buf) <= devstatGenLen {
		return
	}

	body := buf[devstatGenLen:]
	recSize := len(body) / int(numDevs)
	if recSize < dsHeadLen || len(body)%int(numDevs) != 0 {
		return
	}

	for i := 0; i < int(numDevs); i++ {
		rec := body[i*recSize : (i+1)*recSize]

		name := unix.ByteSliceToString(rec[dsNameOff : dsNameOff+16])
		unit := int32(binary.NativeEndian.Uint32(
			rec[dsUnitOff : dsUnitOff+4]))
		full := name + strconv.FormatInt(int64(unit), 10)
		if !diskIoNameReg.MatchString(full) {
			continue
		}

		stats = append(stats, &diskIoStat{
			Name:       full,
			CountRead:  dsU64(rec, dsOpsOff+devstatRead*8),
			CountWrite: dsU64(rec, dsOpsOff+devstatWrite*8),
			BytesRead:  dsU64(rec, dsBytesOff+devstatRead*8),
			BytesWrite: dsU64(rec, dsBytesOff+devstatWrite*8),
			TimeRead:   dsBintimeMs(rec, dsDurOff+devstatRead*16),
			TimeWrite:  dsBintimeMs(rec, dsDurOff+devstatWrite*16),
			TimeIo:     dsBintimeMs(rec, dsBusyOff),
		})
	}

	return
}
