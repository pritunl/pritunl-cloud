package netconf

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/oracle"
	"github.com/pritunl/pritunl-cloud/vm"
)

func New(virt *vm.VirtualMachine) *NetConf {
	return &NetConf{
		Virt: virt,
	}
}

func Destroy(db *database.Database, virt *vm.VirtualMachine) (err error) {
	if virt.CloudVnicAttach == "" {
		return
	}

	pv, err := oracle.NewProvider(node.Self.GetOracleAuthProvider())
	if err != nil {
		return
	}

	err = oracle.RemoveVnic(pv, virt.CloudVnicAttach)
	if err != nil {
		return
	}

	return
}
