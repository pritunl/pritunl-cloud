package qemu

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/dhcps"
	"github.com/pritunl/pritunl-cloud/netconf"
	"github.com/pritunl/pritunl-cloud/vm"
)

func NetworkConfClear(db *database.Database,
	virt *vm.VirtualMachine) (err error) {

	err = dhcps.Stop(virt)
	if err != nil {
		return
	}

	nc := netconf.New(virt)
	err = nc.Clean(db)
	if err != nil {
		return
	}

	return
}

func NetworkConf(db *database.Database,
	virt *vm.VirtualMachine) (err error) {

	nc := netconf.New(virt)
	err = nc.Init(db)
	if err != nil {
		return
	}

	return
}
