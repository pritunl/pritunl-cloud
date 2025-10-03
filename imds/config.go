package imds

import (
	"sync"

	"github.com/pritunl/mongo-go-driver/v2/bson"
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
	curConfs     = map[bson.ObjectID]*types.Config{}
	curConfsLock = sync.Mutex{}
)

func BuildConfig(inst *instance.Instance, virt *vm.VirtualMachine,
	unt *unit.Unit, spc *spec.Spec, vc *vpc.Vpc, subnet *vpc.Subnet,
	pods []*pod.Pod, podUnitsMap map[bson.ObjectID][]*unit.Unit,
	deployments map[bson.ObjectID]*deployment.Deployment,
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
		conf.Journals = types.NewJournals(spc)
	}

	return
}

func SetConfigs(cnfs map[bson.ObjectID]*types.Config) {
	curConfsLock.Lock()
	curConfs = cnfs
	curConfsLock.Unlock()
}

func GetConfigs() (
	cnfs map[bson.ObjectID]*types.Config) {

	curConfsLock.Lock()
	cnfs = curConfs
	curConfsLock.Unlock()
	return
}
