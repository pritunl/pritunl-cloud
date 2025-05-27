package virtiofs

import (
	"github.com/pritunl/pritunl-cloud/utils"
)

const (
	Libexec = "/usr/libexec/virtiofsd"
	System  = "/usr/bin/virtiofsd"
)

func GetVirtioFsdPath() string {
	exists, _ := utils.Exists(System)
	if exists {
		return System
	}
	return Libexec
}
