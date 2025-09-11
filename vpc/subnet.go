package vpc

import (
	"net"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Subnet struct {
	Id      bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name    string        `bson:"name" json:"name"`
	Network string        `bson:"network" json:"network"`
}

func (s *Subnet) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	s.Name = utils.FilterName(s.Name)

	return
}

func (s *Subnet) GetNetwork() (network *net.IPNet, err error) {
	_, network, err = net.ParseCIDR(s.Network)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "vpc: Failed to parse subnet"),
		}
		return
	}
	return
}

func (s *Subnet) GetIndexRange() (start, stop int64, err error) {
	network, err := s.GetNetwork()
	if err != nil {
		return
	}

	start, err = utils.GetFirstIpIndex(network)
	if err != nil {
		return
	}

	stop, err = utils.GetLastIpIndex(network)
	if err != nil {
		return
	}

	return
}
