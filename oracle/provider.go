package oracle

import (
	"crypto/rsa"
	"fmt"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/pritunl/pritunl-cloud/node"
)

type Provider struct {
	privateKey  *rsa.PrivateKey
	tenancy     string
	user        string
	fingerprint string
	region      string
	compartment string
	netClient   *core.VirtualNetworkClient
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
		return
	}

	p.netClient = &client
	netClient = p.netClient

	return
}

func NewProvider(nde *node.Node, mdata *Metadata) (
	prov *Provider, err error) {

	privateKey, fingerprint, err := loadPrivateKey(nde)
	if err != nil {
		return
	}

	prov = &Provider{
		privateKey:  privateKey,
		tenancy:     mdata.TenancyOcid,
		user:        mdata.UserOcid,
		fingerprint: fingerprint,
		region:      mdata.RegionName,
		compartment: mdata.CompartmentOcid,
	}

	return
}
