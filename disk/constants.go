package disk

import "github.com/pritunl/mongo-go-driver/v2/bson"

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

	Linux         = "linux"
	LinuxLegacy   = "linux_legacy"
	LinuxUnsigned = "linux_unsigned"
	Bsd           = "bsd"

	AlpineLinux = "alpinelinux"
	ArchLinux   = "archlinux"
	RedHat      = "redhat"
	Fedora      = "fedora"
	Ubuntu      = "ubuntu"
	FreeBSD     = "freebsd"
)

var (
	Vacant           = bson.NilObjectID
	ValidSystemTypes = set.NewSet(
		Linux,
		LinuxLegacy,
		LinuxUnsigned,
		Bsd,
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
