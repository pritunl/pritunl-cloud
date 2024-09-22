package service

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Service struct {
	Id               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name             string             `bson:"name" json:"name"`
	Comment          string             `bson:"comment" json:"comment"`
	Organization     primitive.ObjectID `json:"organization"`
	DeleteProtection bool               `bson:"delete_protection" json:"delete_protection"`
	Units            []*Unit            `bson:"units" json:"units"`
}

type Unit struct {
	Name string `bson:"name" json:"name"`
	Spec string `bson:"spec" json:"spec"`
}

func (p *Service) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	p.Name = utils.FilterName(p.Name)

	return
}

func (p *Service) Commit(db *database.Database) (err error) {
	coll := db.Services()

	err = coll.Commit(p.Id, p)
	if err != nil {
		return
	}

	return
}

func (p *Service) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Services()

	err = coll.CommitFields(p.Id, p, fields)
	if err != nil {
		return
	}

	return
}

func (p *Service) Insert(db *database.Database) (err error) {
	coll := db.Services()

	_, err = coll.InsertOne(db, p)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
