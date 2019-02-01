package block

import (
	"net"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Block struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Addresses []string           `bson:"addresses" json:"addresses"`
	Excludes  []string           `bson:"excludes" json:"excludes"`
	Netmask   string             `bson:"netmask" json:"netmask"`
	Gateway   string             `bson:"gateway" json:"gateway"`
}

func (b *Block) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	return
}

func (b *Block) GetGateway() net.IP {
	return net.ParseIP(b.Gateway)
}

func (b *Block) GetMask() net.IPMask {
	return utils.ParseIpMask(b.Netmask)
}

func (b *Block) RemoveIp(db *database.Database,
	instId primitive.ObjectID) (err error) {

	coll := db.BlocksIp()
	_, err = coll.DeleteMany(db, &bson.M{
		"instance": instId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (b *Block) Commit(db *database.Database) (err error) {
	coll := db.Blocks()

	err = coll.Commit(b.Id, b)
	if err != nil {
		return
	}

	return
}

func (b *Block) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Blocks()

	err = coll.CommitFields(b.Id, b, fields)
	if err != nil {
		return
	}

	return
}

func (b *Block) Insert(db *database.Database) (err error) {
	coll := db.Blocks()

	if !b.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("block: Block already exists"),
		}
		return
	}

	_, err = coll.InsertOne(db, b)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
