package disk

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
)

const (
	Provision = "provision"
	Available = "available"
	Attached  = "attached"

	Snapshot = "snapshot"
	Backup   = "backup"
	Expand   = "expand"
	Restore  = "restore"
	Destroy  = "destroy"

	Qcow2 = "qcow2"
	Lvm   = "lvm"

	Xfs     = "xfs"
	Ext4    = "ext4"
	LvmXfs  = "lvm_xfs"
	LvmExt4 = "lvm_ext4"

	Linux = "linux"
	Bsd   = "bsd"
)

var (
	Vacant = primitive.NilObjectID
)
