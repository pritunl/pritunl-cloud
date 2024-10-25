package spec

import (
	"regexp"
)

var resourcesRe = regexp.MustCompile("(?s)```yaml(.*?)```")
