package oracle

import (
	"context"

	"github.com/dropbox/godropbox/errors"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Vcn struct {
	Id      string
	Name    string
	Network string
	Subnets []*Subnet
}

type Subnet struct {
	Id      string
	VcnId   string
	Name    string
	Network string
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
	if orcSubnet.VcnId != nil {
		subnet.VcnId = *orcSubnet.VcnId
	}
	if orcSubnet.DisplayName != nil {
		subnet.Name = *orcSubnet.DisplayName
	}
	if orcSubnet.CidrBlock != nil {
		subnet.Network = *orcSubnet.CidrBlock
	}

	return
}

func GetVcns(pv *Provider) (vcns []*Vcn, err error) {
	client, err := pv.GetNetworkClient()
	if err != nil {
		return
	}

	compartmentId, err := pv.CompartmentOCID()
	if err != nil {
		return
	}

	req := core.ListVcnsRequest{
		CompartmentId: &compartmentId,
		Limit:         utils.PointerInt(100),
	}

	orcVcns, err := client.ListVcns(context.Background(), req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "oracle: Failed to get VCNs"),
		}
		return
	}

	vcns = []*Vcn{}

	for _, orcVcn := range orcVcns.Items {
		vcn := &Vcn{}

		if orcVcn.Id != nil {
			vcn.Id = *orcVcn.Id
		}
		if orcVcn.DisplayName != nil {
			vcn.Name = *orcVcn.DisplayName
		}
		if orcVcn.CidrBlock != nil {
			vcn.Network = *orcVcn.CidrBlock
		}

		subnets, e := GetSubnets(pv, vcn.Id)
		if e != nil {
			err = e
			return
		}

		vcn.Subnets = subnets

		vcns = append(vcns, vcn)
	}

	return
}

func GetSubnets(pv *Provider, vcnId string) (subnets []*Subnet, err error) {
	client, err := pv.GetNetworkClient()
	if err != nil {
		return
	}

	compartmentId, err := pv.CompartmentOCID()
	if err != nil {
		return
	}

	req := core.ListSubnetsRequest{
		CompartmentId: &compartmentId,
		VcnId:         utils.PointerString(vcnId),
		Limit:         utils.PointerInt(256),
	}

	orcSubnets, err := client.ListSubnets(context.Background(), req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "oracle: Failed to get subnets"),
		}
		return
	}

	subnets = []*Subnet{}
	for _, orcSubnet := range orcSubnets.Items {
		subnet := &Subnet{}

		if orcSubnet.Id != nil {
			subnet.Id = *orcSubnet.Id
		}
		if orcSubnet.VcnId != nil {
			subnet.VcnId = *orcSubnet.VcnId
		}
		if orcSubnet.DisplayName != nil {
			subnet.Name = *orcSubnet.DisplayName
		}
		if orcSubnet.CidrBlock != nil {
			subnet.Network = *orcSubnet.CidrBlock
		}

		subnets = append(subnets, subnet)
	}

	return
}
