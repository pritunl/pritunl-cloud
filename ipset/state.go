package ipset

import (
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/firewall"
)

type State struct {
	Namespaces map[string]*Sets
}

func (s *State) AddIngress(namespace string, ingress []*firewall.Rule) {
	sets := s.Namespaces[namespace]
	if sets == nil {
		sets = &Sets{
			Namespace: namespace,
			Sets:      map[string]set.Set{},
		}
		s.Namespaces[namespace] = sets
	}

	for _, rule := range ingress {
		name := rule.SetName(false)
		name6 := rule.SetName(true)

		if name == "" || name6 == "" || rule.Protocol == firewall.Multicast ||
			rule.Protocol == firewall.Broadcast {

			continue
		}

		for _, sourceIp := range rule.SourceIps {
			if sourceIp == "0.0.0.0/0" || sourceIp == "::/0" {
				continue
			}

			ruleName := ""
			ipv6 := strings.Contains(sourceIp, ":")
			if ipv6 {
				sourceIp = strings.Replace(sourceIp, "/128", "", 1)
				ruleName = name6
			} else {
				sourceIp = strings.Replace(sourceIp, "/32", "", 1)
				ruleName = name
			}

			ruleSet := sets.Sets[ruleName]
			if ruleSet == nil {
				ruleSet = set.NewSet()
				sets.Sets[ruleName] = ruleSet
			}

			ruleSet.Add(sourceIp)
		}
	}
}

func (s *State) AddSourceDestCheck(namespace, addr6 string) {
	sets := s.Namespaces[namespace]
	if sets == nil {
		sets = &Sets{
			Namespace: namespace,
			Sets:      map[string]set.Set{},
		}
		s.Namespaces[namespace] = sets
	}

	sdcSet := set.NewSet()
	if addr6 != "" {
		sdcSet.Add(addr6)
	}
	sdcSet.Add("fe80::/10")
	sets.Sets["pr6_sdc"] = sdcSet
}

func (s *State) AddMember(namespace string, ruleName, member string) {
	if strings.HasPrefix(ruleName, "prx") {
		return
	}

	sets := s.Namespaces[namespace]
	if sets == nil {
		sets = &Sets{
			Namespace: namespace,
			Sets:      map[string]set.Set{},
		}
		s.Namespaces[namespace] = sets
	}

	ruleSet := sets.Sets[ruleName]
	if ruleSet == nil {
		ruleSet = set.NewSet()
		sets.Sets[ruleName] = ruleSet
	}

	ruleSet.Add(member)
}

type NamesState struct {
	Namespaces map[string]*Names
}

func (n *NamesState) AddIngress(namespace string, ingress []*firewall.Rule) {
	sets := n.Namespaces[namespace]
	if sets == nil {
		sets = &Names{
			Namespace: namespace,
			Sets:      set.NewSet(),
		}
		n.Namespaces[namespace] = sets
	}

	for _, rule := range ingress {
		name := rule.SetName(false)
		name6 := rule.SetName(true)

		if name == "" || name6 == "" || rule.Protocol == firewall.Multicast ||
			rule.Protocol == firewall.Broadcast {

			continue
		}

		for _, sourceIp := range rule.SourceIps {
			if sourceIp == "0.0.0.0/0" || sourceIp == "::/0" {
				continue
			}

			ipv6 := strings.Contains(sourceIp, ":")
			if ipv6 {
				sets.Sets.Add(name6)
			} else {
				sets.Sets.Add(name)
			}

		}
	}
}

func (n *NamesState) AddSourceDestCheck(namespace string) {
	sets := n.Namespaces[namespace]
	if sets == nil {
		sets = &Names{
			Namespace: namespace,
			Sets:      set.NewSet(),
		}
		n.Namespaces[namespace] = sets
	}

	sets.Sets.Add("pr6_sdc")
}

func (n *NamesState) AddName(namespace string, ruleName string) {
	if strings.HasPrefix(ruleName, "prx") {
		return
	}

	sets := n.Namespaces[namespace]
	if sets == nil {
		sets = &Names{
			Namespace: namespace,
			Sets:      set.NewSet(),
		}
		n.Namespaces[namespace] = sets
	}

	sets.Sets.Add(ruleName)
}
