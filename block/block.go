package block

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/requires"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Block struct {
	Id       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name     string             `bson:"name" json:"name"`
	Comment  string             `bson:"comment" json:"comment"`
	Type     string             `bson:"type" json:"type"`
	Subnets  []string           `bson:"subnets" json:"subnets"`
	Subnets6 []string           `bson:"subnets6" json:"subnets6"`
	Excludes []string           `bson:"excludes" json:"excludes"`
	Netmask  string             `bson:"netmask" json:"netmask"`
	Gateway  string             `bson:"gateway" json:"gateway"`
	Gateway6 string             `bson:"gateway6" json:"gateway6"`
}

func (b *Block) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	b.Name = utils.FilterName(b.Name)

	if b.Type == "" {
		b.Type = IPv4
	}

	if b.Subnets == nil {
		b.Subnets = []string{}
	}

	if b.Subnets6 == nil {
		b.Subnets6 = []string{}
	}

	if b.Excludes == nil {
		b.Excludes = []string{}
	}

	if b.Type == IPv4 {
		b.Subnets6 = []string{}

		if b.Gateway != "" {
			gateway := net.ParseIP(b.Gateway)
			if gateway == nil || gateway.To4() == nil {
				errData = &errortypes.ErrorData{
					Error:   "invalid_gateway",
					Message: "Gateway address is invalid",
				}
				return
			}
			b.Gateway = gateway.String()
		}

		if b.Netmask != "" {
			netmask := utils.ParseIpMask(b.Netmask)
			if netmask == nil {
				errData = &errortypes.ErrorData{
					Error:   "invalid_netmask",
					Message: "Netmask is invalid",
				}
				return
			}
		}

		subnets := []string{}
		for _, subnet := range b.Subnets {
			if !strings.Contains(subnet, "/") {
				subnet += "/32"
			}

			_, subnetNet, e := net.ParseCIDR(subnet)
			if e != nil || subnetNet.IP.To4() == nil {
				errData = &errortypes.ErrorData{
					Error:   "invalid_subnet",
					Message: "Invalid subnet address",
				}
				return
			}

			subnets = append(subnets, subnetNet.String())
		}
		b.Subnets = subnets

		excludes := []string{}
		for _, exclude := range b.Excludes {
			if !strings.Contains(exclude, "/") {
				exclude += "/32"
			}

			_, excludeNet, e := net.ParseCIDR(exclude)
			if e != nil || excludeNet.IP.To4() == nil {
				errData = &errortypes.ErrorData{
					Error:   "invalid_exclude",
					Message: "Invalid exclude address",
				}
				return
			}

			excludes = append(excludes, excludeNet.String())
		}
		b.Excludes = excludes
	} else if b.Type == IPv6 {
		b.Subnets = []string{}
		b.Excludes = []string{}
		b.Netmask = ""
		b.Gateway = ""

		if b.Gateway6 != "" {
			gateway6 := net.ParseIP(b.Gateway6)
			if gateway6 == nil || gateway6.To4() != nil {
				errData = &errortypes.ErrorData{
					Error:   "invalid_gateway6",
					Message: "Gateway IPv6 address is invalid",
				}
				return
			}
			b.Gateway6 = gateway6.String()
		}

		subnets6 := []string{}
		for _, subnet6 := range b.Subnets6 {
			if !strings.Contains(subnet6, "/") {
				errData = &errortypes.ErrorData{
					Error:   "invalid_subnet6",
					Message: "Missing subnet6 cidr",
				}
				return
			}

			_, subnetNet, e := net.ParseCIDR(subnet6)
			if e != nil || subnetNet.IP.To4() != nil {
				errData = &errortypes.ErrorData{
					Error:   "invalid_subnet6",
					Message: "Invalid subnet6 address",
				}
				return
			}

			size, _ := subnetNet.Mask.Size()
			if size > 64 {
				errData = &errortypes.ErrorData{
					Error:   "invalid_subnet6_size",
					Message: "Minimum subnet6 size 64 is required",
				}
				return
			}

			subnets6 = append(subnets6, subnetNet.String())
		}
		b.Subnets6 = subnets6

		if len(b.Subnets6) > 1 {
			errData = &errortypes.ErrorData{
				Error:   "invalid_subnets6",
				Message: "Currently only one IPv6 subnet is supported",
			}
			return
		}
	} else {
		errData = &errortypes.ErrorData{
			Error:   "invalid_type",
			Message: "Block type is invalid",
		}
		return
	}

	return
}

func (b *Block) Contains(blckIp *BlockIp) (contains bool, err error) {
	ip := blckIp.GetIp()

	for _, exclude := range b.Excludes {
		_, network, e := net.ParseCIDR(exclude)
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "block: Failed to parse block exclude"),
			}
			return
		}

		if network.Contains(ip) {
			return
		}
	}

	for _, subnet := range b.Subnets {
		_, network, e := net.ParseCIDR(subnet)
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "block: Failed to parse block subnet"),
			}
			return
		}

		if network.Contains(ip) {
			contains = true
			return
		}
	}

	return
}

func (b *Block) GetGateway() net.IP {
	return net.ParseIP(b.Gateway)
}

func (b *Block) GetGateway6() net.IP {
	if b.Gateway6 == "" {
		return nil
	}
	return net.ParseIP(b.Gateway6)
}

func (b *Block) GetMask() net.IPMask {
	return utils.ParseIpMask(b.Netmask)
}

func (b *Block) GetGatewayCidr() string {
	staticGateway := net.ParseIP(b.Gateway)
	staticMask := utils.ParseIpMask(b.Netmask)
	if staticGateway == nil || staticMask == nil {
		return ""
	}

	staticSize, _ := staticMask.Size()
	return fmt.Sprintf("%s/%d", staticGateway.String(), staticSize)
}

func (b *Block) GetNetwork() (staticNet *net.IPNet, err error) {
	staticMask := utils.ParseIpMask(b.Netmask)
	if staticMask == nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "block: Invalid netmask"),
		}
		return
	}
	staticSize, _ := staticMask.Size()

	_, staticNet, err = net.ParseCIDR(
		fmt.Sprintf("%s/%d", b.Gateway, staticSize))
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "block: Failed to parse network cidr"),
		}
		return
	}

	return
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
	instId primitive.ObjectID, typ string) (ip net.IP, err error) {

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

	gatewaySize, _ := b.GetMask().Size()
	_, gatewayCidr, err := net.ParseCIDR(fmt.Sprintf("%s/%d",
		gateway.String(), gatewaySize))
	if err != nil {
		err = &errortypes.ParseError{
			errors.New("block: Failed to parse block gateway cidr"),
		}
		return
	}

	broadcast := utils.GetLastIpAddress(gatewayCidr)

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

	for _, subnet := range b.Subnets {
		_, network, e := net.ParseCIDR(subnet)
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "block: Failed to parse block subnet"),
			}
			return
		}

		first := true
		curIp := utils.CopyIpAddress(network.IP)
		for {
			if first {
				first = false
			} else {
				utils.IncIpAddress(curIp)
			}
			curIpInt := utils.IpAddress2Int(curIp)

			if !network.Contains(curIp) {
				break
			}

			if blckIps.Contains(curIpInt) || gatewayCidr.IP.Equal(curIp) ||
				gateway.Equal(curIp) || broadcast.Equal(curIp) {

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
				Type:     typ,
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

func (b *Block) GetIp6(db *database.Database,
	instId primitive.ObjectID, vlan int) (ip net.IP, cidr int, err error) {

	subnets6 := b.Subnets6
	if subnets6 == nil || len(subnets6) < 1 {
		return
	}

	subnet6 := subnets6[0]

	if vlan == 0 || vlan > 4095 {
		err = &errortypes.ParseError{
			errors.New("block: Failed to split subnet6"),
		}
		return
	}

	_, subnetNet, err := net.ParseCIDR(subnet6)
	if err != nil || subnetNet.IP.To4() != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "block: Invalid subnet6"),
		}
		return
	}

	cidr, _ = subnetNet.Mask.Size()

	subnet6 = subnetNet.String()

	subnet6spl := strings.Split(subnet6, ":/")
	if len(subnet6spl) != 2 {
		err = &errortypes.ParseError{
			errors.New("block: Failed to split subnet6"),
		}
		return
	}
	addr6 := subnet6spl[0]

	if strings.Count(addr6, ":") < 4 {
		addr6 += ":"
	}

	addr6 += "0" + fmt.Sprintf("%03x", vlan) + ":"

	hash := md5.New()
	hash.Write([]byte(instId.Hex()))
	macHash := fmt.Sprintf("%x", hash.Sum(nil))
	macHash = macHash[:12]
	macBuf := bytes.Buffer{}

	for i, run := range macHash {
		if i != 0 && i%4 == 0 {
			macBuf.WriteRune(':')
		}
		macBuf.WriteRune(run)
	}

	addr6 += macBuf.String() + fmt.Sprintf("/%d", cidr)

	ip, _, err = net.ParseCIDR(addr6)
	if err != nil || subnetNet.IP.To4() != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "block: Failed to parse address6"),
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
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		} else {
			return
		}
	}

	return
}

func (b *Block) ValidateAddresses(db *database.Database,
	commitFields set.Set) (err error) {

	coll := db.Blocks()
	ipColl := db.BlocksIp()
	instColl := db.Instances()

	gateway := net.ParseIP(b.Gateway)
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

	subnets := []*net.IPNet{}
	for _, subnet := range b.Subnets {
		_, network, e := net.ParseCIDR(subnet)
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "block: Failed to parse block subnet"),
			}
			return
		}

		subnets = append(subnets, network)
	}

	if commitFields != nil {
		err = coll.CommitFields(b.Id, b, commitFields)
		if err != nil {
			return
		}
	}

	cursor, err := ipColl.Find(db, &bson.M{
		"block": b.Id,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		blckIp := &BlockIp{}
		err = cursor.Decode(blckIp)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		remove := false
		ip := utils.Int2IpAddress(blckIp.Ip)

		if gateway != nil && gateway.Equal(ip) {
			remove = true
		}

		if !remove {
			for _, exclude := range excludes {
				if exclude.Contains(ip) {
					remove = true
					break
				}
			}
		}

		if !remove {
			match := false
			for _, subnet := range subnets {
				if subnet.Contains(ip) {
					match = true
					break
				}
			}

			if !match {
				remove = true
			}
		}

		if remove {
			_, _ = instColl.UpdateOne(db, &bson.M{
				"_id": blckIp.Instance,
			}, &bson.M{
				"$set": &bson.M{
					"restart_block_ip": true,
				},
			})

			_, err = ipColl.DeleteOne(db, &bson.M{
				"_id": blckIp.Id,
			})
			if err != nil {
				err = database.ParseError(err)
				if _, ok := err.(*database.NotFoundError); ok {
					err = nil
				} else {
					return
				}
			}
		}
	}

	err = cursor.Err()
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

	err = b.ValidateAddresses(db, fields)
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

	resp, err := coll.InsertOne(db, b)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	b.Id = resp.InsertedID.(primitive.ObjectID)

	return
}

func init() {
	module := requires.New("block")
	module.After("settings")

	module.Handler = func() (err error) {
		db := database.GetDatabase()
		defer db.Close()

		coll := db.BlocksIp()

		// TODO Upgrade <= 1.0.1173.24
		_, err = coll.UpdateMany(db, &bson.M{
			"type": &bson.M{
				"$exists": false,
			},
		}, &bson.M{
			"$set": &bson.M{
				"type": External,
			},
		})
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		}

		return
	}
}
