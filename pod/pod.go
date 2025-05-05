package pod

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/unit"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Pod struct {
	Id               primitive.ObjectID                  `bson:"_id,omitempty" json:"id"`
	Name             string                              `bson:"name" json:"name"`
	Comment          string                              `bson:"comment" json:"comment"`
	Organization     primitive.ObjectID                  `bson:"organization" json:"organization"`
	DeleteProtection bool                                `bson:"delete_protection" json:"delete_protection"`
	UserDrafts       map[primitive.ObjectID][]*UnitDraft `bson:"drafts" json:"-"`
	Drafts           []*UnitDraft                        `bson:"-" json:"drafts"`
}

type UnitDraft struct {
	Id        primitive.ObjectID `bson:"id" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Spec      string             `bson:"spec" json:"spec"`
	Delete    bool               `bson:"delete" json:"delete"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
	New       bool               `bson:"new" json:"new"`
}

func (p *Pod) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	p.Name = utils.FilterName(p.Name)

	if p.Organization.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "missing_organization",
			Message: "Missing organization",
		}
		return
	}

	if p.UserDrafts == nil {
		p.UserDrafts = map[primitive.ObjectID][]*UnitDraft{}
	}

	return
}

func (p *Pod) Json(usrId primitive.ObjectID) {
	if p.UserDrafts != nil && p.UserDrafts[usrId] != nil {
		p.Drafts = p.UserDrafts[usrId]
	} else {
		p.Drafts = []*UnitDraft{}
	}
}

func (p *Pod) InitUnits(db *database.Database, units []*unit.UnitInput) (
	errData *errortypes.ErrorData, err error) {

	for _, unitData := range units {
		if unitData.Delete {
			continue
		}

		unt := &unit.Unit{
			Id:           primitive.NewObjectID(),
			Pod:          p.Id,
			Organization: p.Organization,
			Name:         unitData.Name,
			Spec:         unitData.Spec,
			SpecIndex:    1,
			Deployments:  []primitive.ObjectID{},
		}

		errData, err = unt.Parse(db, true)
		if err != nil {
			return
		}
		if errData != nil {
			return
		}

		err = unt.Insert(db)
		if err != nil {
			return
		}
	}

	return
}

func (p *Pod) CommitFieldsUnits(db *database.Database,
	units []*unit.UnitInput, fields set.Set) (
	errData *errortypes.ErrorData, err error) {

	curUnitsMap, err := unit.GetAllMap(db, &bson.M{
		"pod": p.Id,
	})
	if err != nil {
		return
	}

	unitsName := set.NewSet()
	for _, unitData := range units {
		if !unitData.Delete {
			if unitsName.Contains(unitData.Name) {
				errData = &errortypes.ErrorData{
					Error:   "unit_duplicate_name",
					Message: "Duplicate unit name",
				}
				return
			}
			unitsName.Add(unitData.Name)
		}

		if unitData.Delete {
			if false {
				errData = &errortypes.ErrorData{
					Error:   "unit_delete_active_deployments",
					Message: "Cannot delete unit with active deployments",
				}
				return
			}
		}
	}

	for _, unitData := range units {
		if unitData.Delete {
			continue
		}

		curUnit := curUnitsMap[unitData.Id]
		if curUnit == nil {
			continue
		}

		updateFields := set.NewSet(
			"name",
			"kind",
			"count",
			"spec",
			"last_spec",
			"deploy_spec",
			"hash",
		)

		curUnit.Name = unitData.Name
		curUnit.Spec = unitData.Spec

		if !unitData.DeploySpec.IsZero() {
			deploySpec, e := spec.Get(db, unitData.DeploySpec)
			if e != nil || deploySpec.Unit != curUnit.Id {
				errData = &errortypes.ErrorData{
					Error:   "unit_deploy_spec_invalid",
					Message: "Invalid unit deployment commit",
				}
				return
			}

			curUnit.DeploySpec = deploySpec.Id
		}

		errData, err = curUnit.Parse(db, false)
		if err != nil {
			return
		}
		if errData != nil {
			return
		}

		err = curUnit.CommitFields(db, updateFields)
		if err != nil {
			return
		}
	}

	for _, unitData := range units {
		if unitData.Delete {
			err = unit.Remove(db, unitData.Id)
			if err != nil {
				return
			}
			continue
		}

		curUnit := curUnitsMap[unitData.Id]
		if curUnit != nil {
			continue
		}

		unt := &unit.Unit{
			Id:           primitive.NewObjectID(),
			Pod:          p.Id,
			Organization: p.Organization,
			Name:         unitData.Name,
			Spec:         unitData.Spec,
			SpecIndex:    1,
			Deployments:  []primitive.ObjectID{},
		}

		errData, err = unt.Parse(db, true)
		if err != nil {
			return
		}
		if errData != nil {
			return
		}

		err = unt.Insert(db)
		if err != nil {
			return
		}
	}

	err = p.CommitFields(db, fields)
	if err != nil {
		return
	}

	return
}

func (p *Pod) Commit(db *database.Database) (err error) {
	coll := db.Pods()

	err = coll.Commit(p.Id, p)
	if err != nil {
		return
	}

	return
}

func (p *Pod) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Pods()

	err = coll.CommitFields(p.Id, p, fields)
	if err != nil {
		return
	}

	return
}

func (p *Pod) Insert(db *database.Database) (err error) {
	coll := db.Pods()

	_, err = coll.InsertOne(db, p)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
