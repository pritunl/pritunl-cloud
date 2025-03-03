package resource

import (
	"github.com/pritunl/pritunl-cloud/finder"
	"github.com/pritunl/pritunl-cloud/imds/server/config"
)

func Query(resrc string, keys ...string) (val string, err error) {
	var resrcInf interface{}

	key := ""

	switch resrc {
	case finder.InstanceKind:
		if len(keys) == 2 {
			if keys[0] != "local" {
				break
			}
			key = keys[1]
		} else if len(keys) == 1 {
			key = keys[0]
		} else {
			break
		}
		resrcInf = config.Config.Instance
		break
	case finder.VpcKind:
		if len(keys) == 2 {
			if keys[0] != "local" {
				break
			}
			key = keys[1]
		} else if len(keys) == 1 {
			key = keys[0]
		} else {
			break
		}
		resrcInf = config.Config.Vpc
		break
	case finder.SubnetKind:
		if len(keys) == 2 {
			if keys[0] != "local" {
				break
			}
			key = keys[1]
		} else if len(keys) == 1 {
			key = keys[0]
		} else {
			break
		}
		resrcInf = config.Config.Subnet
		break
	case finder.SecretKind:
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
	case finder.CertificateKind:
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
	case finder.PodKind:
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
					if keys[1] == finder.UnitKind {
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
	case finder.UnitKind:
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
