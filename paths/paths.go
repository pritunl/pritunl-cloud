package paths

import (
	"fmt"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"gopkg.in/mgo.v2/bson"
	"path"
)

func GetVmPath(instId bson.ObjectId) string {
	return path.Join(node.Self.GetVirtPath(),
		"instances", instId.Hex())
}

func GetDisksPath() string {
	return path.Join(node.Self.GetVirtPath(), "disks")
}

func GetTempPath() string {
	return path.Join(node.Self.GetVirtPath(), "temp")
}

func GetTempDir() string {
	return path.Join(GetTempPath(), bson.NewObjectId().Hex())
}

func GetDiskPath(diskId bson.ObjectId) string {
	return path.Join(GetDisksPath(),
		fmt.Sprintf("%s.qcow2", diskId.Hex()))
}

func GetDiskTempPath() string {
	return path.Join(GetTempPath(),
		fmt.Sprintf("disk-%s", bson.NewObjectId().Hex()))
}

func GetImageTempPath() string {
	return path.Join(GetTempPath(),
		fmt.Sprintf("image-%s", bson.NewObjectId().Hex()))
}

func GetDiskMountPath() string {
	return path.Join(GetTempPath(), bson.NewObjectId().Hex())
}

func GetInitsPath() string {
	return path.Join(node.Self.GetVirtPath(), "inits")
}

func GetInitPath(instId bson.ObjectId) string {
	return path.Join(GetInitsPath(),
		fmt.Sprintf("%s.iso", instId.Hex()))
}

func GetUnitName(virtId bson.ObjectId) string {
	return fmt.Sprintf("pritunl_cloud_%s.service", virtId.Hex())
}

func GetUnitPath(virtId bson.ObjectId) string {
	return path.Join(settings.Hypervisor.SystemdPath, GetUnitName(virtId))
}

func GetPidPath(virtId bson.ObjectId) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.pid", virtId.Hex()))
}

func GetSockPath(virtId bson.ObjectId) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.sock", virtId.Hex()))
}

func GetGuestPath(virtId bson.ObjectId) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.guest", virtId.Hex()))
}
