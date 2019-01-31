package vpc

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Route struct {
	Destination string `bson:"destination" json:"destination"`
	Target      string `bson:"target" json:"target"`
	Link        bool   `bson:"link" json:"link"`
}

type Vpc struct {
	Id            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string             `bson:"name" json:"name"`
	VpcId         int                `bson:"vpc_id" json:"vpc_id"`
	Network       string             `bson:"network" json:"network"`
	Network6      string             `bson:"-" json:"network6"`
	Organization  primitive.ObjectID `bson:"organization" json:"organization"`
	Datacenter    primitive.ObjectID `bson:"datacenter" json:"datacenter"`
	Routes        []*Route           `bson:"routes" json:"routes"`
	LinkUris      []string           `bson:"link_uris" json:"link_uris"`
	LinkNode      primitive.ObjectID `bson:"link_node,omitempty" json:"link_node"`
	LinkTimestamp time.Time          `bson:"link_timestamp" json:"link_timestamp"`
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

	if v.Organization.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "organization_required",
			Message: "Missing required organization",
		}
		return
	}

	if v.Datacenter.IsZero() {
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

	network6, e := v.GetNetwork6()
	if e != nil {
		errData = &errortypes.ErrorData{
			Error:   "network_invalid6",
			Message: "IPv6 network address invalid",
		}
		return
	}

	v.Network = network.String()

	if v.Routes == nil {
		v.Routes = []*Route{}
	}

	if v.LinkUris == nil {
		v.LinkUris = []string{}
	}

	linkUris := []string{}
	for _, linkUri := range v.LinkUris {
		if linkUri != "" {
			linkUris = append(linkUris, linkUri)
		}
	}
	v.LinkUris = linkUris

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

		if strings.Contains(route.Destination, ":") !=
			strings.Contains(route.Target, ":") {

			errData = &errortypes.ErrorData{
				Error:   "route_target_destination_invalid",
				Message: "Route target/destination invalid",
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

		if route.Destination == "0.0.0.0/0" || route.Destination == "::/0" {
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

		if route.Target == "0.0.0.0" {
			errData = &errortypes.ErrorData{
				Error:   "route_target_invalid",
				Message: "Route target invalid",
			}
			return
		}

		if !strings.Contains(route.Target, ":") {
			if !network.Contains(target) {
				errData = &errortypes.ErrorData{
					Error:   "route_target_invalid_network",
					Message: "Route target not in VPC network",
				}
				return
			}
		} else {
			if !network6.Contains(target) {
				errData = &errortypes.ErrorData{
					Error:   "route_target_invalid_network6",
					Message: "Route target not in VPC IPv6 network",
				}
				return
			}
		}
	}

	return
}

func (v *Vpc) Json() {
	netHash := md5.New()
	netHash.Write(v.Id[:])
	netHashSum := fmt.Sprintf("%x", netHash.Sum(nil))[:12]

	ip := fmt.Sprintf("fd97%s", netHashSum)
	ipBuf := bytes.Buffer{}

	for i, run := range ip {
		if i%4 == 0 && i != 0 && i != len(ip)-1 {
			ipBuf.WriteRune(':')
		}
		ipBuf.WriteRune(run)
	}

	v.Network6 = ipBuf.String() + "::/64"
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

func (v *Vpc) GetNetwork6() (network *net.IPNet, err error) {
	netHash := md5.New()
	netHash.Write(v.Id[:])
	netHashSum := fmt.Sprintf("%x", netHash.Sum(nil))[:12]

	ip := fmt.Sprintf("fd97%s", netHashSum)
	ipBuf := bytes.Buffer{}

	for i, run := range ip {
		if i%4 == 0 && i != 0 && i != len(ip)-1 {
			ipBuf.WriteRune(':')
		}
		ipBuf.WriteRune(run)
	}

	_, network, err = net.ParseCIDR(ipBuf.String() + "::/64")
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

func (v *Vpc) GetGateway6() (ip net.IP, err error) {
	network, err := v.GetNetwork6()
	if err != nil {
		return
	}

	ip = network.IP
	utils.IncIpAddress(ip)

	return
}

func (v *Vpc) GetIp(db *database.Database, typ string,
	instId primitive.ObjectID) (

	ip net.IP, err error) {

	coll := db.VpcsIp()
	vpcIp := &VpcIp{}

	err = coll.FindOne(db, &bson.M{
		"vpc":      v.Id,
		"type":     typ,
		"instance": instId,
	}).Decode(vpcIp)
	if err != nil {
		err = database.ParseError(err)
		vpcIp = nil
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		} else {
			return
		}
	}

	if vpcIp == nil {
		vpcIp = &VpcIp{}
		opts := &options.FindOneAndUpdateOptions{}
		opts.SetReturnDocument(options.After)

		err = coll.FindOneAndUpdate(
			db,
			&bson.M{
				"vpc":      v.Id,
				"type":     typ,
				"instance": nil,
			},
			&bson.M{
				"$set": &bson.M{
					"instance": instId,
				},
			},
			opts,
		).Decode(vpcIp)
		if err != nil {
			err = database.ParseError(err)
			vpcIp = nil
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		}
	}

	if vpcIp == nil {
		vpcIp = &VpcIp{}

		sort := 0
		if typ == Gateway {
			sort = 1
		} else {
			sort = -1
		}

		err = coll.FindOne(
			db,
			&bson.M{
				"vpc":  v.Id,
				"type": typ,
			},
			&options.FindOneOptions{
				Sort: &bson.D{
					{"ip", sort},
				},
			},
		).Decode(vpcIp)
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

			_, err = coll.InsertOne(db, vpcIp)
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
	netHash.Write(v.Id[:])
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

func (v *Vpc) PingLink(db *database.Database) (held bool, err error) {
	coll := db.Vpcs()

	query := bson.M{
		"_id":            v.Id,
		"link_timestamp": v.LinkTimestamp,
	}

	if !v.LinkNode.IsZero() {
		query["link_node"] = v.LinkNode
	}

	_, err = coll.UpdateOne(
		db,
		query,
		&bson.M{
			"$set": &bson.M{
				"link_node":      node.Self.Id,
				"link_timestamp": time.Now(),
			},
		},
	)
	if err != nil {
		err = database.ParseError(err)
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		} else {
			return
		}
	} else {
		held = true
	}

	return
}

func (v *Vpc) AddLinkRoutes(db *database.Database, routes []*Route) (
	err error) {

	vc, err := Get(db, v.Id)
	if err != nil {
		return
	}

	linkDsts := set.NewSet()
	for _, route := range routes {
		linkDsts.Add(route.Destination)
	}

	for _, route := range vc.Routes {
		if route.Link || linkDsts.Contains(route.Destination) {
			continue
		}

		routes = append(routes, route)
	}

	vc.Routes = routes

	err = vc.CommitFields(db, set.NewSet("routes"))
	if err != nil {
		return
	}

	return
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

	if !v.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("vpc: Vpc already exists"),
		}
		return
	}

	_, err = coll.InsertOne(db, v)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
