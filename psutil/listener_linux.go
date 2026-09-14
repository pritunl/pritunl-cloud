package psutil

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"golang.org/x/sys/unix"
)

const (
	sockDiagByFamily = 20
	inetDiagReqLen   = 56
	inetDiagMsgLen   = 72

	tcpClose  = 7
	tcpListen = 10
)

type sockDiagQuery struct {
	Protocol string
	Family   uint8
	Proto    uint8
	States   uint32
}

var sockDiagQueries = []sockDiagQuery{
	{"tcp", syscall.AF_INET, syscall.IPPROTO_TCP, 1 << tcpListen},
	{"tcp", syscall.AF_INET6, syscall.IPPROTO_TCP, 1 << tcpListen},
	{"udp", syscall.AF_INET, syscall.IPPROTO_UDP, 1 << tcpClose},
	{"udp", syscall.AF_INET6, syscall.IPPROTO_UDP, 1 << tcpClose},
}

type procNetFile struct {
	Protocol string
	Path     string
	State    string
}

var procNetFiles = []procNetFile{
	{"tcp", "/proc/net/tcp", "0A"},
	{"tcp", "/proc/net/tcp6", "0A"},
	{"udp", "/proc/net/udp", "07"},
	{"udp", "/proc/net/udp6", "07"},
}

type listenerSocket struct {
	Protocol string
	Address  string
	Port     uint16
	Inode    uint64
}

func listenerList() (listeners []*Listener, err error) {
	sockets := []*listenerSocket{}

	for _, query := range sockDiagQueries {
		items, e := sockDiag(query)
		if e != nil {
			if errors.IsError(e, syscall.EPERM) ||
				errors.IsError(e, syscall.EACCES) ||
				errors.IsError(e, syscall.ENOENT) ||
				errors.IsError(e, syscall.EPROTONOSUPPORT) {

				sockets, err = listenerSocketsProc()
				if err != nil {
					return
				}

				listeners = resolveListeners(sockets)
				return
			}

			err = e
			return
		}

		sockets = append(sockets, items...)
	}

	listeners = resolveListeners(sockets)
	return
}

func resolveListeners(sockets []*listenerSocket) (listeners []*Listener) {
	inodePids := socketInodePids()
	names := map[int]string{}
	seen := map[string]bool{}
	listeners = []*Listener{}

	for _, sock := range sockets {
		pids := inodePids[sock.Inode]
		if len(pids) == 0 {
			pids = []int{0}
		}

		for _, pid := range pids {
			key := sock.Protocol + "|" + sock.Address + "|" +
				strconv.Itoa(int(sock.Port)) + "|" + strconv.Itoa(pid)
			if seen[key] {
				continue
			}
			seen[key] = true

			name := ""
			if pid != 0 {
				var ok bool
				name, ok = names[pid]
				if !ok {
					name = processName(strconv.Itoa(pid))
					names[pid] = name
				}
			}

			listeners = append(listeners, &Listener{
				Protocol: sock.Protocol,
				Address:  sock.Address,
				Port:     sock.Port,
				Pid:      pid,
				Name:     name,
			})
		}
	}

	return
}

func socketInodePids() (inodePids map[uint64][]int) {
	inodePids = map[uint64][]int{}

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() || !isPid(entry.Name()) {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		fdDir := filepath.Join("/proc", entry.Name(), "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}

		pidSeen := map[uint64]bool{}
		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err != nil {
				continue
			}

			if !strings.HasPrefix(link, "socket:[") ||
				!strings.HasSuffix(link, "]") {

				continue
			}

			inode, err := strconv.ParseUint(
				link[len("socket:["):len(link)-1], 10, 64)
			if err != nil || pidSeen[inode] {
				continue
			}

			pidSeen[inode] = true
			inodePids[inode] = append(inodePids[inode], pid)
		}
	}

	return
}

func sockDiag(query sockDiagQuery) (sockets []*listenerSocket, err error) {
	fd, err := syscall.Socket(syscall.AF_NETLINK,
		syscall.SOCK_RAW|syscall.SOCK_CLOEXEC, unix.NETLINK_SOCK_DIAG)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to open netlink socket"),
		}
		return
	}
	defer syscall.Close(fd)

	err = syscall.Bind(fd, &syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
	})
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to bind netlink socket"),
		}
		return
	}

	req := make([]byte, syscall.NLMSG_HDRLEN+inetDiagReqLen)
	hdr := (*syscall.NlMsghdr)(unsafe.Pointer(&req[0]))
	hdr.Len = uint32(len(req))
	hdr.Type = sockDiagByFamily
	hdr.Flags = syscall.NLM_F_REQUEST | syscall.NLM_F_DUMP
	hdr.Seq = 1

	body := req[syscall.NLMSG_HDRLEN:]
	body[0] = query.Family
	body[1] = query.Proto
	binary.NativeEndian.PutUint32(body[4:8], query.States)

	err = syscall.Sendto(fd, req, 0, &syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
	})
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to send netlink request"),
		}
		return
	}

	buf := make([]byte, 64*1024)
	sockets = []*listenerSocket{}

	for {
		n, _, e := syscall.Recvfrom(fd, buf, 0)
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "psutil: Failed to read netlink response"),
			}
			return
		}

		msgs, e := syscall.ParseNetlinkMessage(buf[:n])
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "psutil: Failed to parse netlink response"),
			}
			return
		}

		for _, msg := range msgs {
			switch msg.Header.Type {
			case syscall.NLMSG_DONE:
				return
			case syscall.NLMSG_ERROR:
				if len(msg.Data) < 4 {
					err = &errortypes.ReadError{
						errors.New("psutil: Netlink error response"),
					}
					return
				}

				errno := int32(binary.NativeEndian.Uint32(msg.Data[:4]))
				err = &errortypes.ReadError{
					errors.Wrap(syscall.Errno(-errno),
						"psutil: Netlink socket diag error"),
				}
				return
			}

			if msg.Header.Type != sockDiagByFamily ||
				len(msg.Data) < inetDiagMsgLen {

				continue
			}

			data := msg.Data

			var addr net.IP
			if data[0] == syscall.AF_INET {
				addr = net.IP(append([]byte(nil), data[8:12]...))
			} else {
				addr = net.IP(append([]byte(nil), data[8:24]...))
			}

			sockets = append(sockets, &listenerSocket{
				Protocol: query.Protocol,
				Address:  addr.String(),
				Port:     binary.BigEndian.Uint16(data[4:6]),
				Inode:    uint64(binary.NativeEndian.Uint32(data[68:72])),
			})
		}
	}
}

func listenerSocketsProc() (sockets []*listenerSocket, err error) {
	sockets = []*listenerSocket{}

	for _, netFile := range procNetFiles {
		items, e := parseProcNet(netFile)
		if e != nil {
			err = e
			return
		}

		sockets = append(sockets, items...)
	}

	return
}

func parseProcNet(netFile procNetFile) (sockets []*listenerSocket,
	err error) {

	sockets = []*listenerSocket{}

	file, err := os.Open(netFile.Path)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			return
		}

		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to open "+netFile.Path),
		}
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 10 || fields[0] == "sl" {
			continue
		}

		if !strings.EqualFold(fields[3], netFile.State) {
			continue
		}

		addr, port, e := parseProcNetAddr(fields[1])
		if e != nil {
			continue
		}

		inode, e := strconv.ParseUint(fields[9], 10, 64)
		if e != nil {
			inode = 0
		}

		sockets = append(sockets, &listenerSocket{
			Protocol: netFile.Protocol,
			Address:  addr,
			Port:     port,
			Inode:    inode,
		})
	}

	err = scanner.Err()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to read "+netFile.Path),
		}
		return
	}

	return
}

func parseProcNetAddr(field string) (addr string, port uint16, err error) {
	sep := strings.LastIndex(field, ":")
	if sep < 0 {
		err = &errortypes.ParseError{
			errors.New("psutil: Invalid proc net address"),
		}
		return
	}

	portNum, err := strconv.ParseUint(field[sep+1:], 16, 16)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "psutil: Invalid proc net port"),
		}
		return
	}
	port = uint16(portNum)

	raw, err := hex.DecodeString(field[:sep])
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "psutil: Invalid proc net address"),
		}
		return
	}

	if len(raw) != 4 && len(raw) != 16 {
		err = &errortypes.ParseError{
			errors.New("psutil: Invalid proc net address length"),
		}
		return
	}

	ip := make(net.IP, len(raw))
	for i := 0; i < len(raw); i += 4 {
		word := binary.NativeEndian.Uint32(raw[i : i+4])
		binary.BigEndian.PutUint32(ip[i:i+4], word)
	}

	addr = ip.String()

	return
}
