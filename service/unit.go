package service

import (
	"regexp"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"gopkg.in/yaml.v3"
)

var yamlSpec = regexp.MustCompile("(?s)```yaml(.*?)```")

type Unit struct {
	Id          primitive.ObjectID `bson:"id" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Kind        string             `bson:"kind" json:"kind"`
	Count       int                `bson:"count" json:"count"`
	Deployments []*Deployment      `bson:"deployments" json:"deployments"`
	Spec        string             `bson:"spec" json:"spec"`
	Instance    *Instance          `bson:"instance,omitempty" json:"instance,omitempty"`
}

type UnitInput struct {
	Id   primitive.ObjectID `bson:"id" json:"id"`
	Name string             `bson:"name" json:"name"`
	Spec string             `bson:"spec" json:"spec"`
}

type Deployment struct {
	Id    primitive.ObjectID `bson:"id" json:"id"`
	Node  primitive.ObjectID `bson:"node" json:"node"`
	State string             `bson:"state" json:"state"`
}

type Instance struct {
	Zone       primitive.ObjectID `bson:"zone" json:"zone"`
	Node       primitive.ObjectID `bson:"node,omitempty" json:"node"`
	Shape      primitive.ObjectID `bson:"shape,omitempty" json:"shape"`
	Vpc        primitive.ObjectID `bson:"vpc" json:"vpc"`
	Subnet     primitive.ObjectID `bson:"subnet" json:"subnet"`
	Roles      []string           `bson:"roles" json:"roles"`
	Processors int                `bson:"processors" json:"processors"`
	Memory     int                `bson:"memory" json:"memory"`
	Image      primitive.ObjectID `bson:"image" json:"image"`
	DiskSize   int                `bson:"disk_size" json:"disk_size"`
}

type InstanceYaml struct {
	Name       string   `yaml:"name"`
	Kind       string   `yaml:"kind"`
	Count      int      `yaml:"count"`
	Zone       string   `yaml:"zone"`
	Node       string   `yaml:"node,omitempty"`
	Shape      string   `yaml:"shape,omitempty"`
	Vpc        string   `yaml:"vpc"`
	Subnet     string   `yaml:"subnet"`
	Roles      []string `yaml:"roles"`
	Processors int      `yaml:"processors"`
	Memory     int      `yaml:"memory"`
	Image      string   `yaml:"image"`
	DiskSize   int      `yaml:"disk_size"`
}

func (u *Unit) Reserve(db *database.Database) (
	deployementId primitive.ObjectID, reserved bool, err error) {

	coll := db.Services()

	if len(u.Deployments) >= u.Count {
		return
	}

	deployementId = primitive.NewObjectID()

	updateOpts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{
				"elem.id":    u.Id,
				"elem.count": u.Count,
				"elem.deployments": bson.M{
					"$size": len(u.Deployments),
				},
			},
		},
	})
	resp, err := coll.UpdateOne(db, bson.M{
		"_id": u.Service.Id,
	}, bson.M{
		"$push": bson.M{
			"units.$[elem].deployments": &Deployment{
				Id:    deployementId,
				Node:  node.Self.Id,
				State: Reserved,
			},
		},
	}, updateOpts)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if resp.MatchedCount == 1 && resp.ModifiedCount == 1 {
		reserved = true
	}

	return
}

func (u *Unit) UpdateDeployement(db *database.Database,
	deploymentId primitive.ObjectID, state string) (updated bool, err error) {

	coll := db.Services()

	updateOpts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"elem.id": u.Id},
			bson.M{"deploy.id": deploymentId},
		},
	})
	resp, err := coll.UpdateOne(db, bson.M{
		"_id": u.Service.Id,
	}, bson.M{
		"$set": bson.M{
			"units.$[elem].deployments.$[deploy].state": state,
		},
	}, updateOpts)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if resp.MatchedCount == 1 && resp.ModifiedCount == 1 {
		updated = true
	}

	return
}

func (u *Unit) ExtractSpec() (spec string, err error) {
	matches := yamlSpec.FindStringSubmatch(u.Spec)
	if len(matches) > 1 {
		spec = matches[1]
		return
	}

	return
}

func (u *Unit) Parse(db *database.Database, srvc *Service) (
	errData *errortypes.ErrorData, err error) {

	spec, err := u.ExtractSpec()
	if err != nil {
		return
	}

	if spec == "" {
		errData = &errortypes.ErrorData{
			Error:   "unit_yaml_block_missing",
			Message: "Unit missing yaml spec block",
		}
		return
	}

	data := &Instance{}
	dataYaml := &InstanceYaml{}

	err = yaml.Unmarshal([]byte(spec), dataYaml)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "service: Failed to parse yaml spec"),
		}
		return
	}

	if dataYaml.Zone != "" {
		resources := &Resources{
			Organization: srvc.Organization,
		}
		err = resources.Find(db, dataYaml.Zone)
		if err != nil {
			return
		}
		if resources.Zone != nil {
			data.Zone = resources.Zone.Id
		}
	}
	if dataYaml.Node != "" {
		resources := &Resources{
			Organization: srvc.Organization,
		}
		err = resources.Find(db, dataYaml.Node)
		if err != nil {
			return
		}
		if resources.Node != nil {
			data.Node = resources.Node.Id
		}
	}
	if dataYaml.Shape != "" {
		resources := &Resources{
			Organization: srvc.Organization,
		}
		err = resources.Find(db, dataYaml.Shape)
		if err != nil {
			return
		}
		if resources.Shape != nil {
			data.Shape = resources.Shape.Id
		}
	}
	if dataYaml.Vpc != "" {
		resources := &Resources{
			Organization: srvc.Organization,
		}
		err = resources.Find(db, dataYaml.Vpc)
		if err != nil {
			return
		}
		if resources.Vpc != nil {
			data.Vpc = resources.Vpc.Id
		}
	}
	if dataYaml.Subnet != "" {
		resources := &Resources{
			Organization: srvc.Organization,
		}
		err = resources.Find(db, dataYaml.Subnet)
		if err != nil {
			return
		}
		if resources.Subnet != nil {
			data.Subnet = resources.Subnet.Id
		}
	}
	if dataYaml.Image != "" {
		resources := &Resources{
			Organization: srvc.Organization,
		}
		err = resources.Find(db, dataYaml.Image)
		if err != nil {
			return
		}
		if resources.Image != nil {
			data.Image = resources.Image.Id
		}
	}

	u.Instance = data

	return
}
