package aggregate

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
)

type DiskPipe struct {
	disk.Disk `bson:",inline"`
	ImageDocs []*node.Node `bson:"image_docs"`
}

type DiskBackup struct {
	Image primitive.ObjectID `json:"image"`
	Name  string             `json:"name"`
}

type DiskAggregate struct {
	disk.Disk
	Backups []*DiskBackup `json:"backups"`
}

func GetDiskPaged(db *database.Database, query *bson.M, page,
	pageCount int64) (disks []*DiskAggregate, count int64, err error) {

	coll := db.Disks()
	disks = []*DiskAggregate{}

	count, err = coll.Count(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	page = utils.Min64(page, count/pageCount)
	skip := utils.Min64(page*pageCount, count)

	cursor, err := coll.Aggregate(db, []*bson.M{
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
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		doc := &DiskPipe{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

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

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
