package netconf

import (
	"github.com/pritunl/pritunl-cloud/vm"
)

func New(virt *vm.VirtualMachine) *NetConf {
	return &NetConf{
		Virt: virt,
	}
}
