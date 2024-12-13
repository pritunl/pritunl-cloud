package imds

import (
	"sync"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/service"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
)

var (
	curConfs     = map[primitive.ObjectID]*types.Config{}
	curConfsLock = sync.Mutex{}
)

func BuildConfig(inst *instance.Instance, virt *vm.VirtualMachine,
	vc *vpc.Vpc, subnet *vpc.Subnet, services []*service.Service,
	deployments map[primitive.ObjectID]*deployment.Deployment,
	secrs []*secret.Secret, certs []*certificate.Certificate) (
	conf *types.Config, err error) {

	conf = &types.Config{
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

func SetConfigs(cnfs map[primitive.ObjectID]*types.Config) {
	curConfsLock.Lock()
	curConfs = cnfs
	curConfsLock.Unlock()
}

func GetConfigs() (
	cnfs map[primitive.ObjectID]*types.Config) {

	curConfsLock.Lock()
	cnfs = curConfs
	curConfsLock.Unlock()
	return
}
