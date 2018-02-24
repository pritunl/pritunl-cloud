package paths

import (
	"fmt"
	"github.com/pritunl/pritunl-cloud/node"
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
