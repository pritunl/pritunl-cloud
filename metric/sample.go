package metric

import (
	"maps"
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
)

type Sample struct {
	Timestamp time.Time `json:"t"`
	System    *System   `json:"sy,omitempty"`
	Load      *Load     `json:"lo,omitempty"`
	Disk      *Disk     `json:"dk,omitempty"`
	DiskIo    *DiskIo   `json:"di,omitempty"`
	Network   *Network  `json:"nw,omitempty"`
}

func (s *Sample) StaticData() bson.M {
	data := bson.M{}

	if s.System != nil {
		maps.Copy(data, s.System.StaticData())
	}
	if s.Load != nil {
		maps.Copy(data, s.Load.StaticData())
	}
	if s.Disk != nil {
		maps.Copy(data, s.Disk.StaticData())
	}
	if s.DiskIo != nil {
		maps.Copy(data, s.DiskIo.StaticData())
	}
	if s.Network != nil {
		maps.Copy(data, s.Network.StaticData())
	}

	return data
}

func (s *Sample) docs() (docs []Doc) {
	if s.System != nil {
		s.System.Timestamp = s.Timestamp
		docs = append(docs, s.System)
	}
	if s.Load != nil {
		s.Load.Timestamp = s.Timestamp
		docs = append(docs, s.Load)
	}
	if s.Disk != nil {
		s.Disk.Timestamp = s.Timestamp
		docs = append(docs, s.Disk)
	}
	if s.DiskIo != nil {
		s.DiskIo.Timestamp = s.Timestamp
		docs = append(docs, s.DiskIo)
	}
	if s.Network != nil {
		s.Network.Timestamp = s.Timestamp
		docs = append(docs, s.Network)
	}
	return
}

func InsertSamples(db *database.Database, resource bson.ObjectID,
	samples []*Sample) (err error) {

	for _, sample := range samples {
		if sample == nil {
			continue
		}

		for _, doc := range sample.docs() {
			doc.Format(resource)

			coll := doc.GetCollection(db)
			_, e := coll.InsertOne(db, doc)
			if e != nil {
				e = database.ParseError(e)
				if _, ok := e.(*database.DuplicateKeyError); ok {
					continue
				}
				err = e
				return
			}
		}
	}

	return
}
