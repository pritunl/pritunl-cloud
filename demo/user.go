package demo

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/audit"
	"github.com/pritunl/pritunl-cloud/session"
	"github.com/pritunl/pritunl-cloud/user"
	"github.com/pritunl/pritunl-cloud/useragent"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Users = []*user.User{
	&user.User{
		Id:            utils.ObjectIdHex("5b6cd11857e4a9a88cbf072e"),
		Type:          "local",
		Provider:      bson.ObjectID{},
		Username:      "demo",
		Token:         "",
		Secret:        "",
		LastActive:    time.Now(),
		LastSync:      time.Now(),
		Roles:         []string{"demo"},
		Administrator: "super",
		Disabled:      false,
		ActiveUntil:   time.Time{},
		Permissions:   []string{},
	},
	&user.User{
		Id:            utils.ObjectIdHex("5a7542190accad1a8a53b568"),
		Type:          "local",
		Provider:      bson.ObjectID{},
		Username:      "pritunl",
		Token:         "",
		Secret:        "",
		LastActive:    time.Time{},
		LastSync:      time.Time{},
		Roles:         []string{},
		Administrator: "super",
		Disabled:      false,
		ActiveUntil:   time.Time{},
		Permissions:   []string{},
	},
}

var Agent = &useragent.Agent{
	OperatingSystem: useragent.Linux,
	Browser:         useragent.Chrome,
	Ip:              "8.8.8.8",
	Isp:             "Google",
	Continent:       "North America",
	ContinentCode:   "NA",
	Country:         "United States",
	CountryCode:     "US",
	Region:          "Washington",
	RegionCode:      "WA",
	City:            "Seattle",
	Latitude:        47.611,
	Longitude:       -122.337,
}

var Audits = []*audit.Audit{
	&audit.Audit{
		Id:        utils.ObjectIdHex("5a17f9bf051a45ffacf2b352"),
		Timestamp: time.Unix(1498018860, 0),
		Type:      "admin_login",
		Fields: audit.Fields{
			"method": "local",
		},
		Agent: Agent,
	},
}

var Sessions = []*session.Session{
	&session.Session{
		Id:         "jhgRu4n3oY0iXRYmLb77Ql5jNs2o7uWM",
		Type:       session.User,
		Timestamp:  time.Unix(1498018860, 0),
		LastActive: time.Unix(1498018860, 0),
		Removed:    false,
		Agent:      Agent,
	},
}
