package types

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/certificate"
)

type Certificate struct {
	Id          bson.ObjectID `json:"id"`
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Key         string        `json:"key"`
	Certificate string        `json:"certificate"`
}

func NewCertificates(certs []*certificate.Certificate) []*Certificate {
	datas := []*Certificate{}

	for _, cert := range certs {
		if cert == nil {
			continue
		}

		data := &Certificate{
			Id:          cert.Id,
			Name:        cert.Name,
			Type:        cert.Type,
			Key:         cert.Key,
			Certificate: cert.Certificate,
		}

		datas = append(datas, data)
	}

	return datas
}
