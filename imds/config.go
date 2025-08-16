package imds

import (
	"sync"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/unit"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
)

var (
	curConfs     = map[primitive.ObjectID]*types.Config{}
	curConfsLock = sync.Mutex{}
)

func BuildConfig(inst *instance.Instance, virt *vm.VirtualMachine,
	spc *spec.Spec, vc *vpc.Vpc, subnet *vpc.Subnet,
	pods []*pod.Pod, podUnitsMap map[primitive.ObjectID][]*unit.Unit,
	deployments map[primitive.ObjectID]*deployment.Deployment,
	secrs []*secret.Secret, certs []*certificate.Certificate) (
	conf *types.Config, err error) {

	conf = &types.Config{
		ImdsHostSecret: virt.ImdsHostSecret,
		ClientIps:      inst.PrivateIps,
		Node:           types.NewNode(node.Self),
		Instance:       types.NewInstance(inst),
		Vpc:            types.NewVpc(vc),
		Subnet:         types.NewSubnet(subnet),
		Pods:           types.NewPods(pods, podUnitsMap, deployments),
		Secrets:        types.NewSecrets(secrs),
		Certificates:   types.NewCertificates(certs),
	}

	if spc != nil {
		conf.Spec = spc.Id
		conf.SpecData = spc.Data
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
