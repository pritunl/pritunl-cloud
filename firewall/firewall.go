package firewall

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Rule struct {
	SourceIps []string `bson:"source_ips" json:"source_ips"`
	Protocol  string   `bson:"protocol" json:"protocol"`
	Port      string   `bson:"port" json:"port"`
}

type Mapping struct {
	Ipvs         bool   `bson:"ipvs" json:"ipvs"`
	Address      string `bson:"adress" json:"adress"`
	Protocol     string `bson:"protocol" json:"protocol"`
	ExternalPort int    `bson:"external_port" json:"external_port"`
	InternalPort int    `bson:"internal_port" json:"internal_port"`
}

func (r *Rule) SetName(ipv6 bool) (name string) {
	switch r.Protocol {
	case All:
		if ipv6 {
			name = "pr6_all"
		} else {
			name = "pr4_all"
		}
		break
	case Icmp:
		if ipv6 {
			name = "pr6_icmp"
		} else {
			name = "pr4_icmp"
		}
		break
	case Multicast:
		if ipv6 {
			name = "pr6_multi"
		} else {
			name = "pr4_multi"
		}
		break
	case Broadcast:
		if ipv6 {
			name = "pr6_broad"
		} else {
			name = "pr4_broad"
		}
		break
	case Tcp, Udp:
		if ipv6 {
			name = fmt.Sprintf(
				"pr6_%s_%s",
				r.Protocol,
				strings.Replace(r.Port, "-", "_", 1),
			)
		} else {
			name = fmt.Sprintf(
				"pr4_%s_%s",
				r.Protocol,
				strings.Replace(r.Port, "-", "_", 1),
			)
		}
		break
	default:
		break
	}

	return
}

type Firewall struct {
	Id           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Comment      string             `bson:"comment" json:"comment"`
	Organization primitive.ObjectID `bson:"organization,omitempty" json:"organization"`
	NetworkRoles []string           `bson:"network_roles" json:"network_roles"`
	Ingress      []*Rule            `bson:"ingress" json:"ingress"`
}

func (f *Firewall) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	f.Name = utils.FilterName(f.Name)

	if f.NetworkRoles == nil {
		f.NetworkRoles = []string{}
	}

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

		if rule.Protocol == Multicast || rule.Protocol == Broadcast {
			rule.SourceIps = []string{}
		} else {
			for i, sourceIp := range rule.SourceIps {
				if sourceIp == "" {
					errData = &errortypes.ErrorData{
						Error:   "invalid_ingress_rule_source_ip",
						Message: "Empty ingress rule source IP",
					}
					return
				}

				if !strings.Contains(sourceIp, "/") {
					if strings.Contains(sourceIp, ":") {
						sourceIp += "/128"
					} else {
						sourceIp += "/32"
					}
				}

				_, sourceCidr, e := net.ParseCIDR(sourceIp)
				if e != nil {
					errData = &errortypes.ErrorData{
						Error:   "invalid_ingress_rule_source_ip",
						Message: "Invalid ingress rule source IP",
					}
					return
				}

				rule.SourceIps[i] = sourceCidr.String()
			}
		}
	}

	return
}

func (f *Firewall) Commit(db *database.Database) (err error) {
	coll := db.Firewalls()

	err = coll.Commit(f.Id, f)
	if err != nil {
		return
	}

	return
}

func (f *Firewall) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Firewalls()

	err = coll.CommitFields(f.Id, f, fields)
	if err != nil {
		return
	}

	return
}

func (f *Firewall) Insert(db *database.Database) (err error) {
	coll := db.Firewalls()

	if !f.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("firewall: Firewall already exists"),
		}
		return
	}

	resp, err := coll.InsertOne(db, f)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	f.Id = resp.InsertedID.(primitive.ObjectID)

	return
}
