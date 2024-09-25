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
	"gopkg.in/yaml.v3"
)

var yamlSpec = regexp.MustCompile("(?s)```yaml(.*?)```")

type Unit struct {
	Service     *Service           `bson:"-" json:"-"`
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
	Id primitive.ObjectID `bson:"id" json:"id"`
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

func (u *Unit) Reserve(db *database.Database, deployId primitive.ObjectID) (
	reserved bool, err error) {

	coll := db.Services()

	if len(u.Deployments) >= u.Count {
		return
	}

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
				Id: deployId,
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

func (u *Unit) RemoveDeployement(db *database.Database,
	deployId primitive.ObjectID) (err error) {

	coll := db.Services()

	updateOpts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"elem.id": u.Id},
		},
	})
	_, err = coll.UpdateOne(db, bson.M{
		"_id": u.Service.Id,
	}, bson.M{
		"$pull": bson.M{
			"units.$[elem].deployments": &bson.M{
				"id": deployId,
			},
		},
	}, updateOpts)
	if err != nil {
		err = database.ParseError(err)
		return
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

	if dataYaml.Name == "" {
		errData = &errortypes.ErrorData{
			Error:   "unit_name_missing",
			Message: "Unit name is missing",
		}
		return
	}

	if dataYaml.Kind != "instance" {
		errData = &errortypes.ErrorData{
			Error:   "unit_kind_invalid",
			Message: "Unit kind is invalid",
		}
		return
	}

	if dataYaml.Count == 0 {
		errData = &errortypes.ErrorData{
			Error:   "unit_count_missing",
			Message: "Unit count is missing",
		}
		return
	}

	resources := &Resources{
		Organization: srvc.Organization,
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

	if data.Node.IsZero() && data.Shape.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "unit_image_missing",
			Message: "Unit image is missing",
		}
		return
	}

	data.Roles = dataYaml.Roles
	data.Processors = dataYaml.Processors
	data.Memory = dataYaml.Memory
	data.DiskSize = dataYaml.DiskSize

	u.Name = dataYaml.Name
	u.Kind = dataYaml.Kind
	u.Count = dataYaml.Count
	u.Instance = data

	return
}
