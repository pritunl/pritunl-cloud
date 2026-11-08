package statec

import (
	"fmt"
	"strings"
)

const debugKeysLimit = 12

func debugKeys(keys []string) string {
	if len(keys) > debugKeysLimit {
		return strings.Join(keys[:debugKeysLimit], ",") +
			fmt.Sprintf(",+%d", len(keys)-debugKeysLimit)
	}
	return strings.Join(keys, ",")
}
