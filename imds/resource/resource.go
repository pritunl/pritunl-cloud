package resource

import (
	"github.com/pritunl/pritunl-cloud/imds/server/config"
)

func Query(resrc string, keys ...string) (val string, err error) {
	var resrcInf interface{}

	key := ""

	switch resrc {
	case "instance":
		if len(keys) == 1 {
			key = keys[0]
			resrcInf = config.Config.Instance
		}
		break
	case "vpc":
		if len(keys) == 1 {
			key = keys[0]
			resrcInf = config.Config.Vpc
		}
		break
	case "subnet":
		if len(keys) == 1 {
			key = keys[0]
			resrcInf = config.Config.Subnet
		}
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
	case "pods":
		if len(keys) == 2 {
			key = keys[1]
		} else if len(keys) == 4 {
			key = keys[3]
		} else {
			break
		}

		for _, pd := range config.Config.Pods {
			if pd.Name == keys[0] {
				if len(keys) == 4 {
					if keys[1] == "units" {
						for _, unit := range pd.Units {
							if unit.Name == keys[2] {
								resrcInf = unit
								break
							}
						}
					}
				} else {
					resrcInf = pd
				}
				break
			}
		}
		break
	case "units":
		if len(keys) != 2 {
			break
		}
		key = keys[1]

		for _, pd := range config.Config.Pods {
			for _, unit := range pd.Units {
				if unit.Name == keys[0] {
					resrcInf = unit
					break
				}
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
