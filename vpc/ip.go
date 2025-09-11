package vpc

import (
	"net"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/utils"
)

type VpcIp struct {
	Id       bson.ObjectID `bson:"_id,omitempty"`
	Vpc      bson.ObjectID `bson:"vpc"`
	Subnet   bson.ObjectID `bson:"subnet"`
	Ip       int64         `bson:"ip"`
	Instance bson.ObjectID `bson:"instance"`
}

func (i *VpcIp) GetIp() net.IP {
	return utils.Int2IpAddress(i.Ip * 2)
}

func (i *VpcIp) GetIps() (net.IP, net.IP) {
	return utils.IpIndex2Ip(i.Ip)
}
