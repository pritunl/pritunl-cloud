package ipset

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/utils"
	"strings"
)

type Sets struct {
	Namespace string
	Sets      map[string]set.Set
}

func (s *Sets) Apply(curSets *Sets) (err error) {
	namesSet := set.NewSet()

	for name, ipSet := range s.Sets {
		namesSet.Add(name)

		var curIpSet set.Set
		if curSets != nil {
			curIpSet = curSets.Sets[name]
		}
		if curIpSet == nil {
			curIpSet = set.NewSet()
		}

		created := false

		for ipInf := range ipSet.Iter() {
			ip := ipInf.(string)

			if curIpSet.Contains(ip) {
				continue
			}

			if !created {
				family := "inet"
				if strings.HasPrefix(name, "pr6") {
					family = "inet6"
				}

				if s.Namespace == "0" {
					_, err = utils.ExecCombinedOutputLogged(
						[]string{"already exists"},
						"ipset", "create",
						name, "hash:net",
						"family", family,
					)
					if err != nil {
						return
					}
				} else {
					_, err = utils.ExecCombinedOutputLogged(
						[]string{"already exists"},
						"ip", "netns", "exec", s.Namespace,
						"ipset", "create",
						name, "hash:net",
						"family", family,
					)
					if err != nil {
						return
					}
				}
				created = true
			}

			if s.Namespace == "0" {
				_, err = utils.ExecCombinedOutputLogged(
					[]string{"already added"},
					"ipset", "add",
					name, ip,
				)
				if err != nil {
					return
				}
			} else {
				_, err = utils.ExecCombinedOutputLogged(
					[]string{"already added"},
					"ip", "netns", "exec", s.Namespace,
					"ipset", "add",
					name, ip,
				)
				if err != nil {
					return
				}
			}
		}

		delIpSet := curIpSet.Copy()
		delIpSet.Subtract(ipSet)

		for ipInf := range delIpSet.Iter() {
			ip := ipInf.(string)

			if s.Namespace == "0" {
				_, err = utils.ExecCombinedOutputLogged(
					[]string{
						"not exist",
						"not added",
					},
					"ipset", "del",
					name, ip,
				)
				if err != nil {
					return
				}
			} else {
				_, err = utils.ExecCombinedOutputLogged(
					[]string{
						"not exist",
						"not added",
					},
					"ip", "netns", "exec", s.Namespace,
					"ipset", "del",
					name, ip,
				)
				if err != nil {
					return
				}
			}
		}
	}

	return
}
