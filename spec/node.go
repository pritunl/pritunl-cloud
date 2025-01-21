package spec

import (
	"sort"

	"github.com/pritunl/pritunl-cloud/node"
)

type Nodes []*node.Node

func (n Nodes) Len() int {
	return len(n)
}

func (n Nodes) Less(i, j int) bool {
	return n[i].Usage() < n[j].Usage()
}

func (n Nodes) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n Nodes) Sort() {
	sort.Sort(n)
}
