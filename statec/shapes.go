package statec

import (
	"bytes"
	"sort"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/shape"
)

var (
	Shapes      = &ShapesState{}
	ShapesPkg   = NewPackage(Shapes)
	shapesCache = NewCache(
		"shapes",
		[]string{"type"},
		func() Scope {
			return Scope{
				Filter: bson.M{
					"type": shape.Disk,
				},
				Key: shape.Disk,
				Ok:  true,
			}
		},
		func(evt *Event) bool {
			typ, ok := evt.DocString("type")
			return ok && typ == shape.Disk
		},
	)
)

type ShapesState struct {
	diskShapes []*shape.Shape
}

func (p *ShapesState) DiskShapes() []*shape.Shape {
	return p.diskShapes
}

func (p *ShapesState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	err = shapesCache.Sync(db)
	if err != nil {
		return
	}

	shapes, err := Decode[shape.Shape](shapesCache)
	if err != nil {
		return
	}

	sort.SliceStable(shapes, func(i, j int) bool {
		if shapes[i].Name != shapes[j].Name {
			return shapes[i].Name < shapes[j].Name
		}
		return bytes.Compare(shapes[i].Id[:], shapes[j].Id[:]) < 0
	})

	p.diskShapes = shapes

	return
}

func (p *ShapesState) Apply(st *State) {
	st.DiskShapes = p.DiskShapes
}
