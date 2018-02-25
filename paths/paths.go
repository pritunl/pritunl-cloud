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

func GetDiskPath(diskId bson.ObjectId) string {
	return path.Join(GetDisksPath(),
		fmt.Sprintf("%s.qcow2", diskId.Hex()))
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
