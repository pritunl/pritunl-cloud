package zone

import (
	"net"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Zone struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Datacenter  primitive.ObjectID `bson:"datacenter" json:"datacenter"`
	Name        string             `bson:"name" json:"name"`
	Comment     string             `bson:"comment" json:"comment"`
	DnsServers  []string           `bson:"dns_servers" json:"dns_servers"`
	DnsServers6 []string           `bson:"dns_servers6" json:"dns_servers6"`
}

func (z *Zone) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	z.Name = utils.FilterName(z.Name)

	if z.Datacenter.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "datacenter_required",
			Message: "Missing required datacenter",
		}
		return
	}

	for i, dnsServer := range z.DnsServers {
		ip := net.ParseIP(dnsServer)
		if ip == nil || ip.To4() == nil {
			errData = &errortypes.ErrorData{
				Error:   "invalid_dns_server",
				Message: "DNS IPv4 server address is invalid",
			}
			return
		}
		z.DnsServers[i] = ip.String()
	}

	for i, dnsServer := range z.DnsServers6 {
		ip := net.ParseIP(dnsServer)
		if ip == nil || ip.To4() != nil {
			errData = &errortypes.ErrorData{
				Error:   "invalid_dns_server6",
				Message: "DNS IPv6 server address is invalid",
			}
			return
		}
		z.DnsServers6[i] = ip.String()
	}

	return
}

func (z *Zone) Commit(db *database.Database) (err error) {
	coll := db.Zones()

	err = coll.Commit(z.Id, z)
	if err != nil {
		return
	}

	return
}

func (z *Zone) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Zones()

	err = coll.CommitFields(z.Id, z, fields)
	if err != nil {
		return
	}

	return
}

func (z *Zone) Insert(db *database.Database) (err error) {
	coll := db.Zones()

	if !z.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("zone: Zone already exists"),
		}
		return
	}

	resp, err := coll.InsertOne(db, z)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	z.Id = resp.InsertedID.(primitive.ObjectID)

	return
}
