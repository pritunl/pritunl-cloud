package cloud

import (
	"github.com/pritunl/pritunl-cloud/oracle"
)

type Subnet struct {
	Id      string `bson:"id" json:"id"`
	VpcId   string `bson:"vpc_id" json:"vpc_id"`
	Name    string `bson:"name" json:"name"`
	Network string `bson:"network" json:"network"`
}

type Vpc struct {
	Id      string    `bson:"id" json:"id"`
	Name    string    `bson:"name" json:"name"`
	Network string    `bson:"network" json:"network"`
	Subnets []*Subnet `bson:"subnets" json:"subnets"`
}

func GetOracleVpcs(authPv oracle.AuthProvider) (vpcs []*Vpc, err error) {
	pv, err := oracle.NewProvider(authPv)
	if err != nil {
		return
	}

	vcns, err := oracle.GetVcns(pv)
	if err != nil {
		return
	}

	vpcs = []*Vpc{}
	for _, ociVcn := range vcns {
		vpc := &Vpc{
			Id:      ociVcn.Id,
			Name:    ociVcn.Name,
			Network: ociVcn.Network,
			Subnets: []*Subnet{},
		}

		for _, ociSubnet := range ociVcn.Subnets {
			subnet := &Subnet{
				Id:      ociSubnet.Id,
				VpcId:   ociSubnet.VcnId,
				Name:    ociSubnet.Name,
				Network: ociSubnet.Network,
			}

			vpc.Subnets = append(vpc.Subnets, subnet)
		}

		vpcs = append(vpcs, vpc)
	}

	return
}
