package block

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
)

func Get(db *database.Database, blockId primitive.ObjectID) (
	block *Block, err error) {

	coll := db.Blocks()
	block = &Block{}

	err = coll.FindOneId(blockId, block)
	if err != nil {
		return
	}

	return
}

func GetAll(db *database.Database) (blocks []*Block, err error) {
	coll := db.Blocks()
	blocks = []*Block{}

	opts := &options.FindOptions{
		Sort: &bson.D{
			{"name", 1},
		},
	}

	cursor, err := coll.Find(db, bson.M{}, opts)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		blck := &Block{}
		err = cursor.Decode(blck)
		if err != nil {
			err = database.ParseError(err)
			return
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

func GetInstanceIp(db *database.Database, instId primitive.ObjectID,
	typ string) (blck *Block, blckIp *BlockIp, err error) {

	coll := db.BlocksIp()
	blckIp = &BlockIp{}

	err = coll.FindOne(db, &bson.M{
		"instance": instId,
		"type":     typ,
	}).Decode(blckIp)
	if err != nil {
		err = database.ParseError(err)
		blckIp = nil
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		}
		return
	}

	blck, err = Get(db, blckIp.Block)
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			RemoveIp(db, blckIp.Id)
		}
		blckIp = nil
		return
	}

	return
}

func Remove(db *database.Database, blockId primitive.ObjectID) (err error) {
	coll := db.Blocks()
	ipColl := db.BlocksIp()
	instColl := db.Instances()

	cursor, err := ipColl.Find(db, &bson.M{
		"block": blockId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		blckIp := &BlockIp{}
		err = cursor.Decode(blckIp)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		_, _ = instColl.UpdateOne(db, &bson.M{
			"_id": blckIp.Instance,
		}, &bson.M{
			"$set": &bson.M{
				"restart_block_ip": true,
			},
		})
	}

	_, err = ipColl.DeleteMany(db, &bson.M{
		"block": blockId,
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

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": blockId,
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

func RemoveIp(db *database.Database, blockIpId primitive.ObjectID) (
	err error) {

	coll := db.BlocksIp()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": blockIpId,
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

func RemoveInstanceIps(db *database.Database, instId primitive.ObjectID) (
	err error) {

	coll := db.BlocksIp()

	_, err = coll.DeleteMany(db, &bson.M{
		"instance": instId,
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
