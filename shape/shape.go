package shape

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
)

type Shape struct {
	Id               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name             string             `bson:"name" json:"name"`
	Comment          string             `bson:"comment" json:"comment"`
	Type             string             `bson:"type" json:"type"`
	DeleteProtection bool               `bson:"delete_protection" json:"delete_protection"`
	Zone             primitive.ObjectID `bson:"zone" json:"zone"`
	Roles            []string           `bson:"roles" json:"roles"`
	Flexible         bool               `bson:"flexible" json:"flexible"`
	DiskType         string             `bson:"disk_type" json:"disk_type"`
	DiskPool         primitive.ObjectID `bson:"disk_pool" json:"disk_pool"`
	Memory           int                `bson:"memory" json:"memory"`
	Processors       int                `bson:"processors" json:"processors"`
}

func (s *Shape) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if s.Type == "" {
		s.Type = Instance
	}

	switch s.Type {
	case Instance:
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "invalid_secret_type",
			Message: "Shape type invalid",
		}
		return
	}

	return
}

func (s *Shape) FindNode(db *database.Database, processors, memory int) (
	nde *node.Node, err error) {

	ndes, err := node.GetAllShape(db, s.Zone, s.Roles)
	if err != nil {
		return
	}

	NodeUsageSort(ndes)

	for _, nd := range ndes {
		nde = nd
		return
	}

	err = &errortypes.NotFoundError{
		errors.New("shape: Failed to find available node"),
	}
	return
}

func (s *Shape) Commit(db *database.Database) (err error) {
	coll := db.Shapes()

	err = coll.Commit(s.Id, s)
	if err != nil {
		return
	}

	return
}

func (s *Shape) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Shapes()

	err = coll.CommitFields(s.Id, s, fields)
	if err != nil {
		return
	}

	return
}

func (s *Shape) Insert(db *database.Database) (err error) {
	coll := db.Shapes()

	_, err = coll.InsertOne(db, s)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
