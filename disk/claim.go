package disk

import (
	"sync"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/sirupsen/logrus"
)

type ClaimLostError struct {
	errors.DropboxError
}

type Claim struct {
	Disk   bson.ObjectID
	Node   bson.ObjectID
	shared bool
	lock   sync.Mutex
	stop   chan struct{}
}

func claimHeartbeat() time.Duration {
	heartbeat := time.Duration(
		settings.Hypervisor.DiskClaimHeartbeat) * time.Second
	if heartbeat <= 0 {
		heartbeat = 30 * time.Second
	}
	return heartbeat
}

func claimTtl() time.Duration {
	ttl := time.Duration(settings.Hypervisor.DiskClaimTtl) * time.Second
	if ttl <= 0 {
		ttl = 90 * time.Second
	}
	return ttl
}

func NewClaim(dsk *Disk) (clm *Claim) {
	clm = &Claim{
		Disk:   dsk.Id,
		Node:   node.Self.Id,
		shared: dsk.Type == Lvm,
	}

	return
}

func (d *Disk) Check(nodeId bson.ObjectID) bool {
	if d.ActionNode.IsZero() || d.ActionNode == nodeId {
		return false
	}

	return time.Since(d.ActionTimestamp) < claimTtl()
}

func (c *Claim) Acquire(db *database.Database, query bson.M,
	fields bson.M) (dsk *Disk, err error) {

	coll := db.Disks()
	now := time.Now()

	filter := bson.M{
		"_id": c.Disk,
	}
	for key, val := range query {
		filter[key] = val
	}

	update := bson.M{}
	for key, val := range fields {
		update[key] = val
	}

	if c.shared {
		filter["$or"] = []bson.M{
			{
				"action_node": bson.M{
					"$in": []interface{}{nil, bson.NilObjectID, c.Node},
				},
			},
			{
				"action_timestamp": bson.M{
					"$lt": now.Add(-claimTtl()),
				},
			},
		}
		update["action_node"] = c.Node
		update["action_timestamp"] = now
	}

	dsk = &Disk{}
	if len(update) == 0 {
		err = coll.FindOne(db, filter).Decode(dsk)
	} else {
		err = coll.FindOneAndUpdate(
			db,
			filter,
			bson.M{
				"$set": update,
			},
			options.FindOneAndUpdate().SetReturnDocument(options.After),
		).Decode(dsk)
	}
	if err != nil {
		dsk = nil
		err = database.ParseError(err)
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		}
		return
	}

	if c.shared {
		c.Start()
	}

	return
}

func (c *Claim) refresh(db *database.Database) (err error) {
	coll := db.Disks()

	resp, err := coll.UpdateOne(db, &bson.M{
		"_id":         c.Disk,
		"action_node": c.Node,
	}, &bson.M{
		"$set": &bson.M{
			"action_timestamp": time.Now(),
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if resp.MatchedCount == 0 {
		err = &ClaimLostError{
			errors.New("disk: Action claim lost"),
		}
		return
	}

	return
}

func (c *Claim) Release(db *database.Database, dsk *Disk,
	fields set.Set) (err error) {

	coll := db.Disks()

	if !c.shared {
		err = coll.CommitFields(dsk.Id, dsk, fields)
		if err != nil {
			return
		}

		return
	}

	c.Stop()

	dsk.ActionNode = bson.NilObjectID
	dsk.ActionTimestamp = time.Time{}

	claimFields := set.NewSet("action_node", "action_timestamp")
	claimFields.Union(fields)

	resp, err := coll.UpdateOne(db, &bson.M{
		"_id":         c.Disk,
		"action_node": c.Node,
	}, database.SelectFieldsAll(dsk, claimFields))
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if resp.MatchedCount == 0 {
		err = &ClaimLostError{
			errors.New("disk: Action claim lost"),
		}
		return
	}

	return
}

func (c *Claim) run(stop chan struct{}) {
	for {
		select {
		case <-stop:
			return
		case <-time.After(claimHeartbeat()):
		}

		db := database.GetDatabase()
		err := c.refresh(db)
		db.Close()
		if err != nil {
			if _, ok := err.(*ClaimLostError); ok {
				select {
				case <-stop:
					return
				default:
				}

				logrus.WithFields(logrus.Fields{
					"disk_id": c.Disk.Hex(),
				}).Warn("disk: Action claim lost")
				return
			}

			logrus.WithFields(logrus.Fields{
				"disk_id": c.Disk.Hex(),
				"error":   err,
			}).Warn("disk: Failed to refresh action claim")
		}
	}
}

func (c *Claim) Start() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.stop != nil {
		return
	}

	c.stop = make(chan struct{})
	go c.run(c.stop)
}

func (c *Claim) Stop() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.stop == nil {
		return
	}

	close(c.stop)
	c.stop = nil
}
