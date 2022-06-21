package constants

import "github.com/pritunl/pritunl-cloud/utils"

var (
	Target string
)

const (
	Version = "1.0.2401.42"
	Rpm     = "rpm"
	Apt     = "apt"
)

func Init() (err error) {
	exists, err := utils.ExistsDir("/etc/apt/sources.list.d")
	if err != nil {
		return
	}

	if exists {
		Target = Apt
	} else {
		Target = Rpm
	}

	return
}
