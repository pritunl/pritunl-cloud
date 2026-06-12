package aggregate

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/utils"
)

type DiskPipe struct {
	disk.Disk    `bson:",inline"`
	ImageDocs    []*node.Node         `bson:"image_docs"`
	InstanceDocs []*instance.Instance `bson:"instance_docs"`
}

type DiskBackup struct {
	Image bson.ObjectID `json:"image"`
	Name  string        `json:"name"`
}

type DiskInstanceInfo struct {
	Id              bson.ObjectID `json:"id"`
	Name            string        `json:"name"`
	Status          string        `json:"status"`
	Timestamp       time.Time     `json:"timestamp"`
	Uptime          string        `json:"uptime"`
	PublicIps       []string      `json:"public_ips"`
	PublicIps6      []string      `json:"public_ips6"`
	PrivateIps      []string      `json:"private_ips"`
	PrivateIps6     []string      `json:"private_ips6"`
	CloudPublicIps  []string      `json:"cloud_public_ips"`
	CloudPublicIps6 []string      `json:"cloud_public_ips6"`
}

type DiskAggregate struct {
	disk.Disk
	Backups      []*DiskBackup     `json:"backups"`
	InstanceInfo *DiskInstanceInfo `json:"instance_info"`
}

func GetDiskPaged(db *database.Database, query *bson.M, page,
	pageCount int64) (disks []*DiskAggregate, count int64, err error) {

	coll := db.Disks()
	disks = []*DiskAggregate{}

	if len(*query) == 0 {
		count, err = coll.EstimatedDocumentCount(db)
		if err != nil {
			err = database.ParseError(err)
			return
		}
	} else {
		count, err = coll.CountDocuments(db, query)
		if err != nil {
			err = database.ParseError(err)
			return
		}
	}

	if pageCount == 0 {
		pageCount = 20
	}
	maxPage := count / pageCount
	if count == pageCount {
		maxPage = 0
	}
	page = utils.Min64(page, maxPage)
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
		&bson.M{
			"$lookup": &bson.M{
				"from": "instances",
				"let": &bson.M{
					"instance_id": "$instance",
				},
				"pipeline": []*bson.M{
					&bson.M{
						"$match": &bson.M{
							"$expr": &bson.M{
								"$eq": bson.A{"$_id", "$$instance_id"},
							},
						},
					},
					&bson.M{
						"$project": &bson.M{
							"name":              1,
							"action":            1,
							"state":             1,
							"timestamp":         1,
							"public_ips":        1,
							"public_ips6":       1,
							"private_ips":       1,
							"private_ips6":      1,
							"cloud_public_ips":  1,
							"cloud_public_ips6": 1,
						},
					},
				},
				"as": "instance_docs",
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

		var instanceInfo *DiskInstanceInfo
		if len(doc.InstanceDocs) > 0 {
			inst := doc.InstanceDocs[0]
			inst.Json(true)

			uptime := ""
			if !inst.Timestamp.IsZero() && inst.IsActive() {
				uptime = systemd.FormatUptime(inst.Timestamp)
			}

			instanceInfo = &DiskInstanceInfo{
				Id:              inst.Id,
				Name:            inst.Name,
				Status:          inst.Status,
				Timestamp:       inst.Timestamp,
				Uptime:          uptime,
				PublicIps:       inst.PublicIps,
				PublicIps6:      inst.PublicIps6,
				PrivateIps:      inst.PrivateIps,
				PrivateIps6:     inst.PrivateIps6,
				CloudPublicIps:  inst.CloudPublicIps,
				CloudPublicIps6: inst.CloudPublicIps6,
			}
		}

		dsk := &DiskAggregate{
			Disk:         doc.Disk,
			Backups:      backups,
			InstanceInfo: instanceInfo,
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
