package service

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Service struct {
	Id               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name             string             `bson:"name" json:"name"`
	Comment          string             `bson:"comment" json:"comment"`
	Organization     primitive.ObjectID `bson:"organization" json:"organization"`
	DeleteProtection bool               `bson:"delete_protection" json:"delete_protection"`
	Units            []*Unit            `bson:"units" json:"units"`
}

func (p *Service) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	p.Name = utils.FilterName(p.Name)

	return
}

func (p *Service) InitUnits(units []*UnitInput) {
	p.Units = []*Unit{}

	for _, unitData := range units {
		p.Units = append(p.Units, &Unit{
			Id:   primitive.NewObjectID(),
			Name: unitData.Name,
			Spec: unitData.Spec,
		})
	}
}

func (p *Service) CommitFieldsUnits(db *database.Database,
	units []*UnitInput, fields set.Set) (
	errData *errortypes.ErrorData, err error) {

	arraySelect := database.NewArraySelectFields(p, "units", fields)

	curUnitsSet := set.NewSet()
	curUnitsMap := map[primitive.ObjectID]*Unit{}
	for _, unit := range p.Units {
		curUnitsSet.Add(unit.Id)
		curUnitsMap[unit.Id] = unit
	}

	unitsName := set.NewSet()
	newUnitsSet := set.NewSet()

	for _, unitData := range units {
		if unitsName.Contains(unitData.Name) {
			errData = &errortypes.ErrorData{
				Error:   "unit_duplicate_name",
				Message: "Duplicate unit name",
			}
		}
		unitsName.Add(unitData.Name)

		curUnit := curUnitsMap[unitData.Id]
		if curUnit == nil {
			unit := &Unit{
				Id:   primitive.NewObjectID(),
				Name: unitData.Name,
				Spec: unitData.Spec,
			}
			curUnitsSet.Add(unit.Id)
			curUnitsMap[unit.Id] = unit
			newUnitsSet.Add(unit.Id)
			p.Units = append(p.Units, unit)

			arraySelect.Push(unit)

			continue
		}

		newUnitsSet.Add(unitData.Id)
		curUnit.Name = unitData.Name
		curUnit.Spec = unitData.Spec

		arraySelect.Update(unitData.Id, bson.M{
			"name": unitData.Name,
			"spec": unitData.Spec,
		})
	}

	curUnitsSet.Subtract(newUnitsSet)
	for unitIdInf := range curUnitsSet.Iter() {
		arraySelect.Delete(unitIdInf.(primitive.ObjectID))
	}

	updateQuery, arrayFilters := arraySelect.GetQuery()

	coll := db.Services()

	updateOpts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: arrayFilters,
	})
	_, err = coll.UpdateOne(db, &bson.M{
		"_id": p.Id,
	}, updateQuery, updateOpts)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (p *Service) Commit(db *database.Database) (err error) {
	coll := db.Services()

	err = coll.Commit(p.Id, p)
	if err != nil {
		return
	}

	return
}

func (p *Service) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Services()

	err = coll.CommitFields(p.Id, p, fields)
	if err != nil {
		return
	}

	return
}

func (p *Service) Insert(db *database.Database) (err error) {
	coll := db.Services()

	_, err = coll.InsertOne(db, p)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
