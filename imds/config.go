package imds

import (
	"encoding/json"
	"io/ioutil"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/permission"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
)

type Config struct {
	Instance     *instance.Instance         `json:"instance"`
	Certificates []*certificate.Certificate `json:"certificates"`
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
	certs []*certificate.Certificate) (conf *Config, err error) {

	conf = &Config{
		Instance:     inst,
		Certificates: certs,
	}

	return
}
