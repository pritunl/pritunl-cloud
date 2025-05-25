package utils

import (
	"strings"

	"github.com/dropbox/godropbox/container/set"
)

const nameSafeLimit = 128

var nameSafeChar = set.NewSet(
	'a',
	'b',
	'c',
	'd',
	'e',
	'f',
	'g',
	'h',
	'i',
	'j',
	'k',
	'l',
	'm',
	'n',
	'o',
	'p',
	'q',
	'r',
	's',
	't',
	'u',
	'v',
	'w',
	'x',
	'y',
	'z',
	'A',
	'B',
	'C',
	'D',
	'E',
	'F',
	'G',
	'H',
	'I',
	'J',
	'K',
	'L',
	'M',
	'N',
	'O',
	'P',
	'Q',
	'R',
	'S',
	'T',
	'U',
	'V',
	'W',
	'X',
	'Y',
	'Z',
	'0',
	'1',
	'2',
	'3',
	'4',
	'5',
	'6',
	'7',
	'8',
	'9',
	'-',
	'.',
)

var nameCmdSafeChar = set.NewSet(
	'a',
	'b',
	'c',
	'd',
	'e',
	'f',
	'g',
	'h',
	'i',
	'j',
	'k',
	'l',
	'm',
	'n',
	'o',
	'p',
	'q',
	'r',
	's',
	't',
	'u',
	'v',
	'w',
	'x',
	'y',
	'z',
	'A',
	'B',
	'C',
	'D',
	'E',
	'F',
	'G',
	'H',
	'I',
	'J',
	'K',
	'L',
	'M',
	'N',
	'O',
	'P',
	'Q',
	'R',
	'S',
	'T',
	'U',
	'V',
	'W',
	'X',
	'Y',
	'Z',
	'0',
	'1',
	'2',
	'3',
	'4',
	'5',
	'6',
	'7',
	'8',
	'9',
	'-',
	'.',
)

func FilterName(s string) string {
	if len(s) == 0 {
		return ""
	}

	if s == "self" {
		s = "invalid-name"
	}

	if len(s) > nameSafeLimit {
		s = s[:nameSafeLimit]
	}

	ns := ""
	for _, c := range s {
		if nameSafeChar.Contains(c) {
			ns += string(c)
		}
	}

	return ns
}

func FilterNameCmd(s string) string {
	if len(s) == 0 {
		return ""
	}

	if s == "self" {
		s = "invalid-name"
	}

	if len(s) > nameSafeLimit {
		s = s[:nameSafeLimit]
	}

	ns := ""
	for _, c := range s {
		if nameCmdSafeChar.Contains(c) {
			ns += string(c)
		}
	}

	return strings.ToLower(ns)
}

func FilterDomain(s string) string {
	return FilterName(strings.ToLower(s))
}
