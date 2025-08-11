package uhandlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/acme"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/utils"
)

type certificateData struct {
	Id          primitive.ObjectID `json:"id"`
	Name        string             `json:"name"`
	Comment     string             `json:"comment"`
	Type        string             `json:"type"`
	Key         string             `json:"key"`
	Certificate string             `json:"certificate"`
	AcmeDomains []string           `json:"acme_domains"`
	AcmeAuth    string             `json:"acme_auth"`
	AcmeSecret  primitive.ObjectID `json:"acme_secret"`
}

type certificatesData struct {
	Certificates []*certificate.Certificate `json:"certificates"`
	Count        int64                      `json:"count"`
}

func certificatePut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)
	data := &certificateData{}

	certId, ok := utils.ParseObjectId(c.Param("cert_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	cert, err := certificate.GetOrg(db, userOrg, certId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if cert.Type == certificate.LetsEncrypt &&
		cert.AcmeType != certificate.AcmeDNS ||
		cert.AcmeType == certificate.AcmeHTTP {

		errData := &errortypes.ErrorData{
			Error:   "acme_type_blocked",
			Message: "Cannot modify LetsEncrypt HTTP verified certificates",
		}
		c.JSON(400, errData)
		return
	}

	if !data.AcmeSecret.IsZero() {
		exists, err := secret.ExistsOrg(db, userOrg, data.AcmeSecret)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
		if !exists {
			utils.AbortWithStatus(c, 405)
			return
		}
	} else {
		data.AcmeSecret = primitive.NilObjectID
	}

	cert.Name = data.Name
	cert.Comment = data.Comment
	cert.Key = data.Key
	cert.Certificate = data.Certificate
	cert.Type = data.Type
	cert.AcmeDomains = data.AcmeDomains
	cert.AcmeType = certificate.AcmeDNS
	cert.AcmeAuth = data.AcmeAuth
	cert.AcmeSecret = data.AcmeSecret

	fields := set.NewSet(
		"name",
		"comment",
		"type",
		"acme_domains",
		"acme_type",
		"acme_auth",
		"acme_secret",
		"info",
	)

	if cert.Type != certificate.LetsEncrypt {
		cert.Key = data.Key
		fields.Add("key")
		cert.Certificate = data.Certificate
		fields.Add("certificate")
	}

	errData, err := cert.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = cert.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if cert.Type == certificate.LetsEncrypt {
		acme.RenewBackground(cert)
	}

	event.PublishDispatch(db, "certificate.change")

	c.JSON(200, cert)
}

func certificatePost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)
	data := &certificateData{
		Name: "New Certificate",
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	cert := &certificate.Certificate{
		Name:         data.Name,
		Comment:      data.Comment,
		Organization: userOrg,
		Type:         data.Type,
		AcmeDomains:  data.AcmeDomains,
		AcmeType:     certificate.AcmeDNS,
		AcmeAuth:     data.AcmeAuth,
		AcmeSecret:   data.AcmeSecret,
	}

	if cert.Type != certificate.LetsEncrypt {
		cert.Key = data.Key
		cert.Certificate = data.Certificate
	}

	if !cert.AcmeSecret.IsZero() {
		_, err = secret.GetOrg(db, userOrg, cert.AcmeSecret)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
	} else {
		cert.AcmeSecret = primitive.NilObjectID
	}

	errData, err := cert.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = cert.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if cert.Type == certificate.LetsEncrypt {
		acme.RenewBackground(cert)
	}

	event.PublishDispatch(db, "certificate.change")

	c.JSON(200, cert)
}

func certificateDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)

	certId, ok := utils.ParseObjectId(c.Param("cert_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := certificate.RemoveOrg(db, userOrg, certId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "certificate.change")

	c.JSON(200, nil)
}

func certificatesDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)
	data := []primitive.ObjectID{}

	err := c.Bind(&data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	err = certificate.RemoveMultiOrg(db, userOrg, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "certificate.change")

	c.JSON(200, nil)
}

func certificateGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)

	if c.Query("names") == "true" {
		certs, err := certificate.GetAllNames(db, &bson.M{
			"organization": userOrg,
		})
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		c.JSON(200, certs)
		return
	}

	certId, ok := utils.ParseObjectId(c.Param("cert_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	cert, err := certificate.GetOrg(db, userOrg, certId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if demo.IsDemo() {
		cert.Key = "demo"
		cert.AcmeAccount = "demo"
	}

	c.JSON(200, cert)
}

func certificatesGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	userOrg := c.MustGet("organization").(primitive.ObjectID)

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{
		"organization": userOrg,
	}

	certificateId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = certificateId
	}

	name := strings.TrimSpace(c.Query("name"))
	if name != "" {
		query["name"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", regexp.QuoteMeta(name)),
			"$options": "i",
		}
	}

	comment := strings.TrimSpace(c.Query("comment"))
	if comment != "" {
		query["comment"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", comment),
			"$options": "i",
		}
	}

	certs, count, err := certificate.GetAllPaged(db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	data := &certificatesData{
		Certificates: certs,
		Count:        count,
	}

	c.JSON(200, data)
}
