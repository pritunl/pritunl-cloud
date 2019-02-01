package block

import (
	"fmt"
	"net"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Block struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Addresses []string           `bson:"addresses" json:"addresses"`
	Excludes  []string           `bson:"excludes" json:"excludes"`
	Netmask   string             `bson:"netmask" json:"netmask"`
	Gateway   string             `bson:"gateway" json:"gateway"`
}

func (b *Block) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if b.Addresses == nil {
		b.Addresses = []string{}
	}

	if b.Excludes == nil {
		b.Excludes = []string{}
	}

	gateway := net.ParseIP(b.Gateway)
	if gateway == nil {
		errData = &errortypes.ErrorData{
			Error:   "invalid_gateway",
			Message: "Gateway address is invalid",
		}
		return
	}

	netmask := utils.ParseIpMask(b.Netmask)
	if netmask == nil {
		errData = &errortypes.ErrorData{
			Error:   "invalid_netmask",
			Message: "Netmask is invalid",
		}
		return
	}

	subnets := []string{}
	for _, subnet := range b.Addresses {
		if !strings.Contains(subnet, "/") {
			subnet += "/32"
		}

		_, subnetNet, e := net.ParseCIDR(subnet)
		if e != nil {
			errData = &errortypes.ErrorData{
				Error:   "invalid_subnet",
				Message: "Invalid subnet address",
			}
			return
		}

		subnets = append(subnets, subnetNet.String())
	}
	b.Addresses = subnets

	excludes := []string{}
	for _, exclude := range b.Excludes {
		if !strings.Contains(exclude, "/") {
			exclude += "/32"
		}

		_, excludeNet, e := net.ParseCIDR(exclude)
		if e != nil {
			errData = &errortypes.ErrorData{
				Error:   "invalid_exclude",
				Message: "Invalid exclude address",
			}
			return
		}

		excludes = append(excludes, excludeNet.String())
	}
	b.Excludes = excludes

	return
}

func (b *Block) GetGateway() net.IP {
	return net.ParseIP(b.Gateway)
}

func (b *Block) GetMask() net.IPMask {
	return utils.ParseIpMask(b.Netmask)
}

func (b *Block) GetIps(db *database.Database) (blckIps set.Set, err error) {
	coll := db.BlocksIp()

	ipsInf, err := coll.Distinct(db, "ip", &bson.M{
		"block": b.Id,
	})

	blckIps = set.NewSet()
	for _, ipInf := range ipsInf {
		if ip, ok := ipInf.(int64); ok {
			blckIps.Add(ip)
		}
	}

	return
}

func (b *Block) GetIp(db *database.Database,
	instId primitive.ObjectID) (ip net.IP, err error) {

	blckIps, err := b.GetIps(db)
	if err != nil {
		return
	}

	coll := db.BlocksIp()
	gateway := net.ParseIP(b.Gateway)
	if gateway == nil {
		err = &errortypes.ParseError{
			errors.New("block: Failed to parse block gateway"),
		}
		return
	}

	excludes := []*net.IPNet{}
	for _, exclude := range b.Excludes {
		_, network, e := net.ParseCIDR(exclude)
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "block: Failed to parse block exclude"),
			}
			return
		}

		excludes = append(excludes, network)
	}

	for _, subnet := range b.Addresses {
		_, network, e := net.ParseCIDR(subnet)
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "block: Failed to parse block subnet"),
			}
			return
		}

		curIp := utils.CopyIpAddress(network.IP)
		for {
			utils.IncIpAddress(curIp)
			curIpInt := utils.IpAddress2Int(curIp)

			if !network.Contains(curIp) {
				break
			}

			if blckIps.Contains(curIpInt) || gateway.Equal(curIp) {
				continue
			}

			excluded := false
			for _, exclude := range excludes {
				if exclude.Contains(curIp) {
					excluded = true
					break
				}
			}

			if excluded {
				continue
			}

			blckIp := &BlockIp{
				Block:    b.Id,
				Ip:       utils.IpAddress2Int(curIp),
				Instance: instId,
			}

			_, err = coll.InsertOne(db, blckIp)
			if err != nil {
				err = database.ParseError(err)
				if _, ok := err.(*database.DuplicateKeyError); ok {
					err = nil
					continue
				}
				return
			}

			ip = curIp
			break
		}

		if ip != nil {
			break
		}
	}

	if ip == nil {
		err = &BlockFull{
			errors.New("block: Address pool full"),
		}
		return
	}

	return
}

func (b *Block) RemoveIp(db *database.Database,
	instId primitive.ObjectID) (err error) {

	coll := db.BlocksIp()
	_, err = coll.DeleteMany(db, &bson.M{
		"instance": instId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (b *Block) Commit(db *database.Database) (err error) {
	coll := db.Blocks()

	err = coll.Commit(b.Id, b)
	if err != nil {
		return
	}

	return
}

func (b *Block) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Blocks()

	err = coll.CommitFields(b.Id, b, fields)
	if err != nil {
		return
	}

	return
}

func (b *Block) Insert(db *database.Database) (err error) {
	coll := db.Blocks()

	if !b.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("block: Block already exists"),
		}
		return
	}

	_, err = coll.InsertOne(db, b)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
