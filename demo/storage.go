package demo

import (
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Storages = []*storage.Storage{
	{
		Id:        utils.ObjectIdHex("689733b7a7a35eae0dbaea15"),
		Name:      "pritunl-images",
		Comment:   "",
		Type:      "public",
		Endpoint:  "images.pritunl.com",
		Bucket:    "stable",
		AccessKey: "",
		SecretKey: "",
		Insecure:  false,
	},
	{
		Id:        utils.ObjectIdHex("689733b7a7a35eae0dbaea16"),
		Name:      "pritunl-storage",
		Comment:   "",
		Type:      "private",
		Endpoint:  "s3.amazonaws.com",
		Bucket:    "pritunl-cloud-2943",
		AccessKey: "AKIAJTVJ15RORHDU7M1M",
		SecretKey: "VLBGHOVTKDP5SIRSEC8R4XFQWLCIYN4HK",
		Insecure:  false,
	},
}
