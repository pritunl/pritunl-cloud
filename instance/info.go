package instance

import (
	"reflect"
	"sync"
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/drive"
	"github.com/pritunl/pritunl-cloud/iso"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/pci"
	"github.com/pritunl/pritunl-cloud/usb"
	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	infos         = map[bson.ObjectID]*Info{}
	infosLock     = sync.Mutex{}
	infosSynced   = map[bson.ObjectID]*infoSynced{}
	infosSyncLock = sync.Mutex{}
)

type infoSynced struct {
	info      *Info
	timestamp time.Time
}

type Info struct {
	Id            bson.ObjectID       `bson:"_id,omitempty" json:"-"`
	Node          string              `bson:"node" json:"node"`
	NodePublicIp  string              `bson:"node_public_ip" json:"node_public_ip"`
	Mtu           int                 `bson:"mtu" json:"mtu"`
	Iscsi         bool                `bson:"iscsi" json:"iscsi"`
	Disks         []string            `bson:"disks" json:"disks"`
	FirewallRules map[string]string   `bson:"firewall_rules" json:"firewall_rules"`
	Authorities   []string            `bson:"authorities" json:"authorities"`
	Isos          []*iso.Iso          `bson:"isos" json:"isos"`
	UsbDevices    []*usb.Device       `bson:"usb_devices" json:"usb_devices"`
	PciDevices    []*pci.Device       `bson:"pci_devices" json:"pci_devices"`
	DriveDevices  []*drive.Device     `bson:"drive_devices" json:"drive_devices"`
	CloudSubnets  []*node.CloudSubnet `bson:"cloud_subnets" json:"cloud_subnets"`
	Timestamp     time.Time           `bson:"timestamp" json:"timestamp"`
}

func (i *Info) Upsert(db *database.Database, instId bson.ObjectID,
	timestamp time.Time) (err error) {

	coll := db.InstancesInfo()

	doc := *i
	doc.Id = bson.NilObjectID
	doc.Timestamp = timestamp

	err = coll.Upsert(&bson.M{
		"_id": instId,
	}, &doc)
	if err != nil {
		return
	}

	return
}

func InfoLookup() *bson.M {
	return &bson.M{
		"$lookup": &bson.M{
			"from":         "instances_info",
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "info_docs",
		},
	}
}

func GetInfo(db *database.Database, instId bson.ObjectID) (
	inf *Info, err error) {

	coll := db.InstancesInfo()
	inf = &Info{}

	err = coll.FindOneId(instId, inf)
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
			inf = &Info{}
			return
		}
		return
	}

	return
}

func RemoveInfo(db *database.Database, instId bson.ObjectID) (
	err error) {

	coll := db.InstancesInfo()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": instId,
	})
	if err != nil {
		err = database.ParseError(err)
		switch err.(type) {
		case *database.NotFoundError:
			err = nil
		default:
			return
		}
	}

	return
}

func SetInfos(newInfos map[bson.ObjectID]*Info) {
	infosLock.Lock()
	infos = newInfos
	infosLock.Unlock()
}

func SyncInfos(db *database.Database, ttl time.Duration) (err error) {
	infosLock.Lock()
	curInfos := infos
	infosLock.Unlock()

	infosSyncLock.Lock()
	defer infosSyncLock.Unlock()

	now := time.Now()

	for instId, inf := range curInfos {
		synced := infosSynced[instId]
		if synced != nil && now.Sub(synced.timestamp) < ttl &&
			reflect.DeepEqual(synced.info, inf) {

			continue
		}

		err = inf.Upsert(db, instId, now)
		if err != nil {
			return
		}

		infosSynced[instId] = &infoSynced{
			info: inf,
			timestamp: now.Add(-time.Duration(
				utils.RandInt(0, 10)) * time.Second),
		}
	}

	for instId := range infosSynced {
		if _, ok := curInfos[instId]; !ok {
			delete(infosSynced, instId)
		}
	}

	return
}
