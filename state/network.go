package state

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	Network    = &NetworkState{}
	NetworkPkg = NewPackage(Network)
)

type NetworkState struct {
	namespaces    []string
	interfaces    []string
	interfacesSet set.Set
}

func (p *NetworkState) Namespaces() []string {
	return p.namespaces
}

func (p *NetworkState) Interfaces() []string {
	return p.interfaces
}

func (p *NetworkState) HasInterfaces(iface string) bool {
	return p.interfacesSet.Contains(iface)
}

func (p *NetworkState) Refresh(pkg *Package,
	db *database.Database) (err error) {

	namespaces, err := utils.GetNamespaces()
	if err != nil {
		return
	}
	p.namespaces = namespaces

	interfaces, interfacesSet, err := utils.GetInterfacesSet()
	if err != nil {
		return
	}
	p.interfaces = interfaces
	p.interfacesSet = interfacesSet

	return
}

func (p *NetworkState) Apply(st *State) {
	st.Namespaces = p.Namespaces
	st.Interfaces = p.Interfaces
	st.HasInterfaces = p.HasInterfaces
}
