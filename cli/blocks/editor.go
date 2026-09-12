package blocks

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/cli/form"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/sirupsen/logrus"
)

// commitFields are the fields written by the admin block handler.
var commitFields = set.NewSet(
	"name",
	"comment",
	"vlan",
	"subnets",
	"subnets6",
	"excludes",
	"netmask",
	"gateway",
	"gateway6",
)

var typeOptions = []widget.SelectOption{
	{Label: "IPv4", Value: block.IPv4},
	{Label: "IPv6", Value: block.IPv6},
}

// editor is the block settings form mirroring the inputs of the
// detailed block view in the web interface, the network mode can only
// be chosen when creating a block.
type editor struct {
	blck   *aggregate.BlockAggregate
	create bool
	form   *form.Form

	name     *form.Text
	comment  *form.Area
	typ      *form.Select
	vlan     *form.Number
	netmask  *form.Text
	subnets  *form.Tokens
	subnets6 *form.Tokens
	gateway  *form.Text
	gateway6 *form.Text
	excludes *form.Tokens
}

func newEditor(blck *aggregate.BlockAggregate) *editor {
	e := &editor{
		blck: blck,
	}

	e.name = form.NewText("Name", "Enter name", blck.Name)
	e.comment = form.NewArea("Comment", "Block comment", blck.Comment, 4)

	e.typ = form.NewSelect("Network Mode", typeOptions, blck.Type)
	e.typ.Lock = func() bool {
		return !e.create
	}

	e.vlan = form.NewNumber("VLAN", "Enter VLAN", blck.Vlan, 0, 4095, 1)

	ipv6 := func() bool {
		return e.typ.Value() == block.IPv6
	}
	ipv4 := func() bool {
		return e.typ.Value() != block.IPv6
	}

	e.netmask = form.NewText("Netmask", "Enter netmask", blck.Netmask)
	e.netmask.Hide = ipv6

	e.subnets = form.NewTokens("IP Addresses", "Add addresses",
		blck.Subnets)
	e.subnets.Hide = ipv6
	e.subnets6 = form.NewTokens("IPv6 Addresses", "Add addresses",
		blck.Subnets6)
	e.subnets6.Hide = ipv4

	e.gateway = form.NewText("Gateway", "Enter gateway", blck.Gateway)
	e.gateway.Hide = ipv6
	e.gateway6 = form.NewText("IPv6 Gateway", "Enter IPv6 gateway",
		blck.Gateway6)
	e.gateway6.Hide = ipv4

	e.excludes = form.NewTokens("IP Excludes", "Add exclude",
		blck.Excludes)
	e.excludes.Hide = ipv6

	e.form = form.New(
		e.name,
		e.comment,
		e.typ,
		e.vlan,
		e.netmask,
		e.subnets,
		e.subnets6,
		e.gateway,
		e.gateway6,
		e.excludes,
	)

	return e
}

func (e *editor) Form() *form.Form {
	return e.form
}

// apply sets the form values on the block.
func (e *editor) apply(blck *block.Block) {
	blck.Name = e.name.Value()
	blck.Comment = e.comment.Value()
	blck.Vlan = e.vlan.Int()
	blck.Subnets = e.subnets.Values()
	blck.Subnets6 = e.subnets6.Values()
	blck.Excludes = e.excludes.Values()
	blck.Netmask = e.netmask.Value()
	blck.Gateway = e.gateway.Value()
	blck.Gateway6 = e.gateway6.Value()
}

// Save applies the form to the current block the same way as the admin
// block handler, or inserts a new block.
func (e *editor) Save(db *database.Database) (err error) {
	var blck *block.Block
	if e.create {
		blck = &block.Block{
			Type: e.typ.Value(),
		}
	} else {
		blck, err = block.Get(db, e.blck.Id)
		if err != nil {
			return
		}
	}

	e.apply(blck)

	errData, err := blck.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = &errortypes.ParseError{
			errors.New("blocks: " + errData.Message),
		}
		return
	}

	if e.create {
		err = blck.Insert(db)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"block_id": blck.Id.Hex(),
		}).Info("tui: Block created")
	} else {
		err = blck.CommitFields(db, commitFields)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"block_id": blck.Id.Hex(),
		}).Info("tui: Block settings saved")
	}

	err = event.PublishDispatch(db, "block.change")
	if err != nil {
		return
	}

	return
}
