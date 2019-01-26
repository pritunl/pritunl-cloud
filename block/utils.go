package block

import (
	"github.com/pritunl/pritunl-cloud/database"
	"gopkg.in/mgo.v2/bson"
)

func Get(db *database.Database, blockId bson.ObjectId) (
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

	cursor := coll.Find(bson.M{}).Iter()

	nde := &Block{}
	for cursor.Next(nde) {
		blocks = append(blocks, nde)
		nde = &Block{}
	}

	err = cursor.Close()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Remove(db *database.Database, blockId bson.ObjectId) (err error) {
	coll := db.Blocks()

	_, err = coll.RemoveAll(&bson.M{
		"storage": blockId,
	})
	if err != nil {
		return
	}

	coll = db.Blocks()

	err = coll.Remove(&bson.M{
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
