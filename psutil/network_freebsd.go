package psutil

import (
	"encoding/binary"
	"strings"
	"unsafe"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"golang.org/x/net/route"
	"golang.org/x/sys/unix"
)

type freebsdIfData struct {
	Type       uint8
	Physical   uint8
	Addrlen    uint8
	Hdrlen     uint8
	LinkState  uint8
	Vhid       uint8
	Datalen    uint16
	Mtu        uint32
	Metric     uint32
	Baudrate   uint64
	Ipackets   uint64
	Ierrors    uint64
	Opackets   uint64
	Oerrors    uint64
	Collisions uint64
	Ibytes     uint64
	Obytes     uint64
	Imcasts    uint64
	Omcasts    uint64
	Iqdrops    uint64
	Oqdrops    uint64
	Noproto    uint64
	Hwassist   uint64
	Epoch      int64
	Lastchange [16]byte
}

const (
	ifIndexOff = int(unsafe.Offsetof(unix.IfMsghdr{}.Index))
	ifDataOff  = int(unsafe.Offsetof(unix.IfMsghdr{}.Data))
	ifMinLen   = ifDataOff + int(unsafe.Sizeof(freebsdIfData{}))

	ifIpktsOff   = ifDataOff + int(unsafe.Offsetof(freebsdIfData{}.Ipackets))
	ifIerrOff    = ifDataOff + int(unsafe.Offsetof(freebsdIfData{}.Ierrors))
	ifOpktsOff   = ifDataOff + int(unsafe.Offsetof(freebsdIfData{}.Opackets))
	ifOerrOff    = ifDataOff + int(unsafe.Offsetof(freebsdIfData{}.Oerrors))
	ifIbytesOff  = ifDataOff + int(unsafe.Offsetof(freebsdIfData{}.Ibytes))
	ifObytesOff  = ifDataOff + int(unsafe.Offsetof(freebsdIfData{}.Obytes))
	ifIqdropsOff = ifDataOff + int(unsafe.Offsetof(freebsdIfData{}.Iqdrops))
	ifOqdropsOff = ifDataOff + int(unsafe.Offsetof(freebsdIfData{}.Oqdrops))
)

var networkIgnorePrefixes = []string{
	"lo",
	"bridge",
	"epair",
	"tap",
	"vm-",
	"pflog",
	"pfsync",
}

func networkIgnore(name string, flags int) bool {
	if flags&unix.IFF_LOOPBACK != 0 {
		return true
	}
	for _, prefix := range networkIgnorePrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func networkList(skipVirt bool) (stats []*networkStat, err error) {
	rib, err := route.FetchRIB(unix.AF_UNSPEC, route.RIBTypeInterface, 0)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to fetch interface list"),
		}
		return
	}

	msgs, err := route.ParseRIB(route.RIBTypeInterface, rib)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to parse interface list"),
		}
		return
	}

	names := map[int]string{}
	flags := map[int]int{}
	for _, msg := range msgs {
		ifm, ok := msg.(*route.InterfaceMessage)
		if !ok {
			continue
		}
		names[ifm.Index] = ifm.Name
		flags[ifm.Index] = ifm.Flags
	}

	for off := 0; off+4 <= len(rib); {
		msgLen := int(binary.NativeEndian.Uint16(rib[off : off+2]))
		if msgLen < 4 || off+msgLen > len(rib) {
			break
		}

		if rib[off+3] == unix.RTM_IFINFO && msgLen >= ifMinLen {
			idx := int(binary.NativeEndian.Uint16(
				rib[off+ifIndexOff : off+ifIndexOff+2]))
			name := names[idx]

			if name != "" && !networkIgnore(name, flags[idx]) &&
				!(skipVirt && len(name) == 14) {

				u64 := func(rel int) uint64 {
					o := off + rel
					return binary.NativeEndian.Uint64(rib[o : o+8])
				}

				stats = append(stats, &networkStat{
					Name:        name,
					BytesRecv:   u64(ifIbytesOff),
					BytesSent:   u64(ifObytesOff),
					PacketsRecv: u64(ifIpktsOff),
					PacketsSent: u64(ifOpktsOff),
					ErrorsRecv:  u64(ifIerrOff),
					ErrorsSent:  u64(ifOerrOff),
					DropsRecv:   u64(ifIqdropsOff),
					DropsSent:   u64(ifOqdropsOff),
				})
			}
		}

		off += msgLen
	}

	return
}
