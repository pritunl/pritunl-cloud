package image

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
)

const (
	Uefi    = "uefi"
	Bios    = "bios"
	Unknown = "unknown"

	Linux         = "linux"
	LinuxUnsigned = "linux_unsigned"
	Bsd           = "bsd"
)

var (
	Global = primitive.NilObjectID
)
