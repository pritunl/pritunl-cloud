package demo

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/aggregate"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Domains = []*aggregate.Domain{
	{
		Domain: domain.Domain{
			Id:            utils.ObjectIdHex("67b8a1d24866ba90e6c45b5b"),
			Name:          "pritunl-com",
			Comment:       "",
			Organization:  utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
			Type:          "cloudflare",
			Secret:        utils.ObjectIdHex("67b89e8d4866ba90e6c459ba"),
			RootDomain:    "pritunl.com",
			LockId:        bson.ObjectID{},
			LockTimestamp: time.Time{},
			LastUpdate:    time.Now(),
		},
		Records: []*domain.Record{
			{
				Id:              utils.ObjectIdHex("68076c9f06fd0087c078dfdc"),
				Domain:          utils.ObjectIdHex("67b8a1d24866ba90e6c45b5b"),
				Resource:        bson.ObjectID{},
				Deployment:      utils.ObjectIdHex("68076bb954e947708aa6d651"),
				Timestamp:       time.Now(),
				DeleteTimestamp: time.Time{},
				SubDomain:       "demo",
				Type:            "A",
				Value:           "10.196.8.2",
				Operation:       "",
			},
			{
				Id:              utils.ObjectIdHex("68076ca306fd0087c078dfdd"),
				Domain:          utils.ObjectIdHex("67b8a1d24866ba90e6c45b5b"),
				Resource:        bson.ObjectID{},
				Deployment:      utils.ObjectIdHex("68076bb954e947708aa6d651"),
				Timestamp:       time.Now(),
				DeleteTimestamp: time.Time{},
				SubDomain:       "cloud",
				Type:            "A",
				Value:           "10.196.8.12",
				Operation:       "",
			},
			{
				Id:              utils.ObjectIdHex("68076ca406fd0087c078dfde"),
				Domain:          utils.ObjectIdHex("67b8a1d24866ba90e6c45b5b"),
				Resource:        bson.ObjectID{},
				Deployment:      utils.ObjectIdHex("68076bb954e947708aa6d651"),
				Timestamp:       time.Now(),
				DeleteTimestamp: time.Time{},
				SubDomain:       "user.cloud",
				Type:            "A",
				Value:           "10.196.8.12",
				Operation:       "",
			},
			{
				Id:              utils.ObjectIdHex("6813705806fd0087c078dfe1"),
				Domain:          utils.ObjectIdHex("67b8a1d24866ba90e6c45b5b"),
				Resource:        bson.ObjectID{},
				Deployment:      utils.ObjectIdHex("68136f7d43b4ac1351f54f0a"),
				Timestamp:       time.Now(),
				DeleteTimestamp: time.Time{},
				SubDomain:       "demo.cloud",
				Type:            "A",
				Value:           "10.196.8.46",
				Operation:       "",
			},
			{
				Id:              utils.ObjectIdHex("681e01394230fad44c6a5140"),
				Domain:          utils.ObjectIdHex("67b8a1d24866ba90e6c45b5b"),
				Resource:        bson.ObjectID{},
				Deployment:      utils.ObjectIdHex("681e01308d67187e275a847a"),
				Timestamp:       time.Now(),
				DeleteTimestamp: time.Time{},
				SubDomain:       "forum",
				Type:            "AAAA",
				Value:           "2001:db8:85a3:42:d5c:82ca:9ed4:854b",
				Operation:       "",
			},
			{
				Id:              utils.ObjectIdHex("683e86d74230fad44c6a514d"),
				Domain:          utils.ObjectIdHex("67b8a1d24866ba90e6c45b5b"),
				Resource:        bson.ObjectID{},
				Deployment:      utils.ObjectIdHex("683dcdf13249b43a9cc5ec70"),
				Timestamp:       time.Now(),
				DeleteTimestamp: time.Time{},
				SubDomain:       "docs",
				Type:            "CNAME",
				Value:           "docs.pritunl.dev",
				Operation:       "",
			},
		},
	},
}
