package resource

import (
	"sort"

	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
)

func Hypervisors(db *database.Database) (nodes []*node.Node, err error) {
	all, err := node.GetAll(db)
	if err != nil {
		return
	}

	nodes = []*node.Node{}
	for _, nde := range all {
		if nde.IsHypervisor() {
			nodes = append(nodes, nde)
		}
	}

	sort.SliceStable(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})

	return
}
