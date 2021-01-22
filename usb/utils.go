package usb

import (
	"regexp"
	"strings"
)

var (
	reg = regexp.MustCompile("[^a-z0-9]+")
)

func FilterId(deviceId string) string {
	deviceId = strings.ToLower(deviceId)
	deviceId = reg.ReplaceAllString(deviceId, "")
	if len(deviceId) != 4 {
		return ""
	}
	return deviceId
}

func FilterAddr(addr string) string {
	addr = strings.ToLower(addr)
	addr = reg.ReplaceAllString(addr, "")
	if len(addr) != 3 {
		return ""
	}
	return addr
}
