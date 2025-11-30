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
	AlmaLinux11   = "almalinux11"
	AlmaLinux12   = "almalinux12"
	AlmaLinux13   = "almalinux13"
	AlmaLinux14   = "almalinux14"
	AlmaLinux15   = "almalinux15"
	AlmaLinux16   = "almalinux16"
	AlpineLinux   = "alpinelinux"
	ArchLinux     = "archlinux"
	Fedora42      = "fedora42"
	Fedora43      = "fedora43"
	Fedora44      = "fedora44"
	Fedora45      = "fedora45"
	Fedora46      = "fedora46"
	Fedora47      = "fedora47"
	Fedora48      = "fedora48"
	Fedora49      = "fedora49"
	Fedora50      = "fedora50"
	Fedora51      = "fedora51"
	Fedora52      = "fedora52"
	Fedora53      = "fedora53"
	Fedora54      = "fedora54"
	Fedora55      = "fedora55"
	Fedora56      = "fedora56"
	Fedora57      = "fedora57"
	Fedora58      = "fedora58"
	Fedora59      = "fedora59"
	Fedora60      = "fedora60"
	Fedora61      = "fedora61"
	Fedora62      = "fedora62"
	FreeBSD       = "freebsd"
	OracleLinux7  = "oraclelinux7"
	OracleLinux8  = "oraclelinux8"
	OracleLinux9  = "oraclelinux9"
	OracleLinux10 = "oraclelinux10"
	OracleLinux11 = "oraclelinux11"
	OracleLinux12 = "oraclelinux12"
	OracleLinux13 = "oraclelinux13"
	OracleLinux14 = "oraclelinux14"
	OracleLinux15 = "oraclelinux15"
	OracleLinux16 = "oraclelinux16"
	RockyLinux8   = "rockylinux8"
	RockyLinux9   = "rockylinux9"
	RockyLinux10  = "rockylinux10"
	RockyLinux11  = "rockylinux11"
	RockyLinux12  = "rockylinux12"
	RockyLinux13  = "rockylinux13"
	RockyLinux14  = "rockylinux14"
	RockyLinux15  = "rockylinux15"
	RockyLinux16  = "rockylinux16"
	Ubuntu2404    = "ubuntu2404"
	Ubuntu2604    = "ubuntu2604"
	Ubuntu2804    = "ubuntu2804"
	Ubuntu3004    = "ubuntu3004"
	Ubuntu3204    = "ubuntu3204"
	Ubuntu3404    = "ubuntu3404"
	Ubuntu3604    = "ubuntu3604"
	Ubuntu3804    = "ubuntu3804"
	Ubuntu4004    = "ubuntu4004"
	Ubuntu4204    = "ubuntu4204"
	Ubuntu4404    = "ubuntu4404"
)

var (
	Global   = bson.NilObjectID
	Releases = set.NewSet(
		AlmaLinux8,
		AlmaLinux9,
		AlmaLinux10,
		AlmaLinux11,
		AlmaLinux12,
		AlmaLinux13,
		AlmaLinux14,
		AlmaLinux15,
		AlmaLinux16,
		AlpineLinux,
		ArchLinux,
		Fedora42,
		Fedora43,
		Fedora44,
		Fedora45,
		Fedora46,
		Fedora47,
		Fedora48,
		Fedora49,
		Fedora50,
		Fedora51,
		Fedora52,
		Fedora53,
		Fedora54,
		Fedora55,
		Fedora56,
		Fedora57,
		Fedora58,
		Fedora59,
		Fedora60,
		Fedora61,
		Fedora62,
		FreeBSD,
		OracleLinux7,
		OracleLinux8,
		OracleLinux9,
		OracleLinux10,
		OracleLinux11,
		OracleLinux12,
		OracleLinux13,
		OracleLinux14,
		OracleLinux15,
		OracleLinux16,
		RockyLinux8,
		RockyLinux9,
		RockyLinux10,
		RockyLinux11,
		RockyLinux12,
		RockyLinux13,
		RockyLinux14,
		RockyLinux15,
		RockyLinux16,
		Ubuntu2404,
		Ubuntu2604,
		Ubuntu2804,
		Ubuntu3004,
		Ubuntu3204,
		Ubuntu3404,
		Ubuntu3604,
		Ubuntu3804,
		Ubuntu4004,
		Ubuntu4204,
		Ubuntu4404,
	)
)
