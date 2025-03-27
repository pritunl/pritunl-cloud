package iptables

import (
	"github.com/pritunl/pritunl-cloud/ipvs"
)

type Rules struct {
	Namespace        string
	Interface        string
	Header           [][]string
	Header6          [][]string
	SourceDestCheck  [][]string
	SourceDestCheck6 [][]string
	Ingress          [][]string
	Ingress6         [][]string
	Nats             [][]string
	Nats6            [][]string
	Maps             [][]string
	Maps6            [][]string
	Holds            [][]string
	Holds6           [][]string
	Ipvs             *ipvs.State
}

type RulesDiff struct {
	HeaderDiff           bool
	Header6Diff          bool
	SourceDestCheckDiff  bool
	SourceDestCheck6Diff bool
	IngressDiff          bool
	Ingress6Diff         bool
	NatsDiff             bool
	Nats6Diff            bool
	MapsDiff             bool
	Maps6Diff            bool
	HoldsDiff            bool
	Holds6Diff           bool
}
