package image

import (
	"github.com/pritunl/pritunl-cloud/utils"
)

type ImagesSort []*Image

func (x ImagesSort) Len() int {
	return len(x)
}

func (x ImagesSort) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (x ImagesSort) Less(i, j int) bool {
	return utils.NaturalCompare(x[i].Name, x[j].Name) < 0
}

type CompletionsSort []*Completion

func (x CompletionsSort) Len() int {
	return len(x)
}

func (x CompletionsSort) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (x CompletionsSort) Less(i, j int) bool {
	return utils.NaturalCompare(x[i].Name, x[j].Name) < 0
}
