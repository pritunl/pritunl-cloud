package disk

import (
	"path"
	"slices"
	"sort"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/utils"
)

type location struct {
	path  string
	shape bson.ObjectID
}

func defaultDisksPath(nde *node.Node) string {
	return path.Join(nde.GetVirtPath(), "disks")
}

func (d *Disk) locations(nde *node.Node, shapes []*shape.Shape) (
	available []location, others []location) {

	available = []location{}
	others = []location{}
	seen := map[string]bool{}

	for _, shpe := range shapes {
		if shpe.DiskType != shape.Qcow2 || !shpe.MatchRoles(nde.Roles) {
			continue
		}

		seen[shpe.DisksPath] = true
		available = append(available, location{
			path:  shpe.GetDiskPath(d.Id),
			shape: shpe.Id,
		})
	}

	defaultPath := defaultDisksPath(nde)
	if !seen[defaultPath] {
		seen[defaultPath] = true
		others = append(others, location{
			path: paths.JoinDiskPath(defaultPath, d.Id),
		})
	}

	for _, shpe := range shapes {
		if shpe.DiskType != shape.Qcow2 || seen[shpe.DisksPath] {
			continue
		}

		seen[shpe.DisksPath] = true
		others = append(others, location{
			path: shpe.GetDiskPath(d.Id),
		})
	}

	return
}

func (d *Disk) Locate(nde *node.Node, shapes []*shape.Shape) (
	pth string, shapeId bson.ObjectID, exists bool, err error) {

	available, others := d.locations(nde, shapes)

	if !d.Shape.IsZero() {
		for _, loc := range available {
			if loc.shape != d.Shape {
				continue
			}

			exists, err = utils.ExistsFile(loc.path)
			if err != nil {
				return
			}

			if exists {
				pth = loc.path
				shapeId = loc.shape
				return
			}
			break
		}
	}

	for _, loc := range append(available, others...) {
		exists, err = utils.ExistsFile(loc.path)
		if err != nil {
			return
		}

		if exists {
			pth = loc.path
			shapeId = loc.shape
			return
		}
	}

	if !d.Shape.IsZero() {
		for _, loc := range available {
			if loc.shape == d.Shape {
				pth = loc.path
				shapeId = loc.shape
				return
			}
		}
	}

	pth = paths.JoinDiskPath(defaultDisksPath(nde), d.Id)
	for _, loc := range available {
		if loc.path == pth {
			shapeId = loc.shape
			break
		}
	}

	return
}

func (d *Disk) GetPath(db *database.Database) (pth string, err error) {
	shapes, err := shape.GetAllDisks(db)
	if err != nil {
		return
	}

	pth, _, _, err = d.Locate(node.Self, shapes)
	if err != nil {
		return
	}

	return
}

func (d *Disk) FindShape(nde *node.Node, shapes []*shape.Shape) (
	shapeId bson.ObjectID, found bool, err error) {

	if d.Type == Lvm {
		found = true

		if !d.Shape.IsZero() {
			for _, shpe := range shapes {
				if shpe.Id == d.Shape && shpe.HasPool(d.Pool) {
					shapeId = shpe.Id
					return
				}
			}
		}

		for _, shpe := range shapes {
			if shpe.HasPool(d.Pool) {
				shapeId = shpe.Id
				return
			}
		}

		return
	}

	_, shapeId, found, err = d.Locate(nde, shapes)
	if err != nil {
		return
	}

	return
}

func (d *Disk) ApplyShape(db *database.Database, shpe *shape.Shape,
	nde *node.Node, zneId bson.ObjectID) (
	errData *errortypes.ErrorData, err error) {

	if shpe.GetType() != shape.Disk {
		errData = &errortypes.ErrorData{
			Error:   "invalid_disk_shape",
			Message: "Shape is not a disk shape",
		}
		return
	}

	switch shpe.DiskType {
	case shape.Qcow2:
		if nde == nil {
			errData = &errortypes.ErrorData{
				Error:   "node_required",
				Message: "Missing required node",
			}
			return
		}

		if !shpe.MatchRoles(nde.Roles) {
			errData = &errortypes.ErrorData{
				Error:   "disk_shape_unavailable",
				Message: "Disk shape not available on node",
			}
			return
		}

		d.Type = Qcow2
		d.Node = nde.Id
		d.Pool = bson.NilObjectID
		d.Datacenter = nde.Datacenter
		d.Zone = nde.Zone
		d.Shape = shpe.Id

		break
	case shape.Lvm:
		if nde != nil {
			if !shpe.MatchRoles(nde.Roles) {
				errData = &errortypes.ErrorData{
					Error:   "disk_shape_unavailable",
					Message: "Disk shape not available on node",
				}
				return
			}

			zneId = nde.Zone
		}

		if zneId.IsZero() {
			errData = &errortypes.ErrorData{
				Error:   "zone_required",
				Message: "Missing required zone",
			}
			return
		}

		pools, e := pool.GetAll(db, &bson.M{
			"_id": &bson.M{
				"$in": shpe.DiskPools,
			},
			"zone": zneId,
		})
		if e != nil {
			err = e
			return
		}

		sort.Slice(pools, func(i, j int) bool {
			if pools[i].Name == pools[j].Name {
				return pools[i].Id.Hex() < pools[j].Id.Hex()
			}
			return pools[i].Name < pools[j].Name
		})

		var pl *pool.Pool
		for _, p := range pools {
			if nde != nil && !slices.Contains(nde.Pools, p.Id) {
				continue
			}

			pl = p
			break
		}

		if pl == nil {
			if nde != nil {
				errData = &errortypes.ErrorData{
					Error:   "pool_unavailable",
					Message: "Disk shape pool not available on node",
				}
			} else {
				errData = &errortypes.ErrorData{
					Error:   "pool_not_found",
					Message: "Disk shape pool not available in zone",
				}
			}
			return
		}

		d.Type = Lvm
		d.Node = bson.NilObjectID
		d.Pool = pl.Id
		d.Datacenter = pl.Datacenter
		d.Zone = pl.Zone
		d.Shape = shpe.Id

		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "invalid_disk_shape",
			Message: "Disk shape has no disks path or disk pool",
		}
		return
	}

	return
}

func (d *Disk) ApplyShapeId(db *database.Database, shapeId bson.ObjectID,
	nde *node.Node, zneId bson.ObjectID) (
	errData *errortypes.ErrorData, err error) {

	if shapeId.IsZero() {
		if nde == nil {
			errData = &errortypes.ErrorData{
				Error:   "node_required",
				Message: "Missing required node",
			}
			return
		}

		shapes, e := shape.GetAllDisks(db)
		if e != nil {
			err = e
			return
		}

		for _, shpe := range shapes {
			if shpe.DiskType != shape.Qcow2 ||
				!shpe.MatchRoles(nde.Roles) {

				continue
			}

			errData, err = d.ApplyShape(db, shpe, nde, zneId)
			return
		}

		d.Type = Qcow2
		d.Node = nde.Id
		d.Pool = bson.NilObjectID
		d.Datacenter = nde.Datacenter
		d.Zone = nde.Zone
		d.Shape = bson.NilObjectID

		return
	}

	shpe, err := shape.Get(db, shapeId)
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
			errData = &errortypes.ErrorData{
				Error:   "disk_shape_not_found",
				Message: "Disk shape not found",
			}
		}
		return
	}

	errData, err = d.ApplyShape(db, shpe, nde, zneId)
	if err != nil {
		return
	}

	return
}
