package spec

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/shape"
	"gopkg.in/yaml.v2"
)

type Commit struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Service   primitive.ObjectID `bson:"service" json:"service"`
	Unit      primitive.ObjectID `bson:"unit" json:"unit"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
	Name      string             `bson:"name" json:"name"`
	Kind      string             `bson:"kind" json:"kind"`
	Count     int                `bson:"count" json:"count"`
	Hash      string             `bson:"hash" json:"hash"`
	Data      string             `bson:"data" json:"data"`
	Instance  *Instance          `bson:"instance_data,omitempty" json:"-"`
}

func (s *Commit) Validate(db *database.Database) (err error) {
	if s.Timestamp.IsZero() {
		s.Timestamp = time.Now()
	}

	return
}

func (u *Commit) ExtractResources() (resources string, err error) {
	matches := resourcesRe.FindStringSubmatch(u.Data)
	if len(matches) > 1 {
		resources = matches[1]
		resources = strings.TrimSpace(resources)
		return
	}

	return
}

func (u *Commit) Parse(db *database.Database,
	orgId primitive.ObjectID) (errData *errortypes.ErrorData, err error) {

	hash := sha1.New()
	hash.Write([]byte(filterSpecHash(u.Data)))
	hashBytes := hash.Sum(nil)
	u.Hash = fmt.Sprintf("%x", hashBytes)

	resourcesSpec, err := u.ExtractResources()
	if err != nil {
		return
	}

	if resourcesSpec == "" {
		errData = &errortypes.ErrorData{
			Error:   "unit_resources_block_missing",
			Message: "Unit missing yaml resources block",
		}
		return
	}

	data := &Instance{}
	dataYaml := &InstanceYaml{}
	var shpe *shape.Shape

	err = yaml.Unmarshal([]byte(resourcesSpec), dataYaml)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "service: Failed to parse yaml resources"),
		}
		return
	}

	if dataYaml.Name == "" {
		errData = &errortypes.ErrorData{
			Error:   "unit_name_missing",
			Message: "Unit name is missing",
		}
		return
	}

	switch dataYaml.Kind {
	case deployment.Instance:
		break
	case deployment.Image:
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "unit_kind_invalid",
			Message: "Unit kind is invalid",
		}
		return
	}

	resources := &Resources{
		Organization: orgId,
	}

	if dataYaml.Plan != "" {
		kind, e := resources.Find(db, dataYaml.Plan)
		if e != nil {
			err = e
			return
		}
		if kind == "plan" && resources.Plan != nil {
			data.Plan = resources.Plan.Id
		}
	}

	if dataYaml.Zone != "" {
		kind, e := resources.Find(db, dataYaml.Zone)
		if e != nil {
			err = e
			return
		}
		if kind == "zone" && resources.Zone != nil {
			data.Zone = resources.Zone.Id
		}
	}

	if data.Zone.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "unit_zone_missing",
			Message: "Unit zone is missing",
		}
		return
	}

	if dataYaml.Node != "" {
		kind, e := resources.Find(db, dataYaml.Node)
		if e != nil {
			err = e
			return
		}
		if kind == "node" && resources.Node != nil {
			data.Node = resources.Node.Id
		}
	}
	if dataYaml.Shape != "" {
		kind, e := resources.Find(db, dataYaml.Shape)
		if e != nil {
			err = e
			return
		}
		if kind == "shape" && resources.Shape != nil {
			shpe = resources.Shape
			data.Shape = resources.Shape.Id
		}
	}

	if data.Node.IsZero() && data.Shape.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "unit_node_missing",
			Message: "Unit node or shape is missing",
		}
		return
	}

	if dataYaml.Vpc != "" {
		kind, e := resources.Find(db, dataYaml.Vpc)
		if e != nil {
			err = e
			return
		}
		if kind == "vpc" && resources.Vpc != nil {
			data.Vpc = resources.Vpc.Id
		}
	}

	if data.Node.IsZero() && data.Shape.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "unit_vpc_missing",
			Message: "Unit VPC is missing",
		}
		return
	}

	if dataYaml.Subnet != "" {
		kind, e := resources.Find(db, dataYaml.Subnet)
		if e != nil {
			err = e
			return
		}
		if kind == "subnet" && resources.Subnet != nil {
			data.Subnet = resources.Subnet.Id
		}
	}

	if data.Node.IsZero() && data.Shape.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "unit_vpc_missing",
			Message: "Unit subnet is missing",
		}
		return
	}

	if dataYaml.Image != "" {
		kind, e := resources.Find(db, dataYaml.Image)
		if e != nil {
			err = e
			return
		}
		if kind == "image" && resources.Image != nil {
			data.Image = resources.Image.Id
		}
	}

	if dataYaml.Certificates != nil {
		for _, cert := range dataYaml.Certificates {
			kind, e := resources.Find(db, cert)
			if e != nil {
				err = e
				return
			}
			if kind == "certificate" && resources.Certificate != nil {
				data.Certificates = append(
					data.Certificates,
					resources.Certificate.Id,
				)
			}
		}
	}

	if dataYaml.Secrets != nil {
		for _, cert := range dataYaml.Secrets {
			kind, e := resources.Find(db, cert)
			if e != nil {
				err = e
				return
			}
			if kind == "secret" && resources.Secret != nil {
				data.Secrets = append(
					data.Secrets,
					resources.Secret.Id,
				)
			}
		}
	}

	if dataYaml.Services != nil {
		for _, cert := range dataYaml.Services {
			kind, e := resources.Find(db, cert)
			if e != nil {
				err = e
				return
			}
			if kind == "service" && resources.Service != nil {
				data.Services = append(
					data.Services,
					resources.Service.Id,
				)
			}
		}
	}

	if data.Node.IsZero() && data.Shape.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "unit_image_missing",
			Message: "Unit image is missing",
		}
		return
	}

	if shpe != nil {
		data.Processors = shpe.Processors
		data.Memory = shpe.Memory
		if shpe.Flexible {
			if dataYaml.Processors != 0 {
				data.Processors = dataYaml.Processors
			}
			if dataYaml.Memory != 0 {
				data.Memory = dataYaml.Memory
			}
		}
	} else {
		data.Processors = dataYaml.Processors
		data.Memory = dataYaml.Memory
	}

	data.Roles = dataYaml.Roles
	data.DiskSize = dataYaml.DiskSize

	u.Name = dataYaml.Name
	u.Kind = dataYaml.Kind
	u.Count = dataYaml.Count
	u.Instance = data

	u.Count = dataYaml.Count
	if u.Kind == ImageKind && u.Count != 0 {
		errData = &errortypes.ErrorData{
			Error:   "count_invalid",
			Message: "Count not valid for image kind",
		}
		return
	}

	return
}

func (s *Commit) Commit(db *database.Database) (err error) {
	coll := db.Specs()

	err = coll.Commit(s.Id, s)
	if err != nil {
		return
	}

	return
}

func (s *Commit) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Specs()

	err = coll.CommitFields(s.Id, s, fields)
	if err != nil {
		return
	}

	return
}

func (s *Commit) Insert(db *database.Database) (err error) {
	coll := db.Specs()

	resp, err := coll.InsertOne(db, s)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	s.Id = resp.InsertedID.(primitive.ObjectID)

	return
}
