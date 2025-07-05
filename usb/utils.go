package usb

import (
	"regexp"
	"strings"

	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
)

var (
	reg = regexp.MustCompile("[^a-z0-9]+")
)

func Available(db *database.Database, instId primitive.ObjectID,
	device *Device) (available bool, err error) {

	coll := db.Instances()

	query := bson.M{}

	if !instId.IsZero() {
		query["_id"] = &bson.M{
			"$ne": instId,
		}
	}

	if device.Vendor != "" && device.Product != "" {
		query["usb_devices"] = bson.M{
			"$elemMatch": bson.M{
				"vendor":  device.Vendor,
				"product": device.Product,
			},
		}
	} else if device.Bus != "" && device.Address != "" {
		query["usb_devices"] = bson.M{
			"$elemMatch": bson.M{
				"bus":     device.Bus,
				"address": device.Address,
			},
		}
	}

	count, err := coll.CountDocuments(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if count == 0 {
		available = true
	}

	return
}

func FilterId(deviceId string) string {
	deviceId = strings.ToLower(deviceId)
	deviceId = reg.ReplaceAllString(deviceId, "")
	if len(deviceId) != 4 {
		return ""
	}
	return deviceId
}

func FilterAddr(addr string) string {
	addr = strings.ToLower(addr)
	addr = reg.ReplaceAllString(addr, "")
	if len(addr) != 3 {
		return ""
	}
	return addr
}
