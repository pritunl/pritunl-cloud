package node

type NodeOracleAuthProvider struct {
	nde *Node
}

func (n *NodeOracleAuthProvider) OracleUser() string {
	return n.nde.OracleUser
}

func (n *NodeOracleAuthProvider) OraclePrivateKey() string {
	return n.nde.OraclePrivateKey
}
