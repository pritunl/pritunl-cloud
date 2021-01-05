package pci

import (
	"regexp"
)

var (
	reg = regexp.MustCompile(
		"[a-fA-F0-9][a-fA-F0-9]:[a-fA-F0-9][a-fA-F0-9].[0-9]")
)

func CheckSlot(slot string) bool {
	return reg.MatchString(slot)
}
