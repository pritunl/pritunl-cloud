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
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/finder"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/zone"
	"gopkg.in/yaml.v2"
)

type Spec struct {
	Id           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Pod          primitive.ObjectID `bson:"pod" json:"pod"`
	Unit         primitive.ObjectID `bson:"unit" json:"unit"`
	Organization primitive.ObjectID `bson:"organization" json:"organization"`
	Index        int                `bson:"index" json:"index"`
	Timestamp    time.Time          `bson:"timestamp" json:"timestamp"`
	Name         string             `bson:"name" json:"name"`
	Kind         string             `bson:"kind" json:"kind"`
	Count        int                `bson:"count" json:"count"`
	Hash         string             `bson:"hash" json:"hash"`
	Data         string             `bson:"data" json:"data"`
	Instance     *Instance          `bson:"instance,omitempty" json:"-"`
	Firewall     *Firewall          `bson:"firewall,omitempty" json:"-"`
	Domain       *Domain            `bson:"domain,omitempty" json:"-"`
}

func (s *Spec) GetAllNodes(db *database.Database,
	orgId primitive.ObjectID) (ndes Nodes,
	offlineCount, noMountCount int, err error) {

	shpe, err := shape.Get(db, s.Instance.Shape)
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
	if s.Instance.Mounts != nil && len(s.Instance.Mounts) > 0 {
		diskIds := []primitive.ObjectID{}
		for _, mount := range s.Instance.Mounts {
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

		for _, mount := range s.Instance.Mounts {
			mountSet := set.NewSet()

			if mount.Disks != nil {
				for _, dskId := range mount.Disks {
					dsk := disksMap[dskId]
					if dsk == nil || !dsk.IsAvailable() {
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

func (s *Spec) Validate(db *database.Database) (err error) {
	if s.Timestamp.IsZero() {
		s.Timestamp = time.Now()
	}

	return
}

func (s *Spec) ExtractResources() (resources string, err error) {
	matches := resourcesRe.FindStringSubmatch(s.Data)
	if len(matches) > 1 {
		resources = matches[1]
		resources = strings.TrimSpace(resources)
		return
	}

	return
}

func (s *Spec) parseFirewall(db *database.Database,
	orgId primitive.ObjectID, dataYaml *FirewallYaml) (
	errData *errortypes.ErrorData, err error) {

	data := &Firewall{
		Ingress: []*Rule{},
	}

	if dataYaml.Kind != finder.FirewallKind {
		errData = &errortypes.ErrorData{
			Error:   "unit_kind_mismatch",
			Message: "Unit kind unexpected",
		}
		return
	}

	resources := &finder.Resources{
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

		refs := set.NewSet()
		for _, source := range ruleYaml.Source {
			if strings.HasPrefix(source, TokenPrefix) {
				kind, e := resources.Find(db, source)
				if e != nil {
					err = e
					return
				}

				if kind == finder.UnitKind && resources.Pod != nil &&
					resources.Unit != nil {

					selector := resources.Selector
					if selector == "" {
						selector = "private_ips"
					}

					refs.Add(Refrence{
						Id:       resources.Unit.Id,
						Realm:    resources.Pod.Id,
						Kind:     Unit,
						Selector: selector,
					})
				}
			} else {
				rule.SourceIps = append(rule.SourceIps, source)
			}
		}

		for refInf := range refs.Iter() {
			ref := refInf.(Refrence)
			rule.Sources = append(rule.Sources, &ref)
		}

		data.Ingress = append(data.Ingress, rule)
	}

	errData, err = data.Validate()
	if err != nil || errData != nil {
		return
	}

	s.Firewall = data

	return
}

func (s *Spec) parseInstance(db *database.Database,
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
	case finder.InstanceKind:
		break
	case finder.ImageKind:
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "unit_kind_mismatch",
			Message: "Unit kind unexpected",
		}
		return
	}

	resources := &finder.Resources{
		Organization: orgId,
	}

	if dataYaml.Plan != "" {
		kind, e := resources.Find(db, dataYaml.Plan)
		if e != nil {
			err = e
			return
		}
		if kind == finder.PlanKind && resources.Plan != nil {
			data.Plan = resources.Plan.Id
		}
	}

	if dataYaml.Zone != "" {
		kind, e := resources.Find(db, dataYaml.Zone)
		if e != nil {
			err = e
			return
		}
		if kind == finder.ZoneKind && resources.Zone != nil {
			data.Datacenter = resources.Datacenter.Id
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
		if kind == finder.NodeKind && resources.Node != nil {
			data.Node = resources.Node.Id
		}
	}
	if dataYaml.Shape != "" {
		kind, e := resources.Find(db, dataYaml.Shape)
		if e != nil {
			err = e
			return
		}
		if kind == finder.ShapeKind && resources.Shape != nil {
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
		if kind == finder.VpcKind && resources.Vpc != nil {
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
		if kind == finder.SubnetKind && resources.Subnet != nil {
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

		if kind == finder.ImageKind && resources.Image != nil {
			data.Image = resources.Image.Id
		}

		if kind == finder.BuildKind && resources.Deployment != nil &&
			resources.Deployment.ImageReady() {

			data.Image = resources.Deployment.Image
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
				if kind == finder.DiskKind && resources.Disks != nil {
					for _, dskRes := range resources.Disks {
						mnt.Disks = append(mnt.Disks, dskRes.Id)
					}
				}
			}

			data.Mounts = append(data.Mounts, mnt)
		}
	}

	if dataYaml.NodePorts != nil {
		externalNodePorts := set.NewSet()
		for _, nodePrt := range dataYaml.NodePorts {
			mapping := NodePort{
				Protocol:     nodePrt.Protocol,
				ExternalPort: nodePrt.ExternalPort,
				InternalPort: nodePrt.InternalPort,
			}

			extPortKey := fmt.Sprintf("%s:%d",
				mapping.Protocol, mapping.ExternalPort)

			if externalNodePorts.Contains(extPortKey) {
				errData = &errortypes.ErrorData{
					Error:   "node_port_external_duplicate",
					Message: "Duplicate external node port",
				}
				return
			}
			externalNodePorts.Add(extPortKey)

			errData, err = mapping.Validate()
			if err != nil || errData != nil {
				return
			}

			data.NodePorts = append(data.NodePorts)
		}
	}

	if dataYaml.Certificates != nil {
		for _, cert := range dataYaml.Certificates {
			kind, e := resources.Find(db, cert)
			if e != nil {
				err = e
				return
			}
			if kind == finder.CertificateKind && resources.Certificate != nil {
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
			if kind == finder.SecretKind && resources.Secret != nil {
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
			if kind == finder.PodKind && resources.Pod != nil {
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

	data.Uefi = dataYaml.Uefi
	data.SecureBoot = dataYaml.SecureBoot

	switch dataYaml.CloudType {
	case instance.Linux:
		data.CloudType = instance.Linux
		break
	case instance.BSD:
		data.CloudType = instance.BSD
		break
	case "":
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "invalid_unit_cloud_type",
			Message: "Unit instance cloud type is invalid",
		}
		return
	}

	data.Tpm = dataYaml.Tpm
	data.Vnc = dataYaml.Vnc
	data.DeleteProtection = dataYaml.DeleteProtection
	data.SkipSourceDestCheck = dataYaml.SkipSourceDestCheck
	data.HostAddress = dataYaml.HostAddress
	data.PublicAddress = dataYaml.PublicAddress
	data.PublicAddress6 = dataYaml.PublicAddress6
	data.DhcpServer = dataYaml.DhcpServer

	data.Roles = dataYaml.Roles
	data.DiskSize = dataYaml.DiskSize

	s.Name = dataYaml.Name
	s.Kind = dataYaml.Kind
	s.Count = dataYaml.Count
	s.Instance = data

	s.Count = dataYaml.Count
	if s.Kind == finder.ImageKind && s.Count != 0 {
		errData = &errortypes.ErrorData{
			Error:   "count_invalid",
			Message: "Count not valid for image kind",
		}
		return
	}

	return
}

func (s *Spec) parseDomain(db *database.Database,
	orgId primitive.ObjectID, dataYaml *DomainYaml) (
	errData *errortypes.ErrorData, err error) {

	data := &Domain{
		Records: []*Record{},
	}

	if dataYaml.Kind != finder.DomainKind {
		errData = &errortypes.ErrorData{
			Error:   "unit_kind_mismatch",
			Message: "Unit kind unexpected",
		}
		return
	}

	resources := &finder.Resources{
		Organization: orgId,
	}

	for _, recordYaml := range dataYaml.Records {
		if recordYaml.Name == "" || recordYaml.Type == "" {
			continue
		}

		record := &Record{
			Name: utils.FilterName(recordYaml.Name),
			Type: recordYaml.Type,
		}

		kind, e := resources.Find(db, recordYaml.Domain)
		if e != nil {
			err = e
			return
		}

		if kind == finder.DomainKind && resources.Domain != nil {
			record.Domain = resources.Domain.Id
		}

		data.Records = append(data.Records, record)
	}

	errData, err = data.Validate()
	if err != nil || errData != nil {
		return
	}

	s.Domain = data

	return
}

func (s *Spec) Refresh(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	errData, err = s.Parse(db)
	if err != nil || errData != nil {
		return
	}

	err = s.CommitData(db)
	if err != nil {
		return
	}

	return
}

func (s *Spec) Parse(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	hash := sha1.New()
	hash.Write([]byte(filterSpecHash(s.Data)))
	hashBytes := hash.Sum(nil)
	s.Hash = fmt.Sprintf("%x", hashBytes)

	resourcesSpec, err := s.ExtractResources()
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
		case finder.InstanceKind, finder.ImageKind:
			instYaml := &InstanceYaml{}

			err = decoder.Decode(instYaml)
			if err != nil {
				err = &errortypes.ParseError{
					errors.Wrap(err,
						"spec: Failed to decode instance yaml doc"),
				}
				return
			}

			errData, err = s.parseInstance(db, s.Organization, instYaml)
			if err != nil || errData != nil {
				return
			}
		case finder.FirewallKind:
			fireYaml := &FirewallYaml{}

			err = decoder.Decode(fireYaml)
			if err != nil {
				err = &errortypes.ParseError{
					errors.Wrap(err,
						"spec: Failed to decode firewall yaml doc"),
				}
				return
			}

			errData, err = s.parseFirewall(db, s.Organization, fireYaml)
			if err != nil || errData != nil {
				return
			}
		case finder.DomainKind:
			domnYaml := &DomainYaml{}

			err = decoder.Decode(domnYaml)
			if err != nil {
				err = &errortypes.ParseError{
					errors.Wrap(err,
						"spec: Failed to decode domain yaml doc"),
				}
				return
			}

			errData, err = s.parseDomain(db, s.Organization, domnYaml)
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

func (s *Spec) CanMigrate(db *database.Database, spc *Spec) (
	errData *errortypes.ErrorData, err error) {

	if !settings.System.NoMigrateRefresh {
		errData, err = s.Parse(db)
		if err != nil || errData != nil {
			return
		}
	}

	errData, err = spc.Parse(db)
	if err != nil || errData != nil {
		return
	}

	if s.Pod != spc.Pod || s.Unit != spc.Unit {
		err = &errortypes.ParseError{
			errors.Newf("spec: Invalid unit"),
		}
		return
	}

	if s.Kind != spc.Kind {
		errData = &errortypes.ErrorData{
			Error:   "unit_kind_conflict",
			Message: "Cannot migrate to different kind",
		}
		return
	}

	if s.Instance == nil || spc.Instance == nil {
		err = &errortypes.ParseError{
			errors.Newf("spec: Instance not found"),
		}
		return
	}

	if s.Instance.Datacenter != spc.Instance.Datacenter {
		errData = &errortypes.ErrorData{
			Error:   "instance_datacenter_conflict",
			Message: "Cannot migrate to different instance datacenter",
		}
		return
	}

	if s.Instance.Zone != spc.Instance.Zone {
		errData = &errortypes.ErrorData{
			Error:   "instance_zone_conflict",
			Message: "Cannot migrate to different instance zone",
		}
		return
	}

	if s.Instance.Node != spc.Instance.Node {
		errData = &errortypes.ErrorData{
			Error:   "instance_node_coflict",
			Message: "Cannot migrate to different instance node",
		}
		return
	}

	if s.Instance.Shape != spc.Instance.Shape {
		errData = &errortypes.ErrorData{
			Error:   "instance_shape_coflict",
			Message: "Cannot migrate to different instance shape",
		}
		return
	}

	if s.Instance.Subnet != spc.Instance.Subnet {
		errData = &errortypes.ErrorData{
			Error:   "instance_subnet_coflict",
			Message: "Cannot migrate to different instance subnet",
		}
		return
	}

	if s.Instance.Image != spc.Instance.Image {
		errData = &errortypes.ErrorData{
			Error:   "instance_image_coflict",
			Message: "Cannot migrate to different instance image",
		}
		return
	}

	if s.Instance.DiskSize != spc.Instance.DiskSize {
		errData = &errortypes.ErrorData{
			Error:   "instance_disk_size_coflict",
			Message: "Cannot migrate to different instance disk size",
		}
		return
	}

	curMountPaths := set.NewSet()
	curMountDisks := map[string]set.Set{}
	for _, mnt := range s.Instance.Mounts {
		curMountPaths.Add(mnt.Path)
		mntDisks := set.NewSet()
		for _, dskId := range mnt.Disks {
			mntDisks.Add(dskId)
		}
		curMountDisks[mnt.Path] = mntDisks
	}

	newMountPaths := set.NewSet()
	for _, mnt := range spc.Instance.Mounts {
		newMountPaths.Add(mnt.Path)
		curMntDisks := curMountDisks[mnt.Path]
		if curMntDisks == nil {
			errData = &errortypes.ErrorData{
				Error:   "instance_mount_coflict",
				Message: "Cannot migrate to different instance mounts",
			}
			return
		}

		for _, dskId := range mnt.Disks {
			if !curMntDisks.Contains(dskId) {
				errData = &errortypes.ErrorData{
					Error:   "instance_mount_disk_coflict",
					Message: "Cannot migrate to instance with fewer disks",
				}
				return
			}
		}
	}

	if !curMountPaths.IsEqual(newMountPaths) {
		errData = &errortypes.ErrorData{
			Error:   "instance_mount_coflict",
			Message: "Cannot migrate to different instance mounts",
		}
		return
	}

	return
}

func (s *Spec) Commit(db *database.Database) (err error) {
	coll := db.Specs()

	err = coll.Commit(s.Id, s)
	if err != nil {
		return
	}

	return
}

func (s *Spec) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Specs()

	err = coll.CommitFields(s.Id, s, fields)
	if err != nil {
		return
	}

	return
}

func (s *Spec) CommitData(db *database.Database) (err error) {
	coll := db.Specs()

	err = coll.CommitFields(s.Id, s, set.NewSet(
		"name", "count", "data", "instance", "firewall", "domain"))
	if err != nil {
		return
	}

	return
}

func (s *Spec) Insert(db *database.Database) (err error) {
	coll := db.Specs()

	resp, err := coll.InsertOne(db, s)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	s.Id = resp.InsertedID.(primitive.ObjectID)

	return
}
