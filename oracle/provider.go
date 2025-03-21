package oracle

import (
	"crypto/rsa"
	"fmt"

	"github.com/dropbox/godropbox/errors"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/sirupsen/logrus"
)

type Provider struct {
	Metadata      *Metadata
	privateKey    *rsa.PrivateKey
	tenancy       string
	user          string
	fingerprint   string
	region        string
	compartment   string
	netClient     *core.VirtualNetworkClient
	computeClient *core.ComputeClient
}

func (p *Provider) LogInfo() {
	logrus.WithFields(logrus.Fields{
		"region":        p.Metadata.RegionName,
		"tenancy":       p.Metadata.TenancyOcid,
		"compartment":   p.Metadata.CompartmentOcid,
		"instance":      p.Metadata.InstanceOcid,
		"instance_vnic": p.Metadata.VnicOcid,
		"user":          p.Metadata.UserOcid,
		"fingerprint":   p.fingerprint,
	}).Info("oracle: Oracle provider data")
}

func (p *Provider) AuthType() (common.AuthConfig, error) {
	return common.AuthConfig{
		AuthType:         common.UserPrincipal,
		IsFromConfigFile: false,
		OboToken:         nil,
	}, nil
}

func (p *Provider) PrivateRSAKey() (*rsa.PrivateKey, error) {
	return p.privateKey, nil
}

func (p *Provider) KeyID() (string, error) {
	return fmt.Sprintf("%s/%s/%s", p.tenancy, p.user, p.fingerprint), nil
}

func (p *Provider) TenancyOCID() (string, error) {
	return p.tenancy, nil
}

func (p *Provider) UserOCID() (string, error) {
	return p.user, nil
}

func (p *Provider) KeyFingerprint() (string, error) {
	return p.fingerprint, nil
}

func (p *Provider) Region() (string, error) {
	return p.region, nil
}

func (p *Provider) CompartmentOCID() (string, error) {
	return p.compartment, nil
}

func (p *Provider) GetNetworkClient() (
	netClient *core.VirtualNetworkClient, err error) {

	if p.netClient != nil {
		netClient = p.netClient
		return
	}

	client, err := core.NewVirtualNetworkClientWithConfigurationProvider(p)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "oracle: Failed to create oracle client"),
		}
		return
	}

	p.netClient = &client
	netClient = p.netClient

	return
}

func (p *Provider) GetComputeClient() (
	computeClient *core.ComputeClient, err error) {

	if p.computeClient != nil {
		computeClient = p.computeClient
		return
	}

	client, err := core.NewComputeClientWithConfigurationProvider(p)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "oracle: Failed to create oracle client"),
		}
		return
	}

	p.computeClient = &client
	computeClient = p.computeClient

	return
}

func NewProvider(authPv AuthProvider) (prov *Provider, err error) {
	mdata, err := GetMetadata(authPv)
	if err != nil {
		return
	}

	privateKey, fingerprint, err := loadPrivateKey(mdata)
	if err != nil {
		return
	}

	prov = &Provider{
		Metadata:    mdata,
		privateKey:  privateKey,
		tenancy:     mdata.TenancyOcid,
		user:        mdata.UserOcid,
		fingerprint: fingerprint,
		region:      mdata.RegionName,
		compartment: mdata.CompartmentOcid,
	}

	return
}
