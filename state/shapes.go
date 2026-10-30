package state

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/shape"
)

var (
	Shapes    = &ShapesState{}
	ShapesPkg = NewPackage(Shapes)
)

type ShapesState struct {
	diskShapes []*shape.Shape
}

func (p *ShapesState) DiskShapes() []*shape.Shape {
	return p.diskShapes
}

func (p *ShapesState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	shapes, err := shape.GetAllDisks(db)
	if err != nil {
		return
	}
	p.diskShapes = shapes

	return
}

func (p *ShapesState) Apply(st *State) {
	st.DiskShapes = p.DiskShapes
}
