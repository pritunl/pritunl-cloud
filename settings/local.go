package settings

var Local *local

type local struct {
	BridgeName  string
	AppId       string
	Facets      []string
	NoLocalAuth bool
}

func init() {
	Local = &local{}
}
