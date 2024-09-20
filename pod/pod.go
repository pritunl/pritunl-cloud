package pod

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Pod struct {
	Id               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name             string             `bson:"name" json:"name"`
	Comment          string             `bson:"comment" json:"comment"`
	Organization     primitive.ObjectID `json:"organization"`
	Type             string             `bson:"type" json:"type"`
	DeleteProtection bool               `bson:"delete_protection" json:"delete_protection"`
	Zone             primitive.ObjectID `bson:"zone" json:"zone"`
	Roles            []string           `bson:"roles" json:"roles"`
	Spec             string             `bson:"spec" json:"spec"`
}

func (p *Pod) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	p.Name = utils.FilterName(p.Name)

	if p.Type == "" {
		p.Type = Todo
	}

	switch p.Type {
	case Todo:
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "invalid_pod_type",
			Message: "Pod type invalid",
		}
		return
	}

	return
}

func (p *Pod) Commit(db *database.Database) (err error) {
	coll := db.Pods()

	err = coll.Commit(p.Id, p)
	if err != nil {
		return
	}

	return
}

func (p *Pod) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Pods()

	err = coll.CommitFields(p.Id, p, fields)
	if err != nil {
		return
	}

	return
}

func (p *Pod) Insert(db *database.Database) (err error) {
	coll := db.Pods()

	_, err = coll.InsertOne(db, p)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
