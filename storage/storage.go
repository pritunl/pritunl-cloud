package storage

import (
	"net/url"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Storage struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Comment   string             `bson:"comment" json:"comment"`
	Type      string             `bson:"type" json:"type"`
	Endpoint  string             `bson:"endpoint" json:"endpoint"`
	Bucket    string             `bson:"bucket" json:"bucket"`
	AccessKey string             `bson:"access_key" json:"access_key"`
	SecretKey string             `bson:"secret_key" json:"secret_key"`
	Insecure  bool               `bson:"insecure" json:"insecure"`
}

type Completion struct {
	Id   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name string             `bson:"name" json:"name"`
	Type string             `bson:"type" json:"type"`
}

func (s *Storage) IsOracle() bool {
	return strings.Contains(strings.ToLower(s.Endpoint), "oracle")
}

func (s *Storage) GetWebUrl() (u *url.URL) {
	u = &url.URL{}

	if s.Insecure {
		u.Scheme = "http"
	} else {
		u.Scheme = "https"
	}
	u.Host = s.Endpoint
	u.Path = "/" + s.Bucket

	return
}

func (s *Storage) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	s.Name = utils.FilterName(s.Name)

	switch s.Type {
	case Public:
		break
	case Private:
		break
	case Web:
		break
	case "":
		s.Type = Public
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "invalid_type",
			Message: "Storage type is invalid",
		}
		return
	}

	return
}

func (s *Storage) Commit(db *database.Database) (err error) {
	coll := db.Storages()

	err = coll.Commit(s.Id, s)
	if err != nil {
		return
	}

	return
}

func (s *Storage) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Storages()

	err = coll.CommitFields(s.Id, s, fields)
	if err != nil {
		return
	}

	return
}

func (s *Storage) Insert(db *database.Database) (err error) {
	coll := db.Storages()

	if !s.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("storage: Storage already exists"),
		}
		return
	}

	resp, err := coll.InsertOne(db, s)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	s.Id = resp.InsertedID.(primitive.ObjectID)

	return
}
