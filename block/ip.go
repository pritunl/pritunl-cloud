package block

import (
	"net"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/utils"
)

type BlockIp struct {
	Id       primitive.ObjectID `bson:"_id,omitempty"`
	Block    primitive.ObjectID `bson:"block"`
	Ip       int64              `bson:"ip"`
	Instance primitive.ObjectID `bson:"instance"`
}

func (b *BlockIp) GetIp() net.IP {
	return utils.Int2IpAddress(b.Ip)
}
