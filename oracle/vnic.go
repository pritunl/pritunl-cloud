package oracle

import (
	"context"

	"github.com/dropbox/godropbox/errors"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Vnic struct {
	Id                  string
	SubnetId            string
	IsPrimary           bool
	MacAddress          string
	PrivateIp           string
	PrivateIpId         string
	PublicIp            string
	SkipSourceDestCheck bool
}

func (v *Vnic) SetSkipSourceDestCheck(pv *Provider, val bool) (err error) {
	client, err := pv.GetNetworkClient()
	if err != nil {
		return
	}

	req := core.UpdateVnicRequest{
		VnicId: &v.Id,
		UpdateVnicDetails: core.UpdateVnicDetails{
			SkipSourceDestCheck: &val,
		},
	}

	_, err = client.UpdateVnic(context.Background(), req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "oracle: Failed to update vnic"),
		}
		return
	}

	return
}

func GetVnic(pv *Provider, vnicId string) (vnic *Vnic, err error) {
	client, err := pv.GetNetworkClient()
	if err != nil {
		return
	}

	req := core.GetVnicRequest{
		VnicId: &vnicId,
	}

	orcVnic, err := client.GetVnic(context.Background(), req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "oracle: Failed to get vnic"),
		}
		return
	}

	vnic = &Vnic{}
	if orcVnic.Id != nil {
		vnic.Id = *orcVnic.Id
	}
	if orcVnic.SubnetId != nil {
		vnic.SubnetId = *orcVnic.SubnetId
	}
	if orcVnic.IsPrimary != nil {
		vnic.IsPrimary = *orcVnic.IsPrimary
	}
	if orcVnic.MacAddress != nil {
		vnic.MacAddress = *orcVnic.MacAddress
	}
	if orcVnic.PrivateIp != nil {
		vnic.PrivateIp = *orcVnic.PrivateIp
	}
	if orcVnic.PublicIp != nil {
		vnic.PublicIp = *orcVnic.PublicIp
	}
	if orcVnic.SkipSourceDestCheck != nil {
		vnic.SkipSourceDestCheck = *orcVnic.SkipSourceDestCheck
	}

	limit := 10
	ipReq := core.ListPrivateIpsRequest{
		VnicId: &vnic.Id,
		Limit:  &limit,
	}

	orcIps, err := client.ListPrivateIps(context.Background(), ipReq)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "oracle: Failed to get vnic ips"),
		}
		return
	}

	if orcIps.Items != nil {
		for _, orcIp := range orcIps.Items {
			if orcIp.IsPrimary != nil && *orcIp.IsPrimary && orcIp.Id != nil {
				vnic.PrivateIpId = *orcIp.Id
				break
			}
		}
	}

	return
}

func CreateVnic(pv *Provider, name string, subnetId string) (err error) {
	client, err := pv.GetComputeClient()
	if err != nil {
		return
	}

	req := core.AttachVnicRequest{
		AttachVnicDetails: core.AttachVnicDetails{
			InstanceId:  utils.PointerString(pv.Metadata.InstanceOcid),
			DisplayName: utils.PointerString(name),
			CreateVnicDetails: &core.CreateVnicDetails{
				AssignPublicIp: utils.PointerBool(true),
				DisplayName:    utils.PointerString(name),
				SubnetId:       utils.PointerString(subnetId),
			},
		},
	}

	_, err = client.AttachVnic(context.Background(), req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "oracle: Failed to create vnic"),
		}
		return
	}

	return
}
