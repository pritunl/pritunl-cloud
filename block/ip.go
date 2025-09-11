package block

import (
	"net"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/utils"
)

type BlockIp struct {
	Id       bson.ObjectID `bson:"_id,omitempty"`
	Block    bson.ObjectID `bson:"block"`
	Ip       int64         `bson:"ip"`
	Instance bson.ObjectID `bson:"instance"`
	Type     string        `bson:"type"`
}

func (b *BlockIp) GetIp() net.IP {
	return utils.Int2IpAddress(b.Ip)
}
