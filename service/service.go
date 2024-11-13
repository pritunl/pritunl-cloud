package service

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
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

	if p.Organization.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "missing_organization",
			Message: "Missing organization",
		}
		return
	}

	return
}

func (p *Service) InitUnits(db *database.Database, units []*UnitInput) (
	errData *errortypes.ErrorData, err error) {

	p.Units = []*Unit{}

	for _, unitData := range units {
		unit := &Unit{
			Service:      p,
			Id:           primitive.NewObjectID(),
			Name:         unitData.Name,
			Spec:         unitData.Spec,
			DeployCommit: unitData.DeployCommit,
			Deployments:  []*Deployment{},
		}

		errData, err = unit.Parse(db)
		if err != nil {
			return
		}
		if errData != nil {
			return
		}

		p.Units = append(p.Units, unit)
	}

	return
}

func (p *Service) CommitFieldsUnits(db *database.Database,
	units []*UnitInput, fields set.Set) (
	errData *errortypes.ErrorData, err error) {

	arraySelectSet := database.NewArraySelectFields(p, "units", fields)
	arraySelectPush := database.NewArraySelectFields(p, "units", fields)
	arraySelectPull := database.NewArraySelectFields(p, "units", fields)

	curUnitsSet := set.NewSet()
	curUnitsMap := map[primitive.ObjectID]*Unit{}
	for _, unit := range p.Units {
		curUnitsSet.Add(unit.Id)
		unit.Service = p
		curUnitsMap[unit.Id] = unit
	}

	unitsName := set.NewSet()
	newUnitsSet := set.NewSet()

	for _, unitData := range units {
		if unitData.Delete {
			arraySelectPull.Delete(unitData.Id)
			continue
		}

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
				Service:      p,
				Id:           primitive.NewObjectID(),
				Name:         unitData.Name,
				Spec:         unitData.Spec,
				DeployCommit: unitData.DeployCommit,
				Deployments:  []*Deployment{},
			}
			curUnitsSet.Add(unit.Id)
			curUnitsMap[unit.Id] = unit
			newUnitsSet.Add(unit.Id)

			errData, err = unit.Parse(db)
			if err != nil {
				return
			}
			if errData != nil {
				return
			}

			p.Units = append(p.Units, unit)

			arraySelectPush.Push(unit)

			continue
		}

		newUnitsSet.Add(unitData.Id)
		curUnit.Name = unitData.Name
		curUnit.Spec = unitData.Spec
		curUnit.DeployCommit = unitData.DeployCommit

		errData, err = curUnit.Parse(db)
		if err != nil {
			return
		}
		if errData != nil {
			return
		}

		arraySelectSet.Update(unitData.Id, bson.M{
			"name":          curUnit.Name,
			"kind":          curUnit.Kind,
			"count":         curUnit.Count,
			"spec":          curUnit.Spec,
			"last_commit":   curUnit.LastCommit,
			"deploy_commit": curUnit.DeployCommit,
			"hash":          curUnit.Hash,
		})
	}

	coll := db.Services()

	if arraySelectPull.Modified() {
		updateQuery, arrayFilters := arraySelectPull.GetQuery()

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
	}

	if arraySelectPush.Modified() {
		updateQuery, arrayFilters := arraySelectPush.GetQuery()

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
	}

	if arraySelectSet.Modified() {
		updateQuery, arrayFilters := arraySelectSet.GetQuery()

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

func (p *Service) GetUnit(unitId primitive.ObjectID) *Unit {
	for _, unit := range p.Units {
		if unit.Id == unitId {
			unit.Service = p
			return unit
		}
	}
	return nil
}

func (p *Service) IterInstances() <-chan *Unit {
	iter := make(chan *Unit)

	go func() {
		defer close(iter)

		for _, unit := range p.Units {
			if unit.Kind != deployment.Instance &&
				unit.Kind != deployment.Image {

				continue
			}

			unit.Service = p
			iter <- unit
		}
	}()

	return iter
}
