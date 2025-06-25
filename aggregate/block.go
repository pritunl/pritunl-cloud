package aggregate

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
)

type BlockPipe struct {
	block.Block `bson:",inline"`
	IpCount     int64 `bson:"ip_count"`
}

type BlockAggregate struct {
	block.Block
	Available int64 `json:"available"`
	Capacity  int64 `json:"capacity"`
}

func GetBlockPaged(db *database.Database, query *bson.M, page,
	pageCount int64) (blocks []*BlockAggregate, count int64, err error) {

	coll := db.Blocks()
	blocks = []*BlockAggregate{}

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
				"from":         "blocks_ip",
				"localField":   "_id",
				"foreignField": "block",
				"as":           "ips",
			},
		},
		&bson.M{
			"$addFields": &bson.M{
				"ip_count": &bson.M{
					"$size": "$ips",
				},
			},
		},
		&bson.M{
			"$project": &bson.M{
				"ips": 0,
			},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		doc := &BlockPipe{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		total, e := doc.GetIpCount()
		if e != nil {
			err = e
			return
		}

		blck := &BlockAggregate{
			Block:     doc.Block,
			Available: total - doc.IpCount,
			Capacity:  total,
		}

		blocks = append(blocks, blck)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
