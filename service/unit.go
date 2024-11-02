package service

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/spec"
)

type Unit struct {
	Service      *Service           `bson:"-" json:"-"`
	Id           primitive.ObjectID `bson:"id" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Kind         string             `bson:"kind" json:"kind"`
	Count        int                `bson:"count" json:"count"`
	Deployments  []*Deployment      `bson:"deployments" json:"deployments"`
	Spec         string             `bson:"spec" json:"spec"`
	LastCommit   primitive.ObjectID `bson:"last_commit" json:"last_commit"`
	DeployCommit primitive.ObjectID `bson:"deploy_commit" json:"deploy_commit"`
	Hash         string             `bson:"hash" json:"hash"`
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

func (u *Unit) Reserve(db *database.Database, deployId primitive.ObjectID) (
	reserved bool, err error) {

	coll := db.Services()

	if len(u.Deployments) >= u.Count {
		return
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
		"_id": u.Service.Id,
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

	coll := db.Services()

	updateOpts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"elem.id": u.Id},
			bson.M{"deploy.id": deploymentId},
		},
	})
	resp, err := coll.UpdateOne(db, bson.M{
		"_id": u.Service.Id,
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

	coll := db.Services()

	updateOpts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{
				"elem.id": u.Id,
			},
		},
	})
	_, err = coll.UpdateOne(db, bson.M{
		"_id": u.Service.Id,
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

	coll := db.Services()

	updateOpts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"elem.id": u.Id},
		},
	})
	_, err = coll.UpdateOne(db, bson.M{
		"_id": u.Service.Id,
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

func (u *Unit) Parse(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	spc := spec.New(u.Service.Id, u.Id, u.Spec)

	errData, err = spc.Parse(db, u.Service.Organization)
	if err != nil {
		return
	}
	if errData != nil {
		return
	}

	if u.Hash != spc.Hash {
		u.Name = spc.Name
		u.Kind = spc.Kind
		u.Count = spc.Count
		u.Spec = spc.Data

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

		err = curSpc.CommitFields(db, set.NewSet("name", "count", "data"))
		if err != nil {
			return
		}

		u.Name = curSpc.Name
		u.Kind = curSpc.Kind
		u.Count = curSpc.Count
		u.Spec = curSpc.Data

		u.Hash = curSpc.Hash
		u.LastCommit = curSpc.Id
		if u.DeployCommit.IsZero() {
			u.DeployCommit = curSpc.Id
		}
	}

	return
}
