package unit

import (
	"sync"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/tools/errors"
)

type Unit struct {
	Id            bson.ObjectID    `bson:"_id,omitempty" json:"id"`
	Pod           bson.ObjectID    `bson:"pod" json:"pod"`
	Organization  bson.ObjectID    `bson:"organization" json:"organization"`
	Name          string           `bson:"name" json:"name"`
	Kind          string           `bson:"kind" json:"kind"`
	Count         int              `bson:"count" json:"count"`
	Deployments   []bson.ObjectID  `bson:"deployments" json:"deployments"`
	Spec          string           `bson:"spec" json:"spec"`
	SpecIndex     int              `bson:"spec_index" json:"spec_index"`
	SpecTimestamp time.Time        `bson:"spec_timestamp" json:"-"`
	LastSpec      bson.ObjectID    `bson:"last_spec" json:"last_spec"`
	DeploySpec    bson.ObjectID    `bson:"deploy_spec" json:"deploy_spec"`
	Hash          string           `bson:"hash" json:"hash"`
	Journals      map[string]int32 `bson:"journals" json:"-"`
	JournalsIndex int32            `bson:"journals_index" json:"-"`
	journalsLock  sync.Mutex       `bson:"-" json:"-"`
}

type Completion struct {
	Id           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Pod          bson.ObjectID `bson:"pod" json:"pod"`
	Organization bson.ObjectID `bson:"organization" json:"organization"`
	Name         string        `bson:"name" json:"name"`
	Kind         string        `bson:"kind" json:"kind"`
}

type UnitInput struct {
	Id         bson.ObjectID `json:"id"`
	Name       string        `json:"name"`
	Spec       string        `json:"spec"`
	DeploySpec bson.ObjectID `json:"deploy_spec"`
	Delete     bool          `json:"delete"`
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

	u.Id = unt.Id
	u.Pod = unt.Pod
	u.Organization = unt.Organization
	u.Name = unt.Name
	u.Kind = unt.Kind
	u.Count = unt.Count
	u.Deployments = unt.Deployments
	u.Spec = unt.Spec
	u.SpecIndex = unt.SpecIndex
	u.SpecTimestamp = unt.SpecTimestamp
	u.LastSpec = unt.LastSpec
	u.DeploySpec = unt.DeploySpec
	u.Hash = unt.Hash
	u.Journals = unt.Journals
	u.JournalsIndex = unt.JournalsIndex
	return
}

func (u *Unit) RefreshJournals(db *database.Database) (err error) {
	coll := db.Units()

	unt := &Unit{}
	err = coll.FindOne(db, &bson.M{
		"_id": u.Id,
	}).Decode(unt)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	u.journalsLock.Lock()
	u.Journals = unt.Journals
	u.JournalsIndex = unt.JournalsIndex
	u.journalsLock.Unlock()

	return
}

func (u *Unit) HasDeployment(deployId bson.ObjectID) bool {
	if u.Deployments != nil {
		for _, deplyId := range u.Deployments {
			if deplyId == deployId {
				return true
			}
		}
	}

	return false
}

func (u *Unit) Reserve(db *database.Database, deployId bson.ObjectID,
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
	deployId bson.ObjectID) (err error) {

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
	deployId bson.ObjectID) (err error) {

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
	newSpecId bson.ObjectID, deplyIds []bson.ObjectID) (
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

	spcMap := map[bson.ObjectID]*spec.Spec{}

	for _, deply := range deplys {
		oldSpc := spcMap[deply.Spec]
		if oldSpc == nil {
			oldSpc, err = spec.Get(db, deply.Spec)
			if err != nil {
				return
			}

			spcMap[oldSpc.Id] = oldSpc
		}

		errData, err = oldSpc.CanMigrate(db, deply, newSpc)
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

func (u *Unit) newSpec(db *database.Database, spc *spec.Spec, newUnit bool) (
	newSpec *spec.Spec, errData *errortypes.ErrorData, err error) {

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

	if newUnit {
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

	newSpec = spc

	u.Hash = spc.Hash
	u.LastSpec = spc.Id
	if u.DeploySpec.IsZero() {
		u.DeploySpec = spc.Id
	}

	return
}

func (u *Unit) updateSpec(db *database.Database, spc *spec.Spec) (
	updateSpec *spec.Spec, errData *errortypes.ErrorData, err error) {

	curSpc, e := spec.Get(db, u.LastSpec)
	if e != nil {
		err = e
		return
	}

	curSpc.Name = spc.Name
	curSpc.Count = spc.Count
	curSpc.Data = spc.Data

	updateSpec = curSpc

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
	u.LastSpec = curSpc.Id
	if u.DeploySpec.IsZero() {
		u.DeploySpec = curSpc.Id
	}

	return
}

func (u *Unit) getKind(db *database.Database, key string) (
	kind int32, err error) {

	u.journalsLock.Lock()
	defer u.journalsLock.Unlock()

	if u.Journals == nil {
		u.Journals = map[string]int32{}
	}

	kind, ok := u.Journals[key]
	if ok && kind != 0 {
		return
	}

	jrnls := map[string]int32{}
	for key, index := range u.Journals {
		jrnls[key] = index
	}
	index := u.JournalsIndex
	if index == 0 {
		index = 248000
	}
	index += 1
	jrnls[key] = index

	coll := db.Units()

	query := bson.M{
		"_id": u.Id,
	}
	if u.JournalsIndex == 0 {
		query["$or"] = []bson.M{
			{"journals_index": bson.M{"$exists": false}},
			{"journals_index": 0},
		}
	} else {
		query["journals_index"] = u.JournalsIndex
	}

	resp, err := coll.UpdateOne(db, query, bson.M{
		"$set": bson.M{
			"journals_index": index,
			"journals":       jrnls,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if resp.ModifiedCount < 1 {
		kind = 0
		return
	}

	u.Journals = jrnls
	u.JournalsIndex = index
	kind = index

	return
}

func (u *Unit) GetKind(db *database.Database, key string) (
	kind int32, err error) {

	for i := 0; i < 3; i++ {
		kind, err = u.getKind(db, key)
		if err != nil {
			return
		}

		if kind != 0 {
			break
		}

		err = u.RefreshJournals(db)
		if err != nil {
			return
		}
	}

	if kind == 0 {
		err = &errortypes.ParseError{
			errors.New("unit: Failed to get journal kind index"),
		}
		return
	}

	return
}

func (u *Unit) Parse(db *database.Database, newUnit bool) (
	newSpec *spec.Spec, updateSpec *spec.Spec,
	errData *errortypes.ErrorData, err error) {

	spc := spec.New(u.Pod, u.Id, u.Organization, u.Spec)

	errData, err = spc.Parse(db, u)
	if err != nil {
		return
	}
	if errData != nil {
		return
	}

	isNewSpec := u.Hash != spc.Hash
	if !isNewSpec && u.Name != spc.Name || u.Count != spc.Count {
		updateSpec, errData, err = u.updateSpec(db, spc)
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
				isNewSpec = true
			} else {
				return
			}
		}
		if errData != nil {
			return
		}
	}

	if isNewSpec {
		newSpec, errData, err = u.newSpec(db, spc, newUnit)
		if err != nil {
			return
		}
		if errData != nil {
			return
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

	u.Id = resp.InsertedID.(bson.ObjectID)

	return
}
