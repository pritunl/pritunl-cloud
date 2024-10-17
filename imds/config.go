package imds

import (
	"encoding/json"
	"io/ioutil"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/permission"
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

	err = permission.InitImds(virt)
	if err != nil {
		return
	}

	return
}

func BuildConfig(inst *instance.Instance, virt *vm.VirtualMachine,
	vc *vpc.Vpc, subnet *vpc.Subnet,
	certs []*certificate.Certificate) (conf *Config, err error) {

	conf = &Config{
		ClientIps:    inst.PrivateIps,
		Instance:     types.NewInstance(inst),
		Vpc:          types.NewVpc(vc),
		Subnet:       types.NewSubnet(subnet),
		Certificates: types.NewCertificates(certs),
	}

	return
}
