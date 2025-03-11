package oracle

import (
	"context"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
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
	PublicIp6           string
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
		VnicId: utils.PointerString(vnicId),
	}

	orcVnic, err := client.GetVnic(context.Background(), req)
	if err != nil {
		if orcVnic.RawResponse != nil &&
			orcVnic.RawResponse.StatusCode == 404 {

			err = &errortypes.NotFoundError{
				errors.Wrap(err, "oracle: Failed to find vnic"),
			}
			return
		}

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
	if len(orcVnic.Ipv6Addresses) > 0 {
		vnic.PublicIp6 = orcVnic.Ipv6Addresses[0]
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
			if orcIp.IsPrimary != nil && *orcIp.IsPrimary &&
				orcIp.Id != nil {

				vnic.PrivateIpId = *orcIp.Id
				break
			}
		}
	}

	return
}

func getVnicAttachment(pv *Provider, attachmentId string) (
	vnicId string, err error) {

	client, err := pv.GetComputeClient()
	if err != nil {
		return
	}

	req := core.GetVnicAttachmentRequest{
		VnicAttachmentId: utils.PointerString(attachmentId),
	}

	resp, err := client.GetVnicAttachment(context.Background(), req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "oracle: Failed to create vnic"),
		}
		return
	}

	if resp.VnicId != nil {
		vnicId = *resp.VnicId
	}

	return
}

func CreateVnic(pv *Provider, name, subnetId string,
	publicIp, publicIp6 bool) (vnicId, vnicAttachId string, err error) {

	client, err := pv.GetComputeClient()
	if err != nil {
		return
	}

	req := core.AttachVnicRequest{
		AttachVnicDetails: core.AttachVnicDetails{
			InstanceId:  utils.PointerString(pv.Metadata.InstanceOcid),
			DisplayName: utils.PointerString(name),
			CreateVnicDetails: &core.CreateVnicDetails{
				AssignPublicIp: utils.PointerBool(publicIp),
				AssignIpv6Ip:   utils.PointerBool(publicIp6),
				DisplayName:    utils.PointerString(name),
				SubnetId:       utils.PointerString(subnetId),
			},
		},
	}

	var resp core.AttachVnicResponse

	retryCount := settings.System.OracleApiRetryCount
	retryRate := time.Duration(
		settings.System.OracleApiRetryRate) * time.Second

	for i := 0; i < retryCount; i++ {
		resp, err = client.AttachVnic(context.Background(), req)
		if err != nil {
			if i != retryCount-1 && resp.RawResponse != nil &&
				resp.RawResponse.StatusCode == 409 {

				time.Sleep(retryRate)

				continue
			}

			err = &errortypes.RequestError{
				errors.Wrap(err, "oracle: Failed to create vnic"),
			}
			return
		}

		break
	}

	if resp.Id == nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "oracle: Nil vnic attachment id"),
		}
		return
	}

	vnicAttachId = *resp.Id

	for i := 0; i < 60; i++ {
		vnicId, err = getVnicAttachment(pv, vnicAttachId)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			if i == 59 {
				return
			}
			err = nil
			continue
		}

		if vnicId == "" {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		break
	}

	if vnicId == "" {
		err = &errortypes.ParseError{
			errors.Wrap(err, "oracle: Nil vnic id"),
		}
		return
	}

	return
}

func RemoveVnic(pv *Provider, vnicAttachId string) (err error) {
	client, err := pv.GetComputeClient()
	if err != nil {
		return
	}

	req := core.DetachVnicRequest{
		VnicAttachmentId: utils.PointerString(vnicAttachId),
	}

	retryCount := settings.System.OracleApiRetryCount
	retryRate := time.Duration(
		settings.System.OracleApiRetryRate) * time.Second

	for i := 0; i < retryCount; i++ {
		resp, e := client.DetachVnic(context.Background(), req)
		if e != nil {
			if i != retryCount-1 && resp.RawResponse != nil &&
				resp.RawResponse.StatusCode == 409 {

				time.Sleep(retryRate)

				continue
			}

			err = &errortypes.RequestError{
				errors.Wrap(e, "oracle: Failed to remove vnic"),
			}
			return
		}

		break
	}

	return
}
