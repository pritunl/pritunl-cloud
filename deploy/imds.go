package deploy

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/imds"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/unit"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
)

var (
	Hashes = map[bson.ObjectID]uint32{}
)

type Imds struct {
	stat *state.State
}

func (s *Imds) buildInstance(db *database.Database,
	inst *instance.Instance, virt *vm.VirtualMachine) (
	conf *types.Config, err error) {

	vc := s.stat.Vpc(inst.Vpc)

	var subnet *vpc.Subnet
	if vc != nil {
		subnet = vc.GetSubnet(inst.Subnet)
	}

	conf, err = imds.BuildConfig(
		inst, virt, nil, nil,
		vc, subnet,
		[]*pod.Pod{},
		map[bson.ObjectID][]*unit.Unit{},
		map[bson.ObjectID]*deployment.Deployment{},
		[]*secret.Secret{},
		[]*certificate.Certificate{},
		[]*types.Domain{},
	)
	if err != nil {
		return
	}

	err = conf.ComputeHash()
	if err != nil {
		return
	}

	return
}

func (s *Imds) buildDeployInstance(db *database.Database,
	inst *instance.Instance, virt *vm.VirtualMachine) (
	conf *types.Config, err error) {

	vc := s.stat.Vpc(inst.Vpc)

	var subnet *vpc.Subnet
	if vc != nil {
		subnet = vc.GetSubnet(inst.Subnet)
	}

	deply := s.stat.Deployment(virt.Deployment)
	if deply == nil {
		println("**************************************************1")
		println(inst.Id.Hex())
		println("**************************************************1")
		return
	}

	spc := s.stat.Spec(deply.Spec)
	if spc == nil {
		println("**************************************************2")
		println(inst.Id.Hex())
		println("**************************************************2")
		return
	}

	certs := []*certificate.Certificate{}
	for _, certId := range spc.Instance.Certificates {
		cert := s.stat.SpecCert(certId)
		if cert == nil || cert.Organization != inst.Organization {
			continue
		}

		certs = append(certs, cert)
	}

	secrs := []*secret.Secret{}
	for _, secrId := range spc.Instance.Secrets {
		secr := s.stat.SpecSecret(secrId)
		if secr == nil || secr.Organization != inst.Organization {
			continue
		}

		secrs = append(secrs, secr)
	}

	pods := []*pod.Pod{}
	podUnitsMap := map[bson.ObjectID][]*unit.Unit{}

	instPd := s.stat.Pod(deply.Pod)
	if instPd != nil {
		pods = append(pods, instPd)
	}

	instUnt := s.stat.Unit(deply.Unit)
	if instUnt != nil {
		podUnitsMap[deply.Pod] = append(podUnitsMap[deply.Pod], instUnt)
	}

	for _, podId := range spc.Instance.Pods {
		pd := s.stat.SpecPod(podId)
		if pd == nil || pd.Organization != inst.Organization {
			continue
		}

		pods = append(pods, pd)

		podUnits := s.stat.SpecPodUnits(podId)
		if podUnits != nil {
			podUnitsMap[podId] = podUnits
		}
	}

	conf, err = imds.BuildConfig(
		inst, virt, instUnt, spc,
		vc, subnet,
		pods,
		podUnitsMap,
		s.stat.DeploymentsDeployed(),
		secrs,
		certs,
		s.stat.GetDomains(inst.Organization),
	)
	if err != nil {
		return
	}

	err = conf.ComputeHash()
	if err != nil {
		return
	}

	return
}

func (s *Imds) Deploy(db *database.Database) (err error) {
	instances := s.stat.Instances()

	confs := map[bson.ObjectID]*types.Config{}
	for _, inst := range instances {
		if !inst.IsActive() {
			continue
		}

		virt := s.stat.GetVirt(inst.Id)
		if virt == nil {
			continue
		}

		if virt.ImdsVersion < 1 {
			continue
		}

		var conf *types.Config
		if inst.Deployment.IsZero() {
			conf, err = s.buildInstance(db, inst, virt)
			if err != nil {
				return
			}
		} else {
			conf, err = s.buildDeployInstance(db, inst, virt)
			if err != nil {
				return
			}
		}

		if conf != nil {
			confs[inst.Id] = conf
		}
	}

	imds.SetConfigs(confs)

	return
}

func NewImds(stat *state.State) *Imds {
	return &Imds{
		stat: stat,
	}
}
