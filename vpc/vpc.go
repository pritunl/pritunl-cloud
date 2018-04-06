package vpc

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"net"
	"strings"
)

type Route struct {
	Destination string `bson:"destination" json:"destination"`
	Target      string `bson:"target" json:"target"`
}

type Vpc struct {
	Id           bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Name         string        `bson:"name" json:"name"`
	VpcId        int           `bson:"vpc_id" json:"vpc_id"`
	Network      string        `bson:"network" json:"network"`
	Organization bson.ObjectId `bson:"organization" json:"organization"`
	Datacenter   bson.ObjectId `bson:"datacenter" json:"datacenter"`
	Routes       []*Route      `bson:"routes" json:"routes"`
}

func (v *Vpc) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if v.VpcId == 0 {
		errData = &errortypes.ErrorData{
			Error:   "vpc_id_invalid",
			Message: "Vpc ID invalid",
		}
		return
	}

	if v.Organization == "" {
		errData = &errortypes.ErrorData{
			Error:   "organization_required",
			Message: "Missing required organization",
		}
		return
	}

	if v.Datacenter == "" {
		errData = &errortypes.ErrorData{
			Error:   "datacenter_required",
			Message: "Missing required datacenter",
		}
		return
	}

	network, e := v.GetNetwork()
	if e != nil {
		errData = &errortypes.ErrorData{
			Error:   "network_invalid",
			Message: "Network address invalid",
		}
		return
	}
	v.Network = network.String()

	if v.Routes == nil {
		v.Routes = []*Route{}
	}

	destinations := set.NewSet()
	for _, route := range v.Routes {
		if destinations.Contains(route.Destination) {
			errData = &errortypes.ErrorData{
				Error:   "duplicate_destination",
				Message: "Duplicate route destinations",
			}
			return
		}
		destinations.Add(route.Destination)

		if strings.Contains(route.Destination, ":") ||
			strings.Contains(route.Target, ":") {

			errData = &errortypes.ErrorData{
				Error:   "route_ipv6_not_supported",
				Message: "Route IPv6 currently unsupported",
			}
			return
		}

		_, destination, e := net.ParseCIDR(route.Destination)
		if e != nil {
			errData = &errortypes.ErrorData{
				Error:   "route_destination_invalid",
				Message: "Route destination invalid",
			}
			return
		}
		route.Destination = destination.String()

		if route.Destination == "0.0.0.0/0" {
			errData = &errortypes.ErrorData{
				Error:   "route_destination_invalid",
				Message: "Route destination invalid",
			}
			return
		}

		target := net.ParseIP(route.Target)
		if target == nil {
			errData = &errortypes.ErrorData{
				Error:   "route_target_invalid",
				Message: "Route target invalid",
			}
			return
		}
		route.Target = target.String()

		if route.Target == "0.0.0.0" || !network.Contains(target) {
			errData = &errortypes.ErrorData{
				Error:   "route_target_invalid_network",
				Message: "Route target not in VPC network",
			}
			return
		}
	}

	return
}

func (v *Vpc) GetNetwork() (network *net.IPNet, err error) {
	_, network, err = net.ParseCIDR(v.Network)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "vpc: Failed to parse network"),
		}
		return
	}
	return
}

func (v *Vpc) GenerateVpcId() {
	v.VpcId = rand.Intn(4085) + 10
}

func (v *Vpc) GetGateway() (ip net.IP, err error) {
	network, err := v.GetNetwork()
	if err != nil {
		return
	}

	ip = network.IP
	utils.IncIpAddress(ip)

	return
}

func (v *Vpc) GetIp(db *database.Database, typ string, instId bson.ObjectId) (
	ip net.IP, err error) {

	coll := db.VpcsIp()
	vpcIp := &VpcIp{}

	err = coll.FindOne(&bson.M{
		"vpc":      v.Id,
		"type":     typ,
		"instance": instId,
	}, vpcIp)
	if err != nil {
		vpcIp = nil
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		} else {
			return
		}
	}

	if vpcIp == nil {
		vpcIp = &VpcIp{}
		change := mgo.Change{
			Update: &bson.M{
				"$set": &bson.M{
					"instance": instId,
				},
			},
			ReturnNew: true,
		}

		info, e := coll.Find(&bson.M{
			"vpc":      v.Id,
			"type":     typ,
			"instance": nil,
		}).Apply(change, vpcIp)
		if e != nil {
			err = database.ParseError(e)
			vpcIp = nil
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		} else if info.Updated == 0 {
			vpcIp = nil
		}
	}

	if vpcIp == nil {
		vpcIp = &VpcIp{}

		sort := ""
		if typ == Gateway {
			sort = "ip"
		} else {
			sort = "-ip"
		}

		err = coll.Find(&bson.M{
			"vpc":  v.Id,
			"type": typ,
		}).Sort(sort).One(vpcIp)
		if err != nil {
			vpcIp = nil
			err = database.ParseError(err)
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		}

		network, e := v.GetNetwork()
		if e != nil {
			err = e
			return
		}

		var curIp net.IP
		if typ == Gateway {
			if vpcIp == nil {
				curIp = utils.GetLastIpAddress(network)
				utils.DecIpAddress(curIp)
			} else {
				curIp = utils.Int2IpAddress(vpcIp.Ip)
			}
		} else {
			if vpcIp == nil {
				curIp = utils.CopyIpAddress(network.IP)
				utils.IncIpAddress(curIp)
			} else {
				curIp = utils.Int2IpAddress(vpcIp.Ip)
			}
		}

		for {
			if typ == Gateway {
				utils.DecIpAddress(curIp)
			} else {
				utils.IncIpAddress(curIp)
			}

			vpcIp = &VpcIp{
				Vpc:      v.Id,
				Type:     typ,
				Ip:       utils.IpAddress2Int(curIp),
				Instance: instId,
			}

			if !network.Contains(curIp) {
				vpcIp = nil
				err = &errortypes.NotFoundError{
					errors.New("vpc: Address pool full"),
				}
				return
			}

			err = coll.Insert(vpcIp)
			if err != nil {
				vpcIp = nil
				err = database.ParseError(err)
				if _, ok := err.(*database.DuplicateKeyError); ok {
					err = nil
					continue
				}
				return
			}

			break
		}
	}

	if vpcIp == nil {
		err = &errortypes.NotFoundError{
			errors.New("vpc: Address not found"),
		}
		return
	}

	ip = utils.Int2IpAddress(vpcIp.Ip)

	return
}

func (v *Vpc) GetIp6(addr net.IP) net.IP {
	netHash := md5.New()
	netHash.Write([]byte(v.Id))
	netHashSum := fmt.Sprintf("%x", netHash.Sum(nil))[:12]

	macHash := md5.New()
	macHash.Write(addr)
	macHashSum := fmt.Sprintf("%x", macHash.Sum(nil))[:16]

	ip := fmt.Sprintf("fd97%s%s", netHashSum, macHashSum)
	ipBuf := bytes.Buffer{}

	for i, run := range ip {
		if i%4 == 0 && i != 0 && i != len(ip)-1 {
			ipBuf.WriteRune(':')
		}
		ipBuf.WriteRune(run)
	}

	return net.ParseIP(ipBuf.String())
}

func (v *Vpc) Commit(db *database.Database) (err error) {
	coll := db.Vpcs()

	err = coll.Commit(v.Id, v)
	if err != nil {
		return
	}

	return
}

func (v *Vpc) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Vpcs()

	err = coll.CommitFields(v.Id, v, fields)
	if err != nil {
		return
	}

	return
}

func (v *Vpc) Insert(db *database.Database) (err error) {
	coll := db.Vpcs()

	if v.Id != "" {
		err = &errortypes.DatabaseError{
			errors.New("vpc: Vpc already exists"),
		}
		return
	}

	err = coll.Insert(v)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
