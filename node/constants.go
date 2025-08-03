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

	// std
	Std = "std"
	// vmware
	Vmware = "vmware"
	// virtio-vga
	Virtio = "virtio"
	// virtio-gpu-pci
	VirtioPci = "virtio_pci"
	// virtio-vga-gl
	VirtioVgaGl = "virtio_vga_gl"
	// virtio-gpu-gl
	VirtioGl = "virtio_gl"
	// virtio-gpu-gl,venus=true
	VirtioGlVulkan = "virtio_gl_vulkan"
	// virtio-gpu-gl-pci
	VirtioPciGl = "virtio_pci_gl"
	// virtio-gpu-gl-pci,venus=true
	VirtioPciGlVulkan = "virtio_pci_gl_vulkan"
	// virtio-vga prime=1
	VirtioPrime = "virtio_prime"
	// virtio-gpu-pci prime=1
	VirtioPciPrime = "virtio_pci_prime"
	// virtio-vga-gl prime=1
	VirtioVgaGlPrime = "virtio_vga_gl_prime"
	// virtio-gpu-gl prime=1
	VirtioGlPrime = "virtio_gl_prime"
	// virtio-gpu-gl,venus=true prime=1
	VirtioGlVulkanPrime = "virtio_gl_vulkan_prime"
	// virtio-gpu-gl-pci prime=1
	VirtioPciGlPrime = "virtio_pci_gl_prime"
	// virtio-gpu-gl-pci,venus=true prime=1
	VirtioPciGlVulkanPrime = "virtio_pci_gl_vulkan_prime"

	Sdl = "sdl"
	Gtk = "gtk"

	Disabled  = "disabled"
	Dhcp      = "dhcp"
	DhcpSlaac = "dhcp_slaac"
	Slaac     = "slaac"
	Static    = "static"
	Internal  = "internal"
	Cloud     = "cloud"

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
		VirtioPrime,
		VirtioPciPrime,
		VirtioVgaGlPrime,
		VirtioGlPrime,
		VirtioGlVulkanPrime,
		VirtioPciGlPrime,
		VirtioPciGlVulkanPrime,
	)
	VgaRenderModes = set.NewSet(
		VirtioPci,
		VirtioVgaGl,
		VirtioGl,
		VirtioGlVulkan,
		VirtioPciGl,
		VirtioPciGlVulkan,
		VirtioPciPrime,
		VirtioVgaGlPrime,
		VirtioGlPrime,
		VirtioGlVulkanPrime,
		VirtioPciGlPrime,
		VirtioPciGlVulkanPrime,
	)
)
