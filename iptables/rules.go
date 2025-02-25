package iptables

type Rules struct {
	Namespace string
	Interface string
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
}
