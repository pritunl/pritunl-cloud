package iptables

type Rules struct {
	Namespace        string
	Interface        string
	Nat              bool
	NatAddr          string
	NatPubAddr       string
	Nat6             bool
	NatAddr6         string
	NatPubAddr6      string
	OracleNat        bool
	OracleNatAddr    string
	OracleNatPubAddr string
	Header           [][]string
	Header6          [][]string
	SourceDestCheck  [][]string
	SourceDestCheck6 [][]string
	Ingress          [][]string
	Ingress6         [][]string
	Maps             [][]string
	Maps6            [][]string
	Holds            [][]string
	Holds6           [][]string
}
