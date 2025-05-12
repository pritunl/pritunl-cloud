package utils

import (
	"io/ioutil"
	"os/exec"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
)

var isSystemd *bool

var safeChars = set.NewSet(
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
	'+',
	'=',
	'_',
	'/',
	',',
	'.',
	'~',
	'@',
	'#',
	'!',
	'&',
	' ',
)

var pathSafeChars = set.NewSet(
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
	'+',
	'=',
	'_',
	',',
	'.',
	':',
	'~',
	'@',
	'!',
)

func FilterStr(s string, n int) string {
	if len(s) == 0 {
		return ""
	}

	if len(s) > n {
		s = s[:n]
	}

	ns := ""
	for _, c := range s {
		if safeChars.Contains(c) {
			ns += string(c)
		}
	}

	return ns
}

func FilterPath(s string, n int) string {
	if len(s) == 0 {
		return ""
	}

	if len(s) > n {
		s = s[:n]
	}

	ns := ""
	for _, c := range s {
		if pathSafeChars.Contains(c) {
			ns += string(c)
		}
	}

	return ns
}

func SinceAbs(t time.Time) (s time.Duration) {
	s = time.Since(t)
	if s < 0 {
		s = s * -1
	}
	return
}

func PointerBool(x bool) *bool {
	return &x
}

func PointerInt(x int) *int {
	return &x
}

func PointerString(x string) *string {
	return &x
}

func Int8Str(arr []int8) string {
	b := make([]byte, 0, len(arr))
	for _, v := range arr {
		if v == 0x00 {
			break
		}
		b = append(b, byte(v))
	}
	return string(b)
}

func HasPreSuf(src, pre, suf string) bool {
	return strings.HasPrefix(src, pre) && strings.HasSuffix(src, suf)
}

func IsSystemd() bool {
	if isSystemd != nil {
		return *isSystemd
	}

	data, err := ioutil.ReadFile("/proc/1/cmdline")
	if err == nil {
		parts := strings.Split(string(data), "\x00")
		if len(parts) > 0 && strings.Contains(
			strings.ToLower(parts[0]), "systemd") {

			isSysd := true
			isSystemd = &isSysd
			return true
		}
	}

	data, err = ioutil.ReadFile("/proc/1/comm")
	if err == nil {
		if strings.Contains(strings.ToLower(string(data)), "systemd") {
			isSysd := true
			isSystemd = &isSysd
			return true
		}
	}

	cmd := exec.Command("ps", "-p", "1", "-o", "comm=")
	output, err := cmd.Output()
	if err == nil {
		if strings.Contains(strings.ToLower(string(output)), "systemd") {
			isSysd := true
			isSystemd = &isSysd
			return true
		}
	}

	isSysd := false
	isSystemd = &isSysd
	return false
}

func CompareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func CompareStringSlicesUnsorted(a, b []string) bool {
	aSet := set.NewSet()
	for _, val := range a {
		aSet.Add(val)
	}

	bSet := set.NewSet()
	for _, val := range b {
		bSet.Add(val)
	}

	return aSet.IsEqual(bSet)
}
