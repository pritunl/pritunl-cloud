package guestfs

import (
	"regexp"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

var (
	structStartRe = regexp.MustCompile(`^\[\d+\] = \{$`)
)

func parseLines(output string) (lines []string) {
	lines = []string{}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}

	return
}

func parseValue(output string) string {
	return strings.TrimSpace(output)
}

func parsePairs(output string) (keys []string, pairs map[string]string) {
	keys = []string{}
	pairs = map[string]string{}

	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}

		key, val, found := strings.Cut(line, ":")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		if _, exists := pairs[key]; !exists {
			keys = append(keys, key)
		}
		pairs[key] = strings.TrimSpace(val)
	}

	return
}

func parseStructs(output string) (items []map[string]string, err error) {
	items = []map[string]string{}

	var item map[string]string
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if structStartRe.MatchString(trimmed) {
			if item != nil {
				err = &errortypes.ParseError{
					errors.New("guestfs: Unterminated structure in output"),
				}
				return
			}
			item = map[string]string{}
			continue
		}

		if trimmed == "}" {
			if item == nil {
				err = &errortypes.ParseError{
					errors.New("guestfs: Unexpected structure end in output"),
				}
				return
			}
			items = append(items, item)
			item = nil
			continue
		}

		if item == nil {
			err = &errortypes.ParseError{
				errors.Newf("guestfs: Unexpected output line '%s'", trimmed),
			}
			return
		}

		key, val, found := strings.Cut(trimmed, ":")
		if !found {
			err = &errortypes.ParseError{
				errors.Newf("guestfs: Invalid structure line '%s'", trimmed),
			}
			return
		}

		item[strings.TrimSpace(key)] = strings.TrimSpace(val)
	}

	if item != nil {
		err = &errortypes.ParseError{
			errors.New("guestfs: Unterminated structure in output"),
		}
		return
	}

	return
}

func structInt(item map[string]string, key string) (num int64, err error) {
	val, ok := item[key]
	if !ok {
		err = &errortypes.ParseError{
			errors.Newf("guestfs: Missing field %s in output", key),
		}
		return
	}

	num, err = parseInt(val)
	if err != nil {
		return
	}

	return
}
