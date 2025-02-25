package deploy

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/imds"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vpc"
)

var (
	Hashes = map[primitive.ObjectID]uint32{}
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
		inst, virt, nil,
		vc, subnet,
		[]*pod.Pod{},
		map[primitive.ObjectID]*deployment.Deployment{},
		[]*secret.Secret{},
		[]*certificate.Certificate{},
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
	instPd := s.stat.Pod(deply.Pod)
	if instPd != nil {
		pods = append(pods, instPd)
	}
	for _, podId := range spc.Instance.Pods {
		servc := s.stat.SpecPod(podId)
		if servc == nil || servc.Organization != inst.Organization {
			continue
		}

		pods = append(pods, servc)
	}

	conf, err = imds.BuildConfig(
		inst, virt, spc,
		vc, subnet,
		pods,
		s.stat.DeploymentsDeployed(),
		secrs,
		certs,
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

	confs := map[primitive.ObjectID]*types.Config{}
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
