package aggregate

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
)

type DiskPipe struct {
	disk.Disk `bson:",inline"`
	ImageDocs []*node.Node `bson:"image_docs"`
}

type DiskBackup struct {
	Image bson.ObjectId `json:"image"`
	Name  string        `json:"name"`
}

type DiskAggregate struct {
	disk.Disk
	Backups []*DiskBackup `json:"backups"`
}

func GetDiskPaged(db *database.Database, query *bson.M, page,
	pageCount int) (disks []*DiskAggregate, count int, err error) {

	coll := db.Disks()
	disks = []*DiskAggregate{}

	qury := coll.Find(query)

	count, err = qury.Count()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	page = utils.Min(page, count/pageCount)
	skip := utils.Min(page*pageCount, count)

	pipe := coll.Pipe([]*bson.M{
		&bson.M{
			"$match": query,
		},
		&bson.M{
			"$sort": &bson.M{
				"name": 1,
			},
		},
		&bson.M{
			"$skip": skip,
		},
		&bson.M{
			"$limit": pageCount,
		},
		&bson.M{
			"$lookup": &bson.M{
				"from":         "images",
				"localField":   "_id",
				"foreignField": "disk",
				"as":           "image_docs",
			},
		},
	})

	resp := []*DiskPipe{}
	err = pipe.All(&resp)
	if err != nil {
		return
	}

	for _, doc := range resp {
		backups := []*DiskBackup{}

		for _, img := range doc.ImageDocs {
			backup := &DiskBackup{
				Image: img.Id,
				Name:  img.Name,
			}

			backups = append(backups, backup)
		}

		dsk := &DiskAggregate{
			Disk:    doc.Disk,
			Backups: backups,
		}

		disks = append(disks, dsk)
	}

	return
}
