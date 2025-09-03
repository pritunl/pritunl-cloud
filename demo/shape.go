package demo

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Shapes = []*shape.Shape{
	{
		Id:               utils.ObjectIdHex("65e6e303ceeebbb3dabaec96"),
		Name:             "m2-small",
		Comment:          "",
		Type:             "instance",
		DeleteProtection: false,
		Datacenter:       utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Roles: []string{
			"shape-m2",
		},
		Flexible:   true,
		DiskType:   "qcow2",
		DiskPool:   primitive.ObjectID{},
		Memory:     2048,
		Processors: 1,
		NodeCount:  1,
	},
	{
		Id:               utils.ObjectIdHex("65e6e2ecceeebbb3dabaec79"),
		Name:             "m2-medium",
		Comment:          "",
		Type:             "instance",
		DeleteProtection: false,
		Datacenter:       utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Roles: []string{
			"shape-m2",
		},
		Flexible:   true,
		DiskType:   "qcow2",
		DiskPool:   primitive.ObjectID{},
		Memory:     4096,
		Processors: 2,
		NodeCount:  1,
	},
	{
		Id:               utils.ObjectIdHex("66f63282aac06d53e8c9c435"),
		Name:             "m2-large",
		Comment:          "",
		Type:             "instance",
		DeleteProtection: false,
		Datacenter:       utils.ObjectIdHex("689733b7a7a35eae0dbaea1b"),
		Roles: []string{
			"shape-m2",
		},
		Flexible:   true,
		DiskType:   "qcow2",
		DiskPool:   primitive.ObjectID{},
		Memory:     8192,
		Processors: 4,
		NodeCount:  1,
	},
}
