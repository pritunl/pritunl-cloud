package imds

import (
	"encoding/json"
	"io/ioutil"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/permission"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/service"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
)

type Config struct {
	ClientIps    []string             `json:"client_ips"`
	Instance     *types.Instance      `json:"instance"`
	Vpc          *types.Vpc           `json:"vpc"`
	Subnet       *types.Subnet        `json:"subnet"`
	Certificates []*types.Certificate `json:"certificates"`
	Secrets      []*types.Secret      `json:"secrets"`
	Services     []*types.Service     `json:"services"`
	Hash         uint32               `json:"hash"`
}

func (c *Config) ComputeHash() (err error) {
	confHash, err := utils.CrcHash(c)
	if err != nil {
		return
	}

	c.Hash = confHash
	return
}

func (c *Config) Write(virt *vm.VirtualMachine) (err error) {
	pth := paths.GetImdsConfPath(virt.Id)

	imdsDir := paths.GetImdsPath()
	err = utils.ExistsMkdir(imdsDir, 0755)
	if err != nil {
		return
	}

	data, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "imds: File marshal error"),
		}
		return
	}

	err = ioutil.WriteFile(pth, data, 0600)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "imds: File write error"),
		}
		return
	}

	err = permission.InitImdsConf(virt)
	if err != nil {
		return
	}

	return
}

func BuildConfig(inst *instance.Instance, virt *vm.VirtualMachine,
	vc *vpc.Vpc, subnet *vpc.Subnet, services []*service.Service,
	deployments map[primitive.ObjectID]*deployment.Deployment,
	secrs []*secret.Secret, certs []*certificate.Certificate) (
	conf *Config, err error) {

	conf = &Config{
		ClientIps:    inst.PrivateIps,
		Instance:     types.NewInstance(inst),
		Vpc:          types.NewVpc(vc),
		Subnet:       types.NewSubnet(subnet),
		Services:     types.NewServices(services, deployments),
		Secrets:      types.NewSecrets(secrs),
		Certificates: types.NewCertificates(certs),
	}

	return
}
