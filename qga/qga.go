package qga

import (
	"bytes"
	"encoding/json"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"net"
	"strings"
	"time"
)

type Command struct {
	Execute string `json:"execute"`
}

type Address struct {
	Type    string `json:"ip-address-type"`
	Address string `json:"ip-address"`
	Prefix  int    `json:"prefix"`
}

type Interface struct {
	Name       string     `json:"name"`
	MacAddress string     `json:"hardware-address"`
	Addresses  []*Address `json:"ip-addresses"`
}

type Interfaces struct {
	Interfaces []*Interface `json:"return"`
}

func (i *Interfaces) GetAddr(macAddr string) (guestAddr, guestAddr6 string) {
	macAddr = strings.ToLower(macAddr)

	if i.Interfaces != nil {
		for _, iface := range i.Interfaces {
			if strings.ToLower(iface.MacAddress) != macAddr {
				continue
			}

			if iface.Addresses != nil {
				for _, addr := range iface.Addresses {
					if addr.Type == "ipv4" && guestAddr == "" {
						guestAddr = addr.Address
					} else if addr.Type == "ipv6" && guestAddr6 == "" {
						ipAddr := strings.ToLower(addr.Address)
						if !strings.HasPrefix(ipAddr, "fe") {
							guestAddr6 = strings.ToLower(addr.Address)
						}
					}
				}
			}

			break
		}
	}

	return
}

func GetInterfaces(sockPath string) (ifaces *Interfaces, err error) {
	conn, err := net.DialTimeout(
		"unix",
		sockPath,
		1*time.Second,
	)
	if err != nil {
		err = &errortypes.ConnectionError{
			errors.Wrap(err, "qga: Failed to connect to guest agent"),
		}
		return
	}
	defer conn.Close()

	err = conn.SetDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		return
	}

	cmd := &Command{
		Execute: "guest-network-get-interfaces",
	}

	cmdByte, err := json.Marshal(cmd)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "qga: Failed to parse guest agent command"),
		}
		return
	}

	_, err = conn.Write(cmdByte)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "qga: Failed to write to guest agent"),
		}
		return
	}

	buff := make([]byte, 100000)
	_, err = conn.Read(buff)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "qga: Failed to read from guest agent"),
		}
		return
	}

	respByt := bytes.Trim(buff, "\x00")
	respByt = bytes.TrimSpace(respByt)

	ifaces = &Interfaces{}
	err = json.Unmarshal(respByt, ifaces)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "qga: Failed to parse guest agent response"),
		}
		return
	}

	return
}
