package cloud

import (
	"time"

	"github.com/pritunl/pritunl-cloud/oracle"
)

var (
	lastOracleSync time.Time
	oracleVpcs     []*Vpc
)

func GetOracleVpcs(authPv oracle.AuthProvider) (vpcs []*Vpc, err error) {
	if time.Since(lastOracleSync) < 30*time.Second {
		vpcs = oracleVpcs
		return
	}

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

	lastOracleSync = time.Now()
	oracleVpcs = vpcs

	return
}
