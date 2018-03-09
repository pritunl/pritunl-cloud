package vpc

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"net"
)

type Vpc struct {
	Id           bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Name         string        `bson:"name" json:"name"`
	VpcId        int           `bson:"vpc_id" json:"vpc_id"`
	Network      string        `bson:"network" json:"network"`
	Organization bson.ObjectId `bson:"organization" json:"organization"`
	Datacenter   bson.ObjectId `bson:"datacenter" json:"datacenter"`
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
	v.VpcId = rand.Intn(16777100) + 110
}

func (v *Vpc) GetIp(db *database.Database, instId bson.ObjectId) (
	ip net.IP, err error) {

	coll := db.VpcsIp()
	vpcIp := &VpcIp{}

	err = coll.FindOne(&bson.M{
		"vpc":      v.Id,
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

		err = coll.Find(&bson.M{
			"vpc": v.Id,
		}).Sort("-ip").One(vpcIp)
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
		if vpcIp == nil {
			curIp = utils.CopyIpAddress(network.IP)
			utils.IncIpAddress(curIp)
		} else {
			curIp = utils.Int2IpAddress(vpcIp.Ip)
		}

		for {
			utils.IncIpAddress(curIp)

			vpcIp = &VpcIp{
				Vpc:      v.Id,
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
