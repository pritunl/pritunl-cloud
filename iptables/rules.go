package iptables

type Rules struct {
	Namespace            string
	Interface            string
	Header               [][]string
	HeaderDiff           bool
	Header6              [][]string
	Header6Diff          bool
	SourceDestCheck      [][]string
	SourceDestCheckDiff  bool
	SourceDestCheck6     [][]string
	SourceDestCheck6Diff bool
	Ingress              [][]string
	IngressDiff          bool
	Ingress6             [][]string
	Ingress6Diff         bool
	Nats                 [][]string
	NatsDiff             bool
	Nats6                [][]string
	Nats6Diff            bool
	Maps                 [][]string
	MapsDiff             bool
	Maps6                [][]string
	Maps6Diff            bool
	Holds                [][]string
	HoldsDiff            bool
	Holds6               [][]string
	Holds6Diff           bool
}
