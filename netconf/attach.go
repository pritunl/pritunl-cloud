package netconf

import (
	"github.com/pritunl/pritunl-cloud/vm"
)

func (n *NetConf) SaveAttachments() {
	n.VmAdapter.Attachments = &vm.Attachments{
		External: n.PhysicalExternalIface,
		Internal: n.PhysicalInternalIface,
		Host:     n.PhysicalHostIface,
		NodePort: n.PhysicalNodePortIface,
	}
}

func (n *NetConf) LoadAttachments() {
	attachments := n.VmAdapter.Attachments

	n.PhysicalExternalIface = attachments.External
	n.PhysicalInternalIface = attachments.Internal
	n.PhysicalHostIface = attachments.Host
	n.PhysicalNodePortIface = attachments.NodePort
}
