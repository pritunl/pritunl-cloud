package block

import (
	"net"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
)

func GetNodeBlock(ndeId primitive.ObjectID) (blck *Block, err error) {
	hostNetwork := settings.Hypervisor.HostNetwork

	hostAddr, hostNet, err := net.ParseCIDR(hostNetwork)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "block: Failed to parse host network"),
		}
		return
	}

	utils.IncIpAddress(hostAddr)

	blck = &Block{
		Id:   ndeId,
		Name: "host-block",
		Type: Host,
		Subnets: []string{
			hostNetwork,
		},
		Netmask: net.IP(hostNet.Mask).String(),
		Gateway: hostAddr.String(),
	}

	return
}

func GetNodePortBlock(ndeId primitive.ObjectID) (blck *Block, err error) {
	portNetwork := settings.Hypervisor.NodePortNetwork

	portAddr, portNet, err := net.ParseCIDR(portNetwork)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "block: Failed to parse node port network"),
		}
		return
	}

	utils.IncIpAddress(portAddr)

	blck = &Block{
		Id:   ndeId,
		Name: "node-port-block",
		Type: NodePort,
		Subnets: []string{
			portNetwork,
		},
		Netmask: net.IP(portNet.Mask).String(),
		Gateway: portAddr.String(),
	}

	return
}

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

func GetInstanceHostIp(db *database.Database,
	instId primitive.ObjectID) (blckIp *BlockIp, err error) {

	coll := db.BlocksIp()
	blckIp = &BlockIp{}

	err = coll.FindOne(db, &bson.M{
		"instance": instId,
		"type":     Host,
	}).Decode(blckIp)
	if err != nil {
		err = database.ParseError(err)
		blckIp = nil
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		}
		return
	}

	return
}

func GetInstanceNodePortIp(db *database.Database,
	instId primitive.ObjectID) (blckIp *BlockIp, err error) {

	coll := db.BlocksIp()
	blckIp = &BlockIp{}

	err = coll.FindOne(db, &bson.M{
		"instance": instId,
		"type":     NodePort,
	}).Decode(blckIp)
	if err != nil {
		err = database.ParseError(err)
		blckIp = nil
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		}
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
	nodeColl := db.Nodes()

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

	_, err = nodeColl.UpdateMany(db, &bson.M{
		"host_block": blockId,
	}, &bson.M{"$set": &bson.M{
		"host_block": primitive.NilObjectID,
	}})
	if err != nil {
		err = database.ParseError(err)
		return
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

func RemoveInstanceIpsType(db *database.Database,
	instId primitive.ObjectID, typ string) (err error) {

	coll := db.BlocksIp()

	_, err = coll.DeleteMany(db, &bson.M{
		"instance": instId,
		"type":     typ,
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
