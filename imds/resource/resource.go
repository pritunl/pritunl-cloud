package resource

import (
	"github.com/pritunl/pritunl-cloud/imds/server/config"
)

func Query(resrc string, keys ...string) (val string, err error) {
	var resrcInf interface{}

	key := ""

	switch resrc {
	case "instance":
		resrcInf = config.Config.Instance
		break
	case "vpc":
		resrcInf = config.Config.Vpc
		break
	case "subnet":
		resrcInf = config.Config.Subnet
		break
	case "secrets":
		if len(keys) != 2 {
			break
		}
		key = keys[1]

		for _, secr := range config.Config.Secrets {
			if secr.Name == keys[0] {
				resrcInf = secr
				break
			}
		}
		break
	case "certificates":
		if len(keys) != 2 {
			break
		}
		key = keys[1]

		for _, cert := range config.Config.Certificates {
			if cert.Name == keys[0] {
				resrcInf = cert
				break
			}
		}
		break
	case "services":
		if len(keys) == 2 {
			key = keys[1]
		} else if len(keys) == 4 {
			key = keys[3]
		} else {
			break
		}

		for _, srvc := range config.Config.Services {
			if srvc.Name == keys[0] {
				if len(keys) == 4 {
					if keys[1] == "units" {
						for _, unit := range srvc.Units {
							if unit.Name == keys[2] {
								resrcInf = unit
								break
							}
						}
					}
				} else {
					resrcInf = srvc
				}
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
