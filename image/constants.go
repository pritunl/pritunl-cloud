package image

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/tools/set"
)

const (
	Uefi    = "uefi"
	Bios    = "bios"
	Unknown = "unknown"

	Linux         = "linux"
	LinuxLegacy   = "linux_legacy"
	LinuxUnsigned = "linux_unsigned"
	Bsd           = "bsd"

	AlmaLinux8    = "almalinux8"
	AlmaLinux9    = "almalinux9"
	AlmaLinux10   = "almalinux10"
	AlpineLinux   = "alpinelinux"
	Fedora42      = "fedora42"
	FreeBSD       = "freebsd"
	OracleLinux7  = "oraclelinux7"
	OracleLinux8  = "oraclelinux8"
	OracleLinux9  = "oraclelinux9"
	OracleLinux10 = "oraclelinux10"
	RockyLinux8   = "rockylinux8"
	RockyLinux9   = "rockylinux9"
	RockyLinux10  = "rockylinux10"
	Ubuntu2404    = "ubuntu2404"
)

var (
	Global   = bson.NilObjectID
	Releases = set.NewSet(
		AlmaLinux8,
		AlmaLinux9,
		AlmaLinux10,
		AlpineLinux,
		Fedora42,
		FreeBSD,
		OracleLinux7,
		OracleLinux8,
		OracleLinux9,
		OracleLinux10,
		RockyLinux8,
		RockyLinux9,
		RockyLinux10,
		Ubuntu2404,
	)
)
