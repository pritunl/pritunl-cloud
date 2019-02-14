package oracle

import (
	"context"

	"github.com/dropbox/godropbox/errors"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type Subnet struct {
	Id                 string
	AvailabilityDomain string
	CidrBlock          string
	VcnId              string
	VirtualRouterIp    string
}

func GetSubnet(pv *Provider, subnetId string) (subnet *Subnet, err error) {
	client, err := pv.GetNetworkClient()
	if err != nil {
		return
	}

	subReq := core.GetSubnetRequest{
		SubnetId: &subnetId,
	}

	orcSubnet, err := client.GetSubnet(context.Background(), subReq)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "oracle: Failed to get subnet"),
		}
		return
	}

	subnet = &Subnet{}
	if orcSubnet.Id != nil {
		subnet.Id = *orcSubnet.Id
	}
	if orcSubnet.AvailabilityDomain != nil {
		subnet.AvailabilityDomain = *orcSubnet.AvailabilityDomain
	}
	if orcSubnet.CidrBlock != nil {
		subnet.CidrBlock = *orcSubnet.CidrBlock
	}
	if orcSubnet.VcnId != nil {
		subnet.VcnId = *orcSubnet.VcnId
	}
	if orcSubnet.VirtualRouterIp != nil {
		subnet.VirtualRouterIp = *orcSubnet.VirtualRouterIp
	}

	return
}
