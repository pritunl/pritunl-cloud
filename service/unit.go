package service

import (
	"regexp"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
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
	Node  primitive.ObjectID `bson:"node" json:"node"`
	State string             `bson:"state" json:"state"` // reserved, deployed
	Data  interface{}        `bson:"data" json:"data"`
}

type Instance struct {
	Name       string             `bson:"name" json:"name"`
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
	Count      int      `yaml:"number"`
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
