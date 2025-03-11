package oracle

type AuthProvider interface {
	OracleUser() string
	OracleTenancy() string
	OraclePrivateKey() string
}
