package pod

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/nodeport"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/tools/errors"
)

type Unit struct {
	Pod           *Pod                `bson:"-" json:"-"`
	Id            primitive.ObjectID  `bson:"id" json:"id"`
	Name          string              `bson:"name" json:"name"`
	Kind          string              `bson:"kind" json:"kind"`
	Count         int                 `bson:"count" json:"count"`
	Deployments   []*Deployment       `bson:"deployments" json:"deployments"`
	Spec          string              `bson:"spec" json:"spec"`
	SpecIndex     int                 `bson:"spec_index" json:"spec_index"`
	SpecTimestamp time.Time           `bson:"spec_timestamp" json:"-"`
	LastCommit    primitive.ObjectID  `bson:"last_commit" json:"last_commit"`
	DeployCommit  primitive.ObjectID  `bson:"deploy_commit" json:"deploy_commit"`
	Hash          string              `bson:"hash" json:"hash"`
	NodePorts     []*nodeport.Mapping `bson:"node_ports" json:"node_ports"`
}

type UnitInput struct {
	Id           primitive.ObjectID `json:"id"`
	Name         string             `json:"name"`
	Spec         string             `json:"spec"`
	DeployCommit primitive.ObjectID `json:"deploy_commit"`
	Delete       bool               `json:"delete"`
}

type Deployment struct {
	Id primitive.ObjectID `bson:"id" json:"id"`
}

type NodePortMapping struct {
	NodePort     primitive.ObjectID `bson:"node_port" json:"node_port"`
	InternalPort int                `bson:"internal_port" json:"internal_port"`
}

func (u *Unit) Refresh(db *database.Database) (err error) {
	coll := db.Pods()

	pd := &Pod{}
	err = coll.FindOne(db, &bson.M{
		"_id":      u.Pod.Id,
		"units.id": u.Id,
	}, database.FindOneProject(
		"units.$",
	)).Decode(pd)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	for _, unit := range pd.Units {
		if unit.Id != u.Id {
			continue
		}

		unit.Pod = u.Pod
		*u = *unit
		return
	}

	err = &errortypes.DatabaseError{
		errors.New("pod: Unit refresh failed to find unit"),
	}
	return
}

func (u *Unit) HasDeployment(deployId primitive.ObjectID) bool {
	if u.Deployments != nil {
		for _, deply := range u.Deployments {
			if deply.Id == deployId {
				return true
			}
		}
	}

	return false
}

func (u *Unit) Reserve(db *database.Database, deployId primitive.ObjectID,
	overrideCount int) (reserved bool, err error) {

	coll := db.Pods()

	if overrideCount == 0 {
		if len(u.Deployments) >= u.Count {
			return
		}
	} else {
		if len(u.Deployments) >= overrideCount {
			return
		}
	}

	updateOpts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{
				"elem.id":    u.Id,
				"elem.count": u.Count,
				"elem.deployments": bson.M{
					"$size": len(u.Deployments),
				},
			},
		},
	})
	resp, err := coll.UpdateOne(db, bson.M{
		"_id": u.Pod.Id,
	}, bson.M{
		"$push": bson.M{
			"units.$[elem].deployments": &Deployment{
				Id: deployId,
			},
		},
	}, updateOpts)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if resp.MatchedCount == 1 && resp.ModifiedCount == 1 {
		reserved = true
	}

	return
}

func (u *Unit) UpdateDeployementOld(db *database.Database,
	deploymentId primitive.ObjectID, state string) (updated bool, err error) {

	coll := db.Pods()

	updateOpts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"elem.id": u.Id},
			bson.M{"deploy.id": deploymentId},
		},
	})
	resp, err := coll.UpdateOne(db, bson.M{
		"_id": u.Pod.Id,
	}, bson.M{
		"$set": bson.M{
			"units.$[elem].deployments.$[deploy].state": state,
		},
	}, updateOpts)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if resp.MatchedCount == 1 && resp.ModifiedCount == 1 {
		updated = true
	}

	return
}

func (u *Unit) RestoreDeployment(db *database.Database,
	deployId primitive.ObjectID) (err error) {

	coll := db.Pods()

	updateOpts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{
				"elem.id": u.Id,
			},
		},
	})
	_, err = coll.UpdateOne(db, bson.M{
		"_id": u.Pod.Id,
	}, bson.M{
		"$push": bson.M{
			"units.$[elem].deployments": &Deployment{
				Id: deployId,
			},
		},
	}, updateOpts)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (u *Unit) RemoveDeployement(db *database.Database,
	deployId primitive.ObjectID) (err error) {

	coll := db.Pods()

	updateOpts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"elem.id": u.Id},
		},
	})
	_, err = coll.UpdateOne(db, bson.M{
		"_id": u.Pod.Id,
	}, bson.M{
		"$pull": bson.M{
			"units.$[elem].deployments": &bson.M{
				"id": deployId,
			},
		},
	}, updateOpts)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (u *Unit) MigrateDeployements(db *database.Database,
	newSpecId primitive.ObjectID, deplyIds []primitive.ObjectID) (
	errData *errortypes.ErrorData, err error) {

	coll := db.Deployments()

	newSpc, err := spec.Get(db, newSpecId)
	if err != nil {
		return
	}

	if newSpc.Pod != u.Pod.Id || newSpc.Unit != u.Id {
		err = &errortypes.ParseError{
			errors.Newf("spec: Invalid unit"),
		}
		return
	}

	deplys, err := deployment.GetAll(db, &bson.M{
		"_id": &bson.M{
			"$in": deplyIds,
		},
		"pod":  u.Pod.Id,
		"unit": u.Id,
	})
	if err != nil {
		return
	}

	spcMap := map[primitive.ObjectID]*spec.Spec{}

	for _, deply := range deplys {
		oldSpc := spcMap[deply.Spec]
		if oldSpc == nil {
			oldSpc, err = spec.Get(db, deply.Spec)
			if err != nil {
				return
			}

			spcMap[oldSpc.Id] = oldSpc
		}

		errData, err = oldSpc.CanMigrate(db, newSpc)
		if err != nil || errData != nil {
			return
		}
	}

	_, err = coll.UpdateMany(db, &bson.M{
		"_id": &bson.M{
			"$in": deplyIds,
		},
		"pod":  u.Pod.Id,
		"unit": u.Id,
	}, &bson.M{
		"$set": &bson.M{
			"action":   deployment.Migrate,
			"new_spec": newSpc.Id,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (u *Unit) CommitFields(db *database.Database,
	fields set.Set) (err error) {

	coll := db.Pods()

	update := database.SelectFieldsAll(u, fields)

	if setDoc, ok := update["$set"].(bson.M); ok {
		newSetDoc := bson.M{}
		for key, val := range setDoc {
			newSetDoc["units.$[elem]."+key] = val
		}
		update["$set"] = newSetDoc
	}
	if unsetDoc, ok := update["$unset"].(bson.M); ok {
		newUnsetDoc := bson.M{}
		for key, val := range unsetDoc {
			newUnsetDoc["units.$[elem]."+key] = val
		}
		update["$unset"] = newUnsetDoc
	}

	updateOpts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{
				"elem.id": u.Id,
			},
		},
	})
	_, err = coll.UpdateOne(db, bson.M{
		"_id": u.Pod.Id,
	}, update, updateOpts)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (u *Unit) Parse(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	spc := spec.New(u.Pod.Id, u.Id, u.Pod.Organization, u.Spec)

	errData, err = spc.Parse(db)
	if err != nil {
		return
	}
	if errData != nil {
		return
	}

	if u.Hash != spc.Hash {
		u.Name = spc.Name
		u.Count = spc.Count
		u.Spec = spc.Data

		if u.Kind == "" {
			u.Kind = spc.Kind
		} else if u.Kind != spc.Kind {
			errData = &errortypes.ErrorData{
				Error:   "spec_kind_invalid",
				Message: "Cannot change spec kind",
			}
			return
		}

		timestamp, index, e := NewSpec(db, u.Pod.Id, u.Id)
		if e != nil {
			err = e
			return
		}

		spc.Index = index
		spc.Timestamp = timestamp

		err = spc.Insert(db)
		if err != nil {
			return
		}

		u.Hash = spc.Hash
		u.LastCommit = spc.Id
		if u.DeployCommit.IsZero() {
			u.DeployCommit = spc.Id
		}
	} else if u.Name != spc.Name || u.Count != spc.Count {
		curSpc, e := spec.Get(db, u.LastCommit)
		if e != nil {
			err = e
			return
		}

		curSpc.Name = spc.Name
		curSpc.Count = spc.Count
		curSpc.Data = spc.Data

		err = curSpc.CommitData(db)
		if err != nil {
			return
		}

		u.Name = curSpc.Name
		u.Count = curSpc.Count
		u.Spec = curSpc.Data

		if u.Kind == "" {
			u.Kind = spc.Kind
		} else if u.Kind != spc.Kind {
			errData = &errortypes.ErrorData{
				Error:   "spec_kind_invalid",
				Message: "Cannot change spec kind",
			}
			return
		}

		u.Hash = curSpc.Hash
		u.LastCommit = curSpc.Id
		if u.DeployCommit.IsZero() {
			u.DeployCommit = curSpc.Id
		}
	}

	return
}
