package settings

var Local *local

type local struct {
	BridgeName  string
	PublicAddr  string
	PublicAddr6 string
}

func init() {
	Local = &local{
		PublicAddr: "",
	}
}
