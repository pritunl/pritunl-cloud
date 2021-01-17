package paths

import (
	"fmt"
	"path"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
)

func GetVmPath(instId primitive.ObjectID) string {
	return path.Join(node.Self.GetVirtPath(),
		"instances", instId.Hex())
}

func GetDisksPath() string {
	return path.Join(node.Self.GetVirtPath(), "disks")
}

func GetBackingPath() string {
	return path.Join(node.Self.GetVirtPath(), "backing")
}

func GetTempPath() string {
	return path.Join(node.Self.GetVirtPath(), "temp")
}

func GetTempDir() string {
	return path.Join(GetTempPath(), primitive.NewObjectID().Hex())
}

func GetOvmfDir() string {
	return path.Join(node.Self.GetVirtPath(), "ovmf")
}

func GetDiskPath(diskId primitive.ObjectID) string {
	return path.Join(GetDisksPath(),
		fmt.Sprintf("%s.qcow2", diskId.Hex()))
}

func GetOvmfVarsPath(virtId primitive.ObjectID) string {
	return path.Join(GetOvmfDir(),
		fmt.Sprintf("%s_vars.fd", virtId.Hex()))
}

func GetDiskTempPath() string {
	return path.Join(GetTempPath(),
		fmt.Sprintf("disk-%s", primitive.NewObjectID().Hex()))
}

func GetImageTempPath() string {
	return path.Join(GetTempPath(),
		fmt.Sprintf("image-%s", primitive.NewObjectID().Hex()))
}

func GetDiskMountPath() string {
	return path.Join(GetTempPath(), primitive.NewObjectID().Hex())
}

func GetInitsPath() string {
	return path.Join(node.Self.GetVirtPath(), "inits")
}

func GetInitPath(instId primitive.ObjectID) string {
	return path.Join(GetInitsPath(),
		fmt.Sprintf("%s.iso", instId.Hex()))
}

func GetLeasesPath() string {
	return path.Join(node.Self.GetVirtPath(), "leases")
}

func GetLeasePath(instId primitive.ObjectID) string {
	return path.Join(GetLeasesPath(),
		fmt.Sprintf("%s.leases", instId.Hex()))
}

func GetLinkLeasePath() string {
	return path.Join(GetLeasesPath(), "link.leases")
}

func GetUnitName(virtId primitive.ObjectID) string {
	return fmt.Sprintf("pritunl_cloud_%s.service", virtId.Hex())
}

func GetUnitPath(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.SystemdPath, GetUnitName(virtId))
}

func GetPidPath(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.pid", virtId.Hex()))
}

func GetSockPath(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.sock", virtId.Hex()))
}

func GetQmpSockPath(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.qmp.sock", virtId.Hex()))
}

func GetGuestPath(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.guest", virtId.Hex()))
}

func GetNamespacesPath() string {
	return "/etc/netns"
}

func GetNamespacePath(namespace string) string {
	return path.Join(GetNamespacesPath(), namespace)
}
