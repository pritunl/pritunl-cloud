package settings

var Local *local

type local struct {
	BridgeName  string
}

func init() {
	Local = &local{}
}
