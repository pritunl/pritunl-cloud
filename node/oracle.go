package node

type NodeOracleAuthProvider struct {
	nde *Node
}

func (n *NodeOracleAuthProvider) OracleUser() string {
	return n.nde.OracleUser
}

func (n *NodeOracleAuthProvider) OracleTenancy() string {
	return n.nde.OracleTenancy
}

func (n *NodeOracleAuthProvider) OraclePrivateKey() string {
	return n.nde.OraclePrivateKey
}
