package demo

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/audit"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/log"
	"github.com/pritunl/pritunl-cloud/session"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/subscription"
	"github.com/pritunl/pritunl-cloud/user"
	"github.com/pritunl/pritunl-cloud/useragent"
	"github.com/pritunl/pritunl-cloud/utils"
)

func IsDemo() bool {
	return settings.System.Demo
}

func Blocked(c *gin.Context) bool {
	if !IsDemo() {
		return false
	}

	errData := &errortypes.ErrorData{
		Error:   "demo_unavailable",
		Message: "Not available in demo mode",
	}
	c.JSON(400, errData)

	return true
}

func BlockedSilent(c *gin.Context) bool {
	if !IsDemo() {
		return false
	}

	c.JSON(200, nil)
	return true
}

// Users
var Users = []*user.User{
	&user.User{
		Id:            utils.ObjectIdHex("5b6cd11857e4a9a88cbf072e"),
		Type:          "local",
		Provider:      primitive.ObjectID{},
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
		Provider:      primitive.ObjectID{},
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

var Logs = []*log.Entry{
	&log.Entry{
		Id:        utils.ObjectIdHex("5a18e6ae051a45ffac0e5b67"),
		Level:     log.Info,
		Timestamp: time.Unix(1498018860, 0),
		Message:   "router: Starting redirect server",
		Stack:     "",
		Fields: map[string]interface{}{
			"port":       80,
			"production": true,
			"protocol":   "http",
		},
	},
	&log.Entry{
		Id:        utils.ObjectIdHex("5a190b42051a45ffac129bbc"),
		Level:     log.Info,
		Timestamp: time.Unix(1498018860, 0),
		Message:   "router: Starting web server",
		Stack:     "",
		Fields: map[string]interface{}{
			"port":       443,
			"production": true,
			"protocol":   "https",
		},
	},
}

var Subscription = &subscription.Subscription{
	Active:            true,
	Status:            "active",
	Plan:              "zero",
	Quantity:          1,
	Amount:            5000,
	PeriodEnd:         time.Unix(1893499200, 0),
	TrialEnd:          time.Time{},
	CancelAtPeriodEnd: false,
	Balance:           0,
	UrlKey:            "demo",
}
