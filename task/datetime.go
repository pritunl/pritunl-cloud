package task

import (
	"encoding/binary"
	"net"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/sirupsen/logrus"
)

const (
	ntpServer      = "time.cloudflare.com:123"
	ntpTimeout     = 10 * time.Second
	ntpAttempts    = 3
	ntpEpochOffset = 2208988800
	maxTimeSkew    = 5 * time.Second
)

var timeSync = &Task{
	Name:    "time_sync",
	Version: 1,
	Hours: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
	Minutes:    []int{0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55},
	Handler:    timeSyncHandler,
	Local:      true,
	RunOnStart: true,
}

func ntpTimestamp(sec, frac uint32) time.Time {
	return time.Unix(
		int64(sec)-ntpEpochOffset,
		(int64(frac)*1e9)>>32,
	)
}

func ntpOffset() (offset time.Duration, err error) {
	conn, err := net.DialTimeout("udp", ntpServer, ntpTimeout)
	if err != nil {
		err = &errortypes.ConnectionError{
			errors.Wrap(err, "task: Failed to connect to ntp server"),
		}
		return
	}
	defer conn.Close()

	err = conn.SetDeadline(time.Now().Add(ntpTimeout))
	if err != nil {
		err = &errortypes.ConnectionError{
			errors.Wrap(err, "task: Failed to set ntp connection deadline"),
		}
		return
	}

	req := make([]byte, 48)
	req[0] = 0x23

	sent := time.Now()
	_, err = conn.Write(req)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "task: Failed to send ntp request"),
		}
		return
	}

	resp := make([]byte, 48)
	n, err := conn.Read(resp)
	recv := time.Now()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "task: Failed to read ntp response"),
		}
		return
	}

	if n < 48 {
		err = &errortypes.ParseError{
			errors.New("task: Invalid ntp response length"),
		}
		return
	}

	mode := resp[0] & 0x7
	stratum := resp[1]
	if mode != 4 || stratum == 0 {
		err = &errortypes.ParseError{
			errors.New("task: Invalid ntp response"),
		}
		return
	}

	serverRecv := ntpTimestamp(
		binary.BigEndian.Uint32(resp[32:36]),
		binary.BigEndian.Uint32(resp[36:40]),
	)
	serverSent := ntpTimestamp(
		binary.BigEndian.Uint32(resp[40:44]),
		binary.BigEndian.Uint32(resp[44:48]),
	)

	offset = (serverRecv.Sub(sent) + serverSent.Sub(recv)) / 2

	return
}

func timeSyncHandler(db *database.Database) (err error) {
	var offset time.Duration

	for i := 0; i < ntpAttempts; i++ {
		offset, err = ntpOffset()
		if err == nil {
			break
		}
	}
	if err != nil {
		return
	}

	skew := offset
	if skew < 0 {
		skew = -skew
	}

	if skew > maxTimeSkew {
		logrus.WithFields(logrus.Fields{
			"time_offset": offset.Seconds(),
			"ntp_server":  ntpServer,
		}).Error("task: Node time skew detected, cluster unstable")
	}

	return
}

func init() {
	register(timeSync)
}
