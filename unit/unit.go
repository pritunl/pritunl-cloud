package unit

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/nodeport"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/tools/errors"
)

type Unit struct {
	Id            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Pod           primitive.ObjectID   `bson:"pod" json:"pod"`
	Organization  primitive.ObjectID   `bson:"organization" json:"organization"`
	Name          string               `bson:"name" json:"name"`
	Kind          string               `bson:"kind" json:"kind"`
	Count         int                  `bson:"count" json:"count"`
	Deployments   []primitive.ObjectID `bson:"deployments" json:"deployments"`
	Spec          string               `bson:"spec" json:"spec"`
	SpecIndex     int                  `bson:"spec_index" json:"spec_index"`
	SpecTimestamp time.Time            `bson:"spec_timestamp" json:"-"`
	LastCommit    primitive.ObjectID   `bson:"last_commit" json:"last_commit"`
	DeployCommit  primitive.ObjectID   `bson:"deploy_commit" json:"deploy_commit"`
	Hash          string               `bson:"hash" json:"hash"`
	NodePorts     []*nodeport.Mapping  `bson:"node_ports" json:"node_ports"`
}

type UnitDraft struct {
	Id        primitive.ObjectID `bson:"id" json:"id"`
	Pod       primitive.ObjectID `bson:"pod" json:"pod"`
	User      primitive.ObjectID `bson:"user" json:"user"`
	Name      string             `bson:"name" json:"name"`
	Kind      string             `bson:"kind" json:"kind"`
	Spec      string             `bson:"spec" json:"spec"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
}

type UnitInput struct {
	Id           primitive.ObjectID `json:"id"`
	Name         string             `json:"name"`
	Spec         string             `json:"spec"`
	DeployCommit primitive.ObjectID `json:"deploy_commit"`
	Delete       bool               `json:"delete"`
}

func (u *Unit) Refresh(db *database.Database) (err error) {
	coll := db.Units()

	unt := &Unit{}
	err = coll.FindOne(db, &bson.M{
		"_id": u.Id,
	}).Decode(unt)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	*u = *unt
	return
}

func (u *Unit) HasDeployment(deployId primitive.ObjectID) bool {
	if u.Deployments != nil {
		for _, deplyId := range u.Deployments {
			if deplyId == deployId {
				return true
			}
		}
	}

	return false
}

func (u *Unit) Reserve(db *database.Database, deployId primitive.ObjectID,
	overrideCount int) (reserved bool, err error) {

	coll := db.Units()

	if overrideCount == 0 {
		if len(u.Deployments) >= u.Count {
			return
		}
	} else {
		if len(u.Deployments) >= overrideCount {
			return
		}
	}

	resp, err := coll.UpdateOne(db, bson.M{
		"_id":   u.Id,
		"pod":   u.Pod,
		"count": u.Count,
		"deployments": bson.M{
			"$size": len(u.Deployments),
		},
	}, bson.M{
		"$push": bson.M{
			"deployments": deployId,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if resp.MatchedCount == 1 && resp.ModifiedCount == 1 {
		reserved = true
	}

	return
}

func (u *Unit) RestoreDeployment(db *database.Database,
	deployId primitive.ObjectID) (err error) {

	coll := db.Units()

	_, err = coll.UpdateOne(db, bson.M{
		"_id": u.Id,
	}, bson.M{
		"$push": bson.M{
			"deployments": deployId,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (u *Unit) RemoveDeployement(db *database.Database,
	deployId primitive.ObjectID) (err error) {

	coll := db.Units()

	_, err = coll.UpdateOne(db, bson.M{
		"_id": u.Id,
	}, bson.M{
		"$pull": bson.M{
			"deployments": deployId,
		},
	})
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

	if newSpc.Pod != u.Pod || newSpc.Unit != u.Id {
		err = &errortypes.ParseError{
			errors.Newf("spec: Invalid unit"),
		}
		return
	}

	deplys, err := deployment.GetAll(db, &bson.M{
		"_id": &bson.M{
			"$in": deplyIds,
		},
		"pod":  u.Pod,
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
		"pod":  u.Pod,
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

func (u *Unit) Parse(db *database.Database, new bool) (
	errData *errortypes.ErrorData, err error) {

	spc := spec.New(u.Pod, u.Id, u.Organization, u.Spec)

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

		if new {
			spc.Index = 1
			spc.Timestamp = time.Now()
		} else {
			timestamp, index, e := NewSpec(db, u.Pod, u.Id)
			if e != nil {
				err = e
				return
			}

			spc.Index = index
			spc.Timestamp = timestamp
		}

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

func (u *Unit) Commit(db *database.Database) (err error) {
	coll := db.Units()

	err = coll.Commit(u.Id, u)
	if err != nil {
		return
	}

	return
}

func (u *Unit) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Units()

	err = coll.CommitFields(u.Id, u, fields)
	if err != nil {
		return
	}

	return
}

func (u *Unit) Insert(db *database.Database) (err error) {
	coll := db.Units()

	if u.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("unit: Cannot insert unit without id"),
		}
		return
	}

	resp, err := coll.InsertOne(db, u)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	u.Id = resp.InsertedID.(primitive.ObjectID)

	return
}
