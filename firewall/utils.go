package firewall

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
)

func Get(db *database.Database, fireId bson.ObjectId) (
	fire *Firewall, err error) {

	coll := db.Firewalls()
	fire = &Firewall{}

	err = coll.FindOneId(fireId, fire)
	if err != nil {
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	fires []*Firewall, err error) {

	coll := db.Firewalls()
	fires = []*Firewall{}

	cursor := coll.Find(query).Iter()

	nde := &Firewall{}
	for cursor.Next(nde) {
		fires = append(fires, nde)
		nde = &Firewall{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetRoles(db *database.Database, roles []string) (
	fires []*Firewall, err error) {

	coll := db.Firewalls()
	fires = []*Firewall{}

	cursor := coll.Find(&bson.M{
		"network_roles": &bson.M{
			"$in": roles,
		},
	}).Sort("_id").Iter()

	nde := &Firewall{}
	for cursor.Next(nde) {
		fires = append(fires, nde)
		nde = &Firewall{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOrgMapRoles(db *database.Database, orgId bson.ObjectId) (
	fires map[string][]*Firewall, err error) {

	coll := db.Firewalls()
	fires = map[string][]*Firewall{}

	cursor := coll.Find(&bson.M{
		"organization": orgId,
	}).Iter()

	fire := &Firewall{}
	for cursor.Next(fire) {
		for _, role := range fire.NetworkRoles {
			roleFires := fires[role]
			if roleFires == nil {
				roleFires = []*Firewall{}
			}
			fires[role] = append(roleFires, fire)
		}
		fire = &Firewall{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOrgRoles(db *database.Database, orgId bson.ObjectId,
	roles []string) (fires []*Firewall, err error) {

	coll := db.Firewalls()
	fires = []*Firewall{}

	cursor := coll.Find(&bson.M{
		"organization": orgId,
		"network_roles": &bson.M{
			"$in": roles,
		},
	}).Sort("_id").Iter()

	nde := &Firewall{}
	for cursor.Next(nde) {
		fires = append(fires, nde)
		nde = &Firewall{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M, page, pageCount int) (
	fires []*Firewall, count int, err error) {

	coll := db.Firewalls()
	fires = []*Firewall{}

	qury := coll.Find(query)

	count, err = qury.Count()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	skip := utils.Min(page*pageCount, utils.Max(0, count-pageCount))

	cursor := qury.Sort("name").Skip(skip).Limit(pageCount).Iter()

	fire := &Firewall{}
	for cursor.Next(fire) {
		fires = append(fires, fire)
		fire = &Firewall{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, fireId bson.ObjectId) (err error) {
	coll := db.Firewalls()

	err = coll.Remove(&bson.M{
		"_id": fireId,
	})
	if err != nil {
		err = database.ParseError(err)
		switch err.(type) {
		case *database.NotFoundError:
			err = nil
		default:
			return
		}
	}

	return
}

func RemoveMulti(db *database.Database, fireIds []bson.ObjectId) (err error) {
	coll := db.Firewalls()

	_, err = coll.RemoveAll(&bson.M{
		"_id": &bson.M{
			"$in": fireIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
