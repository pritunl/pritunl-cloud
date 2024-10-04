package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/imds/server/config"
)

type certificateData struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Type        string             `bson:"type" json:"type"`
	Key         string             `bson:"key" json:"key"`
	Certificate string             `bson:"certificate" json:"certificate"`
}

func certificatesGet(c *gin.Context) {
	certs := []*certificateData{}

	for _, cert := range config.Config.Certificates {
		certData := &certificateData{
			Id:          cert.Id,
			Name:        cert.Name,
			Type:        cert.Type,
			Key:         cert.Key,
			Certificate: cert.Certificate,
		}

		certs = append(certs, certData)
	}

	c.JSON(200, certs)
}
