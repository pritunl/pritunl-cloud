package demo

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
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
		Roles: []string{
			"shape-m2",
		},
		Flexible:   true,
		DiskPools:  []bson.ObjectID{},
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
		Roles: []string{
			"shape-m2",
		},
		Flexible:   true,
		DiskPools:  []bson.ObjectID{},
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
		Roles: []string{
			"shape-m2",
		},
		Flexible:   true,
		DiskPools:  []bson.ObjectID{},
		Memory:     8192,
		Processors: 4,
		NodeCount:  1,
	},
	{
		Id:               utils.ObjectIdHex("68f4c1d2aac06d53e8c9c4a1"),
		Name:             "gp1",
		Comment:          "",
		Type:             "disk",
		DeleteProtection: false,
		Roles: []string{
			"shape-gp1",
		},
		DiskType:  "qcow2",
		DisksPath: "/var/lib/pritunl-cloud/disks",
		DiskPools: []bson.ObjectID{},
		NodeCount: 1,
	},
}
