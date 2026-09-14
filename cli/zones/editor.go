package zones

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/zone"
	"github.com/sirupsen/logrus"
)

// commitFields are the fields written by the admin zone handler.
var commitFields = set.NewSet(
	"name",
	"comment",
	"dns_servers",
	"dns_servers6",
	"announce_rate",
	"startup_rate",
)

// editor is the zone settings form mirroring the inputs of the detailed
// zone view in the web interface, the datacenter is only chosen when
// creating a zone.
type editor struct {
	zne    *zone.Zone
	create bool
	form   *form.Form

	name          *form.Text
	comment       *form.Area
	datacenter    *form.Select
	dnsPrimary    *form.Text
	dnsSecondary  *form.Text
	dns6Primary   *form.Text
	dns6Secondary *form.Text
	announceRate  *form.Number
	startupRate   *form.Number
}

func index(values []string, i int) string {
	if i < len(values) {
		return values[i]
	}
	return ""
}

func newEditor(zne *zone.Zone, dcs []*database.Named) *editor {
	e := &editor{
		zne: zne,
	}

	e.name = form.NewText("Name", "Enter name", zne.Name)
	e.comment = form.NewArea("Comment", "Zone comment", zne.Comment, 4)

	dcOpts := resource.SelectOptions(dcs)
	if len(dcOpts) > 0 {
		dcOpts = resource.WithSelect("Select Datacenter", dcOpts)
	} else {
		dcOpts = resource.WithSelect("No Datacenters", nil)
	}
	e.datacenter = form.NewSelect("Datacenter", dcOpts,
		resource.IdHex(zne.Datacenter))

	e.dnsPrimary = form.NewText("Primary DNS Server", "Enter DNS server",
		index(zne.DnsServers, 0))
	e.dnsSecondary = form.NewText("Secondary DNS Server",
		"Enter DNS server", index(zne.DnsServers, 1))
	e.dns6Primary = form.NewText("Primary DNS Server IPv6",
		"Enter DNS server", index(zne.DnsServers6, 0))
	e.dns6Secondary = form.NewText("Secondary DNS Server IPv6",
		"Enter DNS server", index(zne.DnsServers6, 1))

	e.announceRate = form.NewNumber("Announce Rate", "Announce rate",
		zne.AnnounceRate, 0, 600, 10)
	e.startupRate = form.NewNumber("Startup Rate", "Startup rate",
		zne.StartupRate, 0, 600, 5)

	inputs := []form.Input{
		e.name,
		e.comment,
	}
	if dcs != nil {
		inputs = append(inputs, e.datacenter)
	}
	inputs = append(inputs,
		e.dnsPrimary,
		e.dnsSecondary,
		e.dns6Primary,
		e.dns6Secondary,
		e.announceRate,
		e.startupRate,
	)

	e.form = form.New(inputs...)

	return e
}

func (e *editor) Form() *form.Form {
	return e.form
}

// servers returns the non empty DNS servers like the web interface.
func servers(values ...string) []string {
	servers := []string{}
	for _, val := range values {
		if val != "" {
			servers = append(servers, val)
		}
	}
	return servers
}

// apply sets the form values on the zone.
func (e *editor) apply(zne *zone.Zone) {
	zne.Name = e.name.Value()
	zne.Comment = e.comment.Value()
	zne.DnsServers = servers(e.dnsPrimary.Value(), e.dnsSecondary.Value())
	zne.DnsServers6 = servers(e.dns6Primary.Value(),
		e.dns6Secondary.Value())
	zne.AnnounceRate = e.announceRate.Int()
	zne.StartupRate = e.startupRate.Int()
}

// Save applies the form to the current zone the same way as the admin
// zone handler, or inserts a new zone.
func (e *editor) Save(db *database.Database) (err error) {
	var zne *zone.Zone
	if e.create {
		zne = &zone.Zone{
			Datacenter: resource.ParseId(e.datacenter.Value()),
		}
	} else {
		zne, err = zone.Get(db, e.zne.Id)
		if err != nil {
			return
		}
	}

	e.apply(zne)

	errData, err := zne.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = &errortypes.ParseError{
			errors.New("zones: " + errData.Message),
		}
		return
	}

	if e.create {
		err = zne.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"zone_id": zne.Id.Hex(),
		}).Info("tui: Zone created")
	} else {
		err = zne.CommitFields(db, commitFields)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"zone_id": zne.Id.Hex(),
		}).Info("tui: Zone settings saved")
	}

	err = event.PublishDispatch(db, "zone.change")
	if err != nil {
		return
	}

	return
}
