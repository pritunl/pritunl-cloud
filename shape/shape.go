package shape

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/zone"
)

type Shape struct {
	Id               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name             string             `bson:"name" json:"name"`
	Comment          string             `bson:"comment" json:"comment"`
	Type             string             `bson:"type" json:"type"`
	DeleteProtection bool               `bson:"delete_protection" json:"delete_protection"`
	Datacenter       primitive.ObjectID `bson:"datacenter" json:"datacenter"`
	Roles            []string           `bson:"roles" json:"roles"`
	Flexible         bool               `bson:"flexible" json:"flexible"`
	DiskType         string             `bson:"disk_type" json:"disk_type"`
	DiskPool         primitive.ObjectID `bson:"disk_pool" json:"disk_pool"`
	Memory           int                `bson:"memory" json:"memory"`
	Processors       int                `bson:"processors" json:"processors"`
	NodeCount        int                `bson:"-" json:"node_count"`
}

func (s *Shape) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	s.Name = utils.FilterName(s.Name)

	if s.Type == "" {
		s.Type = Instance
	}

	switch s.Type {
	case Instance:
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "invalid_shape_type",
			Message: "Shape type invalid",
		}
		return
	}

	if s.Datacenter.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "missing_datacenter",
			Message: "Shape datacenter required",
		}
		return
	}

	return
}

func (s *Shape) FindNode(db *database.Database, processors, memory int) (
	nde *node.Node, err error) {

	zones, err := zone.GetAllDatacenter(db, s.Datacenter)
	if err != nil {
		return
	}

	zoneIds := []primitive.ObjectID{}
	for _, zne := range zones {
		zoneIds = append(zoneIds, zne.Id)
	}

	ndes, err := node.GetAllShape(db, zoneIds, s.Roles)
	if err != nil {
		return
	}

	Nodes(ndes).Sort()

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
