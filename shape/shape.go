package shape

import (
	"path"
	"slices"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Shape struct {
	Id               bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name             string          `bson:"name" json:"name"`
	Comment          string          `bson:"comment" json:"comment"`
	Type             string          `bson:"type" json:"type"`
	DeleteProtection bool            `bson:"delete_protection" json:"delete_protection"`
	Roles            []string        `bson:"roles" json:"roles"`
	Flexible         bool            `bson:"flexible" json:"flexible"`
	DiskType         string          `bson:"disk_type" json:"disk_type"`
	DisksPath        string          `bson:"disks_path" json:"disks_path"`
	DiskPools        []bson.ObjectID `bson:"disk_pools" json:"disk_pools"`
	Memory           int             `bson:"memory" json:"memory"`
	Processors       int             `bson:"processors" json:"processors"`
	NodeCount        int             `bson:"-" json:"node_count"`
}

type Completion struct {
	Id         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string        `bson:"name" json:"name"`
	Type       string        `bson:"type" json:"type"`
	Roles      []string      `bson:"roles" json:"roles"`
	Flexible   bool          `bson:"flexible" json:"flexible"`
	DiskType   string        `bson:"disk_type" json:"disk_type"`
	Memory     int           `bson:"memory" json:"memory"`
	Processors int           `bson:"processors" json:"processors"`
}

func (s *Shape) GetType() string {
	if s.Type == "" {
		return Instance
	}
	return s.Type
}

func (s *Shape) HasPool(poolId bson.ObjectID) bool {
	return slices.Contains(s.DiskPools, poolId)
}

func (s *Shape) GetDiskPath(diskId bson.ObjectID) string {
	return paths.JoinDiskPath(s.DisksPath, diskId)
}

func (s *Shape) MatchRoles(roles []string) bool {
	for _, role := range s.Roles {
		if slices.Contains(roles, role) {
			return true
		}
	}
	return false
}

func (s *Shape) MatchNode(nde *node.Node) bool {
	if nde == nil || !s.MatchRoles(nde.Roles) {
		return false
	}

	if s.GetType() == Disk && s.DiskType == Lvm {
		for _, poolId := range s.DiskPools {
			if slices.Contains(nde.Pools, poolId) {
				return true
			}
		}
		return false
	}

	return true
}

func (s *Shape) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	s.Name = utils.FilterName(s.Name)

	if s.Type == "" {
		s.Type = Instance
	}

	if s.Roles == nil {
		s.Roles = []string{}
	}
	slices.Sort(s.Roles)

	if s.DiskPools == nil {
		s.DiskPools = []bson.ObjectID{}
	}

	switch s.Type {
	case Instance:
		s.DiskType = ""
		s.DisksPath = ""
		s.DiskPools = []bson.ObjectID{}

		if s.Memory < 256 {
			errData = &errortypes.ErrorData{
				Error:   "memory_invalid",
				Message: "Shape memory below minimum",
			}
			return
		}

		if s.Processors < 1 {
			errData = &errortypes.ErrorData{
				Error:   "processors_invalid",
				Message: "Shape processors below minimum",
			}
			return
		}

		break
	case Disk:
		s.Flexible = false
		s.Memory = 0
		s.Processors = 0

		s.DisksPath = strings.TrimSpace(s.DisksPath)

		diskPools := []bson.ObjectID{}
		for _, poolId := range s.DiskPools {
			if poolId.IsZero() || slices.Contains(diskPools, poolId) {
				continue
			}
			diskPools = append(diskPools, poolId)
		}
		s.DiskPools = diskPools

		if s.DiskType == "" {
			if s.DisksPath == "" && len(s.DiskPools) > 0 {
				s.DiskType = Lvm
			} else {
				s.DiskType = Qcow2
			}
		}

		switch s.DiskType {
		case Qcow2:
			s.DiskPools = []bson.ObjectID{}

			if s.DisksPath == "" {
				errData = &errortypes.ErrorData{
					Error:   "disks_path_required",
					Message: "Disk shape requires a disks path",
				}
				return
			}

			if !path.IsAbs(s.DisksPath) {
				errData = &errortypes.ErrorData{
					Error:   "disks_path_invalid",
					Message: "Disks path must be an absolute path",
				}
				return
			}
			s.DisksPath = path.Clean(s.DisksPath)

			break
		case Lvm:
			s.DisksPath = ""

			if len(s.DiskPools) == 0 {
				errData = &errortypes.ErrorData{
					Error:   "disk_pools_required",
					Message: "Disk shape requires a disk pool",
				}
				return
			}

			pools, e := pool.GetAll(db, &bson.M{
				"_id": &bson.M{
					"$in": s.DiskPools,
				},
			})
			if e != nil {
				err = e
				return
			}

			if len(pools) != len(s.DiskPools) {
				errData = &errortypes.ErrorData{
					Error:   "pool_not_found",
					Message: "Disk pool not found",
				}
				return
			}

			break
		default:
			errData = &errortypes.ErrorData{
				Error:   "invalid_disk_type",
				Message: "Disk type invalid",
			}
			return
		}

		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "invalid_shape_type",
			Message: "Shape type invalid",
		}
		return
	}

	return
}

func (s *Shape) FindNode(db *database.Database, zneId bson.ObjectID,
	dskShpe *Shape, processors, memory int) (nde *node.Node, err error) {

	ndes, err := node.GetAllShape(db, []bson.ObjectID{zneId}, s.Roles)
	if err != nil {
		return
	}

	Nodes(ndes).Sort()

	for _, nd := range ndes {
		if dskShpe != nil && !dskShpe.MatchNode(nd) {
			continue
		}

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
