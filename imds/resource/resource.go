package resource

import (
	"github.com/pritunl/pritunl-cloud/imds/server/config"
)

func Query(resrc, name, key string) (val string, err error) {
	var resrcInf interface{}

	switch resrc {
	case "instance":
		resrcInf = config.Config.Instance
		key = name
		break
	case "vpc":
		resrcInf = config.Config.Vpc
		key = name
		break
	case "subnet":
		resrcInf = config.Config.Subnet
		key = name
		break
	case "secrets":
		for _, secr := range config.Config.Secrets {
			if secr.Name == name {
				resrcInf = secr
				break
			}
		}
		break
	case "certificates":
		for _, cert := range config.Config.Certificates {
			if cert.Name == name {
				resrcInf = cert
				break
			}
		}
		break
	default:
		return
	}

	if resrcInf == nil {
		return
	}

	val, err = selector(resrcInf, key)
	if err != nil {
		return
	}

	return
}
