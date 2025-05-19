package node

import (
	"github.com/dropbox/godropbox/container/set"
)

const (
	Admin      = "admin"
	User       = "user"
	Balancer   = "balancer"
	Hypervisor = "hypervisor"

	Qemu = "qemu"
	Kvm  = "kvm"

	Std               = "std"
	Vmware            = "vmware"
	Virtio            = "virtio"               // virtio-vga
	VirtioPci         = "virtio_pci"           // virtio-gpu-pci
	VirtioVgaGl       = "virtio_vga_gl"        // virtio-vga-gl
	VirtioGl          = "virtio_gl"            // virtio-gpu-gl
	VirtioGlVulkan    = "virtio_gl_vulkan"     // virtio-gpu-gl,venus=true
	VirtioPciGl       = "virtio_pci_gl"        // virtio-gpu-gl-pci
	VirtioPciGlVulkan = "virtio_pci_gl_vulkan" // virtio-gpu-gl-pci,venus=true

	Sdl = "sdl"
	Gtk = "gtk"

	Disabled  = "disabled"
	Dhcp      = "dhcp"
	DhcpSlaac = "dhcp_slaac"
	Slaac     = "slaac"
	Static    = "static"
	Internal  = "internal"
	Oracle    = "oracle"

	Restart = "restart"

	HostPath = "host_path"
)

var (
	VgaModes = set.NewSet(
		Std,
		Vmware,
		Virtio,
		VirtioPci,
		VirtioVgaGl,
		VirtioGl,
		VirtioGlVulkan,
		VirtioPciGl,
		VirtioPciGlVulkan,
	)
	VgaRenderModes = set.NewSet(
		VirtioPci,
		VirtioVgaGl,
		VirtioGl,
		VirtioGlVulkan,
		VirtioPciGl,
		VirtioPciGlVulkan,
	)
)
