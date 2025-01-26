package spec

import (
	"crypto/sha1"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/zone"
	"gopkg.in/yaml.v2"
)

type Commit struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Pod       primitive.ObjectID `bson:"pod" json:"pod"`
	Unit      primitive.ObjectID `bson:"unit" json:"unit"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
	Name      string             `bson:"name" json:"name"`
	Kind      string             `bson:"kind" json:"kind"`
	Count     int                `bson:"count" json:"count"`
	Hash      string             `bson:"hash" json:"hash"`
	Data      string             `bson:"data" json:"data"`
	Instance  *Instance          `bson:"instance,omitempty" json:"-"`
	Firewall  *Firewall          `bson:"firewall,omitempty" json:"-"`
}

func (c *Commit) GetAllNodes(db *database.Database,
	orgId primitive.ObjectID) (ndes Nodes,
	offlineCount, noMountCount int, err error) {

	shpe, err := shape.Get(db, c.Instance.Shape)
	if err != nil {
		return
	}

	zones, err := zone.GetAllDatacenter(db, shpe.Datacenter)
	if err != nil {
		return
	}

	zoneIds := []primitive.ObjectID{}
	for _, zne := range zones {
		zoneIds = append(zoneIds, zne.Id)
	}

	allNdes, err := node.GetAllShape(db, zoneIds, shpe.Roles)
	if err != nil {
		return
	}

	var mountNodes []set.Set
	if c.Instance.Mounts != nil && len(c.Instance.Mounts) > 0 {
		diskIds := []primitive.ObjectID{}
		for _, mount := range c.Instance.Mounts {
			if mount.Disks == nil {
				continue
			}
			diskIds = append(diskIds, mount.Disks...)
		}

		disksMap, e := disk.GetAllMap(db, &bson.M{
			"_id": &bson.M{
				"$in": diskIds,
			},
			"organization": orgId,
		})
		if e != nil {
			err = e
			return
		}

		for _, mount := range c.Instance.Mounts {
			mountSet := set.NewSet()

			if mount.Disks != nil {
				for _, dskId := range mount.Disks {
					dsk := disksMap[dskId]
					if dsk == nil {
						continue
					}
					mountSet.Add(dsk.Node)
				}
			}

			mountNodes = append(mountNodes, mountSet)
		}
	}

	ndes = Nodes{}
	for _, nde := range allNdes {
		if !nde.IsOnline() {
			offlineCount += 1
			continue
		}

		if mountNodes != nil {
			match := true
			for _, mountSet := range mountNodes {
				if !mountSet.Contains(nde.Id) {
					match = false
					break
				}
			}

			if !match {
				noMountCount += 1
				continue
			}
		}

		ndes = append(ndes, nde)
	}

	return
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

func (u *Commit) parseFirewall(db *database.Database,
	orgId primitive.ObjectID, dataYaml *FirewallYaml) (
	errData *errortypes.ErrorData, err error) {

	data := &Firewall{
		Ingress: []*Rule{},
	}

	if dataYaml.Kind != deployment.Firewall {
		errData = &errortypes.ErrorData{
			Error:   "unit_kind_mismatch",
			Message: "Unit kind unexpected",
		}
		return
	}

	resources := &Resources{
		Organization: orgId,
	}

	for _, ruleYaml := range dataYaml.Ingress {
		if ruleYaml.Source == nil {
			continue
		}

		rule := &Rule{
			Protocol: ruleYaml.Protocol,
			Port:     ruleYaml.Port,
		}
		units := set.NewSet()

		for _, source := range ruleYaml.Source {
			kind, e := resources.Find(db, source)
			if e != nil {
				err = e
				return
			}

			if kind == "unit" && resources.Unit != nil {
				units.Add(resources.Unit.Id)
			}
		}

		for unitId := range units.Iter() {
			rule.Units = append(rule.Units, unitId.(primitive.ObjectID))
		}

		data.Ingress = append(data.Ingress, rule)
	}

	errData, err = data.Validate()
	if err != nil || errData != nil {
		return
	}

	u.Firewall = data

	return
}

func (u *Commit) parseInstance(db *database.Database,
	orgId primitive.ObjectID, dataYaml *InstanceYaml) (
	errData *errortypes.ErrorData, err error) {

	data := &Instance{}
	var shpe *shape.Shape

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
			Error:   "unit_kind_mismatch",
			Message: "Unit kind unexpected",
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

	if dataYaml.Mounts != nil {
		for _, mount := range dataYaml.Mounts {
			mnt := Mount{
				Path:  filepath.Clean(mount.Path),
				Disks: []primitive.ObjectID{},
			}

			for _, dsk := range mount.Disks {
				kind, e := resources.Find(db, dsk)
				if e != nil {
					err = e
					return
				}
				if kind == "disk" && resources.Disks != nil {
					for _, dskRes := range resources.Disks {
						mnt.Disks = append(mnt.Disks, dskRes.Id)
					}
				}
			}

			data.Mounts = append(data.Mounts, mnt)
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

	if dataYaml.Pods != nil {
		for _, cert := range dataYaml.Pods {
			kind, e := resources.Find(db, cert)
			if e != nil {
				err = e
				return
			}
			if kind == "pod" && resources.Pod != nil {
				data.Pods = append(
					data.Pods,
					resources.Pod.Id,
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

	baseDecode := yaml.NewDecoder(strings.NewReader(resourcesSpec))
	decoder := yaml.NewDecoder(strings.NewReader(resourcesSpec))
	for {
		baseDoc := &Base{}

		err = baseDecode.Decode(baseDoc)
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}

			err = &errortypes.ParseError{
				errors.Wrap(err, "spec: Failed to decode yaml doc"),
			}
			return
		}

		switch baseDoc.Kind {
		case deployment.Instance, deployment.Image:
			instYaml := &InstanceYaml{}

			err = decoder.Decode(instYaml)
			if err != nil {
				err = &errortypes.ParseError{
					errors.Wrap(err,
						"spec: Failed to decode instance yaml doc"),
				}
				return
			}

			errData, err = u.parseInstance(db, orgId, instYaml)
			if err != nil || errData != nil {
				return
			}
		case deployment.Firewall:
			fireYaml := &FirewallYaml{}

			err = decoder.Decode(fireYaml)
			if err != nil {
				err = &errortypes.ParseError{
					errors.Wrap(err,
						"spec: Failed to decode firewall yaml doc"),
				}
				return
			}

			errData, err = u.parseFirewall(db, orgId, fireYaml)
			if err != nil || errData != nil {
				return
			}
		default:
			errData = &errortypes.ErrorData{
				Error:   "unit_kind_invalid",
				Message: "Unit kind is invalid",
			}
			return
		}
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
