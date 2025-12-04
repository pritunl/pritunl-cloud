package instance

import (
	"github.com/dropbox/godropbox/container/set"
)

const (
	Starting     = "starting"
	Running      = "running"
	Stopped      = "stopped"
	Failed       = "failed"
	Updating     = "updating"
	Provisioning = "provisioning"
	Bridge       = "bridge"
	Vxlan        = "vxlan"

	Start       = "start"
	Stop        = "stop"
	Cleanup     = "cleanup"
	Restart     = "restart"
	Destroy     = "destroy"
	Linux       = "linux"
	LinuxLegacy = "linux_legacy"
	BSD         = "bsd"

	AlpineLinux = "alpinelinux"
	ArchLinux   = "archlinux"
	RedHat      = "redhat"
	Fedora      = "fedora"
	Ubuntu      = "ubuntu"
	FreeBSD     = "freebsd"

	HostPath = "host_path"
)

var (
	ValidStates = set.NewSet(
		Starting,
		Running,
		Stopped,
		Failed,
		Updating,
		Provisioning,
		Bridge,
		Vxlan,
	)
	ValidActions = set.NewSet(
		Start,
		Stop,
		Cleanup,
		Restart,
		Destroy,
	)
	ValidCloudTypes = set.NewSet(
		Linux,
		LinuxLegacy,
		BSD,
	)
	ValidSystemKinds = set.NewSet(
		AlpineLinux,
		ArchLinux,
		RedHat,
		Fedora,
		Ubuntu,
		FreeBSD,
	)
)
