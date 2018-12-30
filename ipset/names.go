package ipset

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Names struct {
	Namespace string
	Sets      set.Set
}

func (n *Names) Apply(curNames *Names) (err error) {
	if curNames != nil {
		for nameInf := range curNames.Sets.Iter() {
			name := nameInf.(string)

			if !n.Sets.Contains(name) {
				if n.Namespace == "0" {
					_, err = utils.ExecCombinedOutputLogged(
						[]string{"not exist"},
						"ipset", "destroy",
						name,
					)
					if err != nil {
						return
					}
				} else {
					_, err = utils.ExecCombinedOutputLogged(
						[]string{"not exist"},
						"ip", "netns", "exec", n.Namespace,
						"ipset", "destroy",
						name,
					)
					if err != nil {
						return
					}
				}

			}

		}
	}

	return
}
