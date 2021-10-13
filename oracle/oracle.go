package oracle

type AuthProvider interface {
	OracleUser() string
	OraclePrivateKey() string
}
