package metric

import (
	"context"
	"encoding/binary"
	"hash/fnv"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type Doc interface {
	GetCollection(*database.Database) *database.Collection
	Format(bson.ObjectID) time.Time
	StaticData() bson.M
}

type Point struct {
	X int64       `json:"x"`
	Y interface{} `json:"y"`
}

type ChartData = map[string][]*Point

func GenerateId(resourceId bson.ObjectID,
	timestamp time.Time) bson.ObjectID {

	var b [12]byte

	hash := fnv.New64a()
	hash.Write(resourceId[:])
	sum := hash.Sum(nil)

	binary.BigEndian.PutUint32(b[0:4], uint32(timestamp.Unix()))
	copy(b[4:12], sum[:])

	return b
}

func GetObj(typ string) Doc {
	switch typ {
	case "system":
		return &System{}
	case "load":
		return &Load{}
	case "disk":
		return &Disk{}
	case "diskio":
		return &DiskIo{}
	case "network":
		return &Network{}
	default:
		return nil
	}
}

func GetChart(c context.Context, db *database.Database,
	resource bson.ObjectID, typ string, start, end time.Time,
	interval time.Duration) (ChartData, error) {

	start = start.Add(time.Duration(start.UnixMilli()%
		interval.Milliseconds()) * -time.Millisecond)
	end = end.Add(time.Duration(end.UnixMilli()%
		interval.Milliseconds()) * -time.Millisecond)

	switch typ {
	case "system":
		return GetSystemChart(c, db, resource, start, end, interval)
	case "load":
		return GetLoadChart(c, db, resource, start, end, interval)
	case "disk":
		return GetDiskChart(c, db, resource, start, end, interval)
	case "diskio":
		return GetDiskIoChart(c, db, resource, start, end, interval)
	case "network":
		return GetNetworkChart(c, db, resource, start, end, interval)
	default:
		return nil, &errortypes.UnknownError{
			errors.New("metric: Unknown resource type"),
		}
	}
}

type Chart struct {
	start    int64
	end      int64
	intv     int64
	data     ChartData
	curTimes map[string]int64
}

func (c *Chart) add(resource string, timestamp int64, value interface{}) {
	c.data[resource] = append(c.data[resource], &Point{
		X: timestamp,
		Y: value,
	})
}

func (c *Chart) Add(resource string, timestamp int64, value interface{}) {
	cur := c.curTimes[resource]
	if cur == 0 {
		cur = c.start - c.intv
	}

	for timestamp-c.intv > cur {
		cur += c.intv
		c.add(resource, cur, 0)
	}

	c.add(resource, timestamp, value)
	c.curTimes[resource] = timestamp
}

func (c *Chart) Export() map[string][]*Point {
	for resource, cur := range c.curTimes {
		for c.end > cur {
			cur += c.intv
			c.add(resource, cur, 0)
		}
	}

	return c.data
}

func NewChart(start, end time.Time, interval time.Duration) (chrt *Chart) {
	chrt = &Chart{
		start:    start.UnixMilli(),
		end:      end.UnixMilli(),
		intv:     interval.Milliseconds(),
		data:     ChartData{},
		curTimes: map[string]int64{},
	}

	if interval == time.Minute {
		chrt.end -= time.Minute.Milliseconds()
	}

	return
}
