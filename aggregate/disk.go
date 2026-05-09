package aggregate

import (
	"sync"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/node"
)

type DiskPipe struct {
	disk.Disk `bson:",inline"`
	ImageDocs []*node.Node `bson:"image_docs"`
}

type DisksPipe struct {
	Metadata []*Metadata `bson:"meta"`
	Disks    []*DiskPipe `bson:"disks"`
}

type DiskBackup struct {
	Image bson.ObjectID `json:"image"`
	Name  string        `json:"name"`
}

type DiskAggregate struct {
	disk.Disk
	Backups []*DiskBackup `json:"backups"`
}

func GetDiskPaged(db *database.Database, query *bson.M, page,
	pageCount int64) (disks []*DiskAggregate, count int64, err error) {

	coll := db.Disks()
	disks = []*DiskAggregate{}

	if pageCount == 0 {
		pageCount = 20
	}
	skip := page * pageCount

	addDisk := func(doc *DiskPipe) {
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

	var cursor *mongo.Cursor
	if len(*query) == 0 {
		waiter := &sync.WaitGroup{}
		var countErr error

		waiter.Add(1)
		go func() {
			defer waiter.Done()

			count, countErr = coll.EstimatedDocumentCount(db)
			if countErr != nil {
				countErr = database.ParseError(countErr)
				return
			}
		}()

		cursor, err = coll.Aggregate(db, []*bson.M{
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

			addDisk(doc)
		}

		waiter.Wait()
		if countErr != nil {
			err = countErr
			return
		}
	} else {
		cursor, err = coll.Aggregate(db, []*bson.M{
			&bson.M{
				"$match": query,
			},
			&bson.M{
				"$sort": &bson.M{
					"name": 1,
				},
			},
			&bson.M{
				"$facet": &bson.M{
					"meta": []*bson.M{
						&bson.M{
							"$count": "count",
						},
					},
					"disks": []*bson.M{
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
					},
				},
			},
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}
		defer cursor.Close(db)

		if !cursor.Next(db) {
			err = &database.NotFoundError{
				errors.New("aggregate: Not found"),
			}
			return
		}

		doc := &DisksPipe{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		if len(doc.Metadata) > 0 {
			count = doc.Metadata[0].Count
		}

		for _, diskDoc := range doc.Disks {
			addDisk(diskDoc)
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
