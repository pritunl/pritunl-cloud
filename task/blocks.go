package task

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

var blocksCheck = &Task{
	Name:    "blocks_check",
	Hours:   []int{7},
	Minutes: []int{30},
	Handler: blocksCheckHandler,
}

func blocksCheckHandler(db *database.Database) (err error) {
	coll := db.Blocks()
	ipColl := db.BlocksIp()
	instColl := db.Instances()

	ipBlocksInf, err := ipColl.Distinct(db, "block", &bson.M{
		"type": &bson.M{
			"$ne": block.Host,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	blocks := set.NewSet()
	ipBlocks := set.NewSet()

	for _, ipBlockInf := range ipBlocksInf {
		if ipBlock, ok := ipBlockInf.(primitive.ObjectID); ok {
			ipBlocks.Add(ipBlock)
		}
	}

	cursor, err := coll.Find(db, &bson.M{})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		blck := &block.Block{}
		err = cursor.Decode(blck)
		if err != nil {
			err = database.ParseError(err)
			return
		}
		blocks.Add(blck.Id)

		err = blck.ValidateAddresses(db, nil)
		if err != nil {
			return
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	ipBlocks.Subtract(blocks)
	for blckIdInf := range ipBlocks.Iter() {
		blckId := blckIdInf.(primitive.ObjectID)

		cursor2, e := ipColl.Find(db, &bson.M{
			"block": blckId,
		})
		if e != nil {
			err = database.ParseError(e)
			return
		}
		defer cursor2.Close(db)

		for cursor2.Next(db) {
			blckIp := &block.BlockIp{}
			err = cursor2.Decode(blckIp)
			if err != nil {
				err = database.ParseError(err)
				return
			}

			logrus.WithFields(logrus.Fields{
				"ip_address":  utils.Int2IpAddress(blckIp.Ip).String(),
				"block_id":    blckIp.Id.Hex(),
				"instance_id": blckIp.Instance.Hex(),
			}).Warn("task: Removing lost block IP")

			_, _ = instColl.UpdateOne(db, &bson.M{
				"_id": blckIp.Instance,
			}, &bson.M{
				"$set": &bson.M{
					"restart_block_ip": true,
				},
			})

			_, err = ipColl.DeleteOne(db, &bson.M{
				"_id": blckIp.Id,
			})
			if err != nil {
				err = database.ParseError(err)
				if _, ok := err.(*database.NotFoundError); ok {
					err = nil
				} else {
					return
				}
			}
		}

		err = cursor2.Err()
		if err != nil {
			err = database.ParseError(err)
			return
		}
	}

	return
}

func init() {
	register(blocksCheck)
}
