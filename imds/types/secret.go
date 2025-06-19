package types

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/secret"
)

type Secret struct {
	Id         primitive.ObjectID `json:"id"`
	Name       string             `json:"name"`
	Type       string             `json:"type"`
	Key        string             `json:"key"`
	Value      string             `json:"value"`
	Data       string             `json:"data"`
	Region     string             `json:"region"`
	PublicKey  string             `json:"public_key"`
	PrivateKey string             `json:"private_key"`
}

func NewSecrets(secrs []*secret.Secret) []*Secret {
	datas := []*Secret{}

	for _, secr := range secrs {
		if secr == nil {
			continue
		}

		data := &Secret{
			Id:         secr.Id,
			Name:       secr.Name,
			Type:       secr.Type,
			Key:        secr.Key,
			Value:      secr.Value,
			Data:       secr.Data,
			Region:     secr.Region,
			PublicKey:  secr.PublicKey,
			PrivateKey: secr.PrivateKey,
		}

		datas = append(datas, data)
	}

	return datas
}
