package zone

import (
	"net"
	"slices"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Zone struct {
	Id           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Datacenter   bson.ObjectID `bson:"datacenter" json:"datacenter"`
	Name         string        `bson:"name" json:"name"`
	Comment      string        `bson:"comment" json:"comment"`
	DnsServers   []string      `bson:"dns_servers" json:"dns_servers"`
	DnsServers6  []string      `bson:"dns_servers6" json:"dns_servers6"`
	AnnounceRate int           `bson:"announce_rate" json:"announce_rate"`
	StartupRate  int           `bson:"startup_rate" json:"startup_rate"`
}

type Completion struct {
	Id         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Datacenter bson.ObjectID `bson:"datacenter" json:"datacenter"`
	Name       string        `bson:"name" json:"name"`
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

	dnsServers := []string{}
	for i, dnsServer := range z.DnsServers {
		if i > 1 {
			break
		}

		ip := net.ParseIP(dnsServer)
		if ip == nil || ip.To4() == nil {
			errData = &errortypes.ErrorData{
				Error:   "invalid_dns_server",
				Message: "DNS IPv4 server address is invalid",
			}
			return
		}
		dnsServers = append(dnsServers, ip.String())
	}
	z.DnsServers = dnsServers

	dnsServers6 := []string{}
	for i, dnsServer := range z.DnsServers6 {
		if i > 1 {
			break
		}

		ip := net.ParseIP(dnsServer)
		if ip == nil || ip.To4() != nil {
			errData = &errortypes.ErrorData{
				Error:   "invalid_dns_server6",
				Message: "DNS IPv6 server address is invalid",
			}
			return
		}
		dnsServers6 = append(dnsServers6, ip.String())
	}
	z.DnsServers6 = dnsServers6

	if z.AnnounceRate < 0 {
		z.AnnounceRate = 0
	} else if z.AnnounceRate > 600 {
		z.AnnounceRate = 600
	}

	if z.StartupRate < 0 {
		z.StartupRate = 0
	} else if z.StartupRate > 600 {
		z.StartupRate = 600
	}

	return
}

func (z *Zone) GetDnsServerPrimary() string {
	if len(z.DnsServers) > 0 {
		return z.DnsServers[0]
	}
	return settings.Hypervisor.DnsServerPrimary
}

func (z *Zone) GetDnsServerSecondary() string {
	if len(z.DnsServers) > 1 {
		return z.DnsServers[1]
	}
	return settings.Hypervisor.DnsServerSecondary
}

func (z *Zone) GetDnsServers() (dnsServers []string) {
	if len(z.DnsServers) > 0 {
		dnsServers = slices.Clone(z.DnsServers)
		return
	}

	dnsPrimary := settings.Hypervisor.DnsServerPrimary
	if dnsPrimary != "" {
		dnsServers = append(dnsServers, dnsPrimary)
	}

	dnsSecondary := settings.Hypervisor.DnsServerSecondary
	if dnsSecondary != "" {
		dnsServers = append(dnsServers, dnsSecondary)
	}

	return
}

func (z *Zone) GetDnsServerPrimary6() string {
	if len(z.DnsServers6) > 0 {
		return z.DnsServers6[0]
	}
	return ""
}

func (z *Zone) GetDnsServerSecondary6() string {
	if len(z.DnsServers6) > 1 {
		return z.DnsServers6[1]
	}
	return ""
}

func (z *Zone) GetDnsServers6() (dnsServers6 []string) {
	if len(z.DnsServers6) > 0 {
		dnsServers6 = slices.Clone(z.DnsServers6)
		return
	}

	return
}

func (z *Zone) GetAnnounceRate() time.Duration {
	if z.AnnounceRate > 0 {
		return time.Duration(z.AnnounceRate) * time.Second
	}
	return time.Duration(settings.Hypervisor.AnnounceRate) * time.Second
}

func (z *Zone) GetStartupRate() time.Duration {
	if z.StartupRate > 0 {
		return time.Duration(z.StartupRate) * time.Second
	}
	return time.Duration(settings.Hypervisor.StartupRate) * time.Second
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

	z.Id = resp.InsertedID.(bson.ObjectID)

	return
}
