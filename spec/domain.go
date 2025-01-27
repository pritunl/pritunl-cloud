package spec

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Domain struct {
	Records []*Record `bson:"record" json:"record"`
}

func (d *Domain) Validate() (errData *errortypes.ErrorData, err error) {
	for _, rec := range d.Records {
		rec.Name = utils.FilterDomain(rec.Name)

		switch rec.Type {
		case Private:
			break
		case Private6:
			break
		case Public:
			break
		case Public6:
			break
		case OraclePublic:
			break
		case OraclePrivate:
			break
		default:
			errData = &errortypes.ErrorData{
				Error:   "unknown_domain_record_type",
				Message: "Unknown domain record type",
			}
			return
		}
	}

	return
}

type Record struct {
	Name   string             `bson:"name" json:"name"`
	Domain primitive.ObjectID `bson:"domain" json:"domain"`
	Type   string             `bson:"type" json:"type"`
}

type DomainYaml struct {
	Name    string             `yaml:"name"`
	Kind    string             `yaml:"kind"`
	Records []DomainYamlRecord `yaml:"records"`
}

type DomainYamlRecord struct {
	Name   string `yaml:"name"`
	Domain string `yaml:"domain"`
	Type   string `yaml:"type"`
}
