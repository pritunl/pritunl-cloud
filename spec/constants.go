package spec

import (
	"regexp"
)

var resourcesRe = regexp.MustCompile("(?s)```yaml(.*?)```")

type Base struct {
	Kind string `yaml:"kind"`
}
