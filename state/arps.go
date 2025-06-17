package state

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/arp"
	"github.com/pritunl/pritunl-cloud/database"
)

var (
	Arps    = &ArpsState{}
	ArpsPkg = NewPackage(Arps)
)

type ArpsState struct {
	arpRecords map[string]set.Set
}

func (p *ArpsState) ArpRecords(namespace string) set.Set {
	return p.arpRecords[namespace]
}

func (p *ArpsState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	p.arpRecords = arp.BuildState(Instances.Instances(),
		Vpcs.VpcsMap(), Vpcs.VpcIpsMap())

	return
}

func (p *ArpsState) Apply(st *State) {
	st.ArpRecords = p.ArpRecords
}

func init() {
	ArpsPkg.
		After(Instances).
		After(Vpcs)
}
