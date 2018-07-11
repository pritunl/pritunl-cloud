package settings

var Local *local

type local struct {
	BridgeName  string
	NoLocalAuth bool
}

func init() {
	Local = &local{}
}
