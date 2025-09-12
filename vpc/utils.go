package vpc

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

func GetIp6(vpcId, instId bson.ObjectID) net.IP {
	netHash := md5.New()
	netHash.Write(vpcId[:])
	netHashSum := fmt.Sprintf("%x", netHash.Sum(nil))[:12]

	instHash := md5.New()
	instHash.Write(instId[:])
	instHashSum := fmt.Sprintf("%x", instHash.Sum(nil))[:16]

	ip := fmt.Sprintf("fd97%s%s", netHashSum, instHashSum)
	ipBuf := bytes.Buffer{}

	for i, run := range ip {
		if i%4 == 0 && i != 0 && i != len(ip)-1 {
			ipBuf.WriteRune(':')
		}
		ipBuf.WriteRune(run)
	}

	return net.ParseIP(ipBuf.String())
}

func GetGatewayIp6(vpcId, instId bson.ObjectID) net.IP {
	netHash := md5.New()
	netHash.Write(vpcId[:])
	netHashSum := fmt.Sprintf("%x", netHash.Sum(nil))[:12]

	instHash := md5.New()
	instHash.Write([]byte("gateway"))
	instHash.Write(instId[:])
	instHashSum := fmt.Sprintf("%x", instHash.Sum(nil))[:16]

	ip := fmt.Sprintf("fd97%s%s", netHashSum, instHashSum)
	ipBuf := bytes.Buffer{}

	for i, run := range ip {
		if i%4 == 0 && i != 0 && i != len(ip)-1 {
			ipBuf.WriteRune(':')
		}
		ipBuf.WriteRune(run)
	}

	return net.ParseIP(ipBuf.String())
}

func GetLinkIp6(vpcId, instId bson.ObjectID) net.IP {
	netHash := md5.New()
	netHash.Write(vpcId[:])
	netHashSum := fmt.Sprintf("%x", netHash.Sum(nil))[:12]

	instHash := md5.New()
	instHash.Write(instId[:])
	instHashSum := fmt.Sprintf("%x", instHash.Sum(nil))[:16]

	ip := fmt.Sprintf("fd97%s%s", netHashSum, instHashSum)
	ipBuf := bytes.Buffer{}

	for i, run := range ip {
		if i%4 == 0 && i != 0 && i != len(ip)-1 {
			ipBuf.WriteRune(':')
		}
		ipBuf.WriteRune(run)
	}

	return net.ParseIP(ipBuf.String())
}

func GetGatewayLinkIp6(vpcId, instId bson.ObjectID) net.IP {
	netHash := md5.New()
	netHash.Write(vpcId[:])
	netHashSum := fmt.Sprintf("%x", netHash.Sum(nil))[:12]

	instHash := md5.New()
	instHash.Write([]byte("gateway"))
	instHash.Write(instId[:])
	instHashSum := fmt.Sprintf("%x", instHash.Sum(nil))[:16]

	ip := fmt.Sprintf("fd97%s%s", netHashSum, instHashSum)
	ipBuf := bytes.Buffer{}

	for i, run := range ip {
		if i%4 == 0 && i != 0 && i != len(ip)-1 {
			ipBuf.WriteRune(':')
		}
		ipBuf.WriteRune(run)
	}

	return net.ParseIP(ipBuf.String())
}

func Get(db *database.Database, vcId bson.ObjectID) (
	vc *Vpc, err error) {

	coll := db.Vpcs()
	vc = &Vpc{}

	err = coll.FindOneId(vcId, vc)
	if err != nil {
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, vcId bson.ObjectID) (
	vc *Vpc, err error) {

	coll := db.Vpcs()
	vc = &Vpc{}

	err = coll.FindOne(db, &bson.M{
		"_id":          vcId,
		"organization": orgId,
	}).Decode(vc)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func ExistsOrg(db *database.Database, orgId, vcId bson.ObjectID) (
	exists bool, err error) {

	coll := db.Vpcs()
	n, err := coll.CountDocuments(
		db,
		&bson.M{
			"_id":          vcId,
			"organization": orgId,
		},
	)
	if err != nil {
		return
	}

	if n > 0 {
		exists = true
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	vcs []*Vpc, err error) {

	coll := db.Vpcs()
	vcs = []*Vpc{}

	cursor, err := coll.Find(
		db,
		query,
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		vc := &Vpc{}
		err = cursor.Decode(vc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		vcs = append(vcs, vc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOne(db *database.Database, query *bson.M) (vc *Vpc, err error) {
	coll := db.Vpcs()
	vc = &Vpc{}

	err = coll.FindOne(db, query).Decode(vc)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllNames(db *database.Database, query *bson.M) (
	vpcs []*Vpc, err error) {

	coll := db.Vpcs()
	vpcs = []*Vpc{}

	cursor, err := coll.Find(
		db,
		query,
		options.Find().
			SetSort(bson.D{{"name", 1}}).
			SetProjection(bson.D{
				{"name", 1},
				{"organization", 1},
				{"datacenter", 1},
				{"type", 1},
				{"subnets", 1},
			}),
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		vc := &Vpc{}
		err = cursor.Decode(vc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		vpcs = append(vpcs, vc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (vcs []*Vpc, count int64, err error) {

	coll := db.Vpcs()
	vcs = []*Vpc{}

	if len(*query) == 0 {
		count, err = coll.EstimatedDocumentCount(db)
		if err != nil {
			err = database.ParseError(err)
			return
		}
	} else {
		count, err = coll.CountDocuments(db, query)
		if err != nil {
			err = database.ParseError(err)
			return
		}
	}

	maxPage := count / pageCount
	if count == pageCount {
		maxPage = 0
	}
	page = utils.Min64(page, maxPage)
	skip := utils.Min64(page*pageCount, count)

	cursor, err := coll.Find(
		db,
		query,
		options.Find().
			SetSort(bson.D{{"name", 1}}).
			SetSkip(skip).
			SetLimit(pageCount),
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		vc := &Vpc{}
		err = cursor.Decode(vc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		vcs = append(vcs, vc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetIds(db *database.Database, ids []bson.ObjectID) (
	vcs []*Vpc, err error) {

	coll := db.Vpcs()
	vcs = []*Vpc{}

	cursor, err := coll.Find(
		db,
		&bson.M{
			"_id": &bson.M{
				"$in": ids,
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		vc := &Vpc{}
		err = cursor.Decode(vc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		vcs = append(vcs, vc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetDatacenter(db *database.Database, dcId bson.ObjectID) (
	vcs []*Vpc, err error) {

	coll := db.Vpcs()
	vcs = []*Vpc{}

	cursor, err := coll.Find(
		db,
		&bson.M{
			"datacenter": dcId,
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		vc := &Vpc{}
		err = cursor.Decode(vc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		vcs = append(vcs, vc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func DistinctIds(db *database.Database, matchIds []bson.ObjectID) (
	idsSet set.Set, err error) {

	coll := db.Images()
	idsSet = set.NewSet()

	idsInf, err := coll.Distinct(
		db,
		"_id",
		&bson.M{
			"_id": &bson.M{
				"$in": matchIds,
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	for _, idInf := range idsInf {
		if id, ok := idInf.(bson.ObjectID); ok {
			idsSet.Add(id)
		}
	}

	return
}

func Remove(db *database.Database, vcId bson.ObjectID) (err error) {
	coll := db.VpcsIp()

	_, err = coll.DeleteMany(db, &bson.M{
		"vpc": vcId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Vpcs()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": vcId,
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

func RemoveOrg(db *database.Database, orgId, vcId bson.ObjectID) (
	err error) {

	coll := db.VpcsIp()

	_, err = coll.DeleteMany(db, &bson.M{
		"vpc": vcId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Vpcs()

	_, err = coll.DeleteOne(db, &bson.M{
		"organization": orgId,
		"_id":          vcId,
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

func RemoveMulti(db *database.Database, vcIds []bson.ObjectID) (err error) {
	coll := db.VpcsIp()

	_, err = coll.DeleteMany(db, &bson.M{
		"vpc": &bson.M{
			"$in": vcIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Vpcs()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": vcIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveMultiOrg(db *database.Database, orgId bson.ObjectID,
	vcIds []bson.ObjectID) (err error) {

	coll := db.VpcsIp()

	_, err = coll.DeleteMany(db, &bson.M{
		"vpc": &bson.M{
			"$in": vcIds,
		},
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Vpcs()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": vcIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetIpsMapped(db *database.Database, ids []bson.ObjectID) (
	vpcsMap map[bson.ObjectID][]*VpcIp, err error) {

	coll := db.VpcsIp()
	vpcsMap = map[bson.ObjectID][]*VpcIp{}

	cursor, err := coll.Find(
		db,
		&bson.M{
			"vpc": &bson.M{
				"$in": ids,
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		vc := &VpcIp{}
		err = cursor.Decode(vc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		vpcsMap[vc.Vpc] = append(vpcsMap[vc.Vpc], vc)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveInstanceIps(db *database.Database, instId bson.ObjectID) (
	err error) {

	coll := db.VpcsIp()

	_, err = coll.UpdateMany(db, &bson.M{
		"instance": instId,
	}, &bson.M{
		"$set": &bson.M{
			"instance": nil,
		},
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

func RemoveInstanceIp(db *database.Database, instId,
	vpcId bson.ObjectID) (err error) {

	coll := db.VpcsIp()

	_, err = coll.UpdateOne(
		db,
		&bson.M{
			"vpc":      vpcId,
			"instance": instId,
		},
		&bson.M{
			"$set": &bson.M{
				"instance": nil,
			},
		},
	)
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
