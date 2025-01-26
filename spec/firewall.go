package spec

import (
	"strconv"
	"strings"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type Firewall struct {
	Ingress []*Rule `bson:"ingress" json:"ingress"`
}

type Rule struct {
	Protocol string               `bson:"protocol" json:"protocol"`
	Port     string               `bson:"port" json:"port"`
	Pods     []primitive.ObjectID `bson:"pods" json:"pods"`
	Units    []primitive.ObjectID `bson:"units" json:"units"`
}

func (f *Firewall) Validate() (errData *errortypes.ErrorData, err error) {
	if f.Ingress == nil {
		f.Ingress = []*Rule{}
	}

	for _, rule := range f.Ingress {
		switch rule.Protocol {
		case All:
			rule.Port = ""
			break
		case Icmp:
			rule.Port = ""
			break
		case Tcp, Udp, Multicast, Broadcast:
			ports := strings.Split(rule.Port, "-")

			portInt, e := strconv.Atoi(ports[0])
			if e != nil {
				errData = &errortypes.ErrorData{
					Error:   "invalid_ingress_rule_port",
					Message: "Invalid ingress rule port",
				}
				return
			}

			if portInt < 1 || portInt > 65535 {
				errData = &errortypes.ErrorData{
					Error:   "invalid_ingress_rule_port",
					Message: "Invalid ingress rule port",
				}
				return
			}

			parsedPort := strconv.Itoa(portInt)
			if len(ports) > 1 {
				portInt2, e := strconv.Atoi(ports[1])
				if e != nil {
					errData = &errortypes.ErrorData{
						Error:   "invalid_ingress_rule_port",
						Message: "Invalid ingress rule port",
					}
					return
				}

				if portInt < 1 || portInt > 65535 || portInt2 <= portInt {
					errData = &errortypes.ErrorData{
						Error:   "invalid_ingress_rule_port",
						Message: "Invalid ingress rule port",
					}
					return
				}

				parsedPort += "-" + strconv.Itoa(portInt2)
			}

			rule.Port = parsedPort

			break
		default:
			errData = &errortypes.ErrorData{
				Error:   "invalid_ingress_rule_protocol",
				Message: "Invalid ingress rule protocol",
			}
			return
		}

		if rule.Protocol == Multicast || rule.Protocol == Broadcast ||
			rule.Units == nil {

			rule.Pods = []primitive.ObjectID{}
			rule.Units = []primitive.ObjectID{}
		}
	}

	return
}

type FirewallYaml struct {
	Name    string                `yaml:"name"`
	Kind    string                `yaml:"kind"`
	Ingress []FirewallYamlIngress `yaml:"ingress"`
}

type FirewallYamlIngress struct {
	Protocol string   `yaml:"protocol"`
	Port     string   `yaml:"port"`
	Source   []string `yaml:"source"`
}
