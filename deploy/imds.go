package deploy

import (
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/imds"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/vpc"
)

type Imds struct {
	stat *state.State
}

func (s *Imds) buildInstance(db *database.Database,
	inst *instance.Instance) (err error) {

	virt := s.stat.GetVirt(inst.Id)
	if virt == nil {
		return
	}

	vc := s.stat.Vpc(inst.Vpc)

	var subnet *vpc.Subnet
	if vc != nil {
		subnet = vc.GetSubnet(inst.Subnet)
	}

	conf, err := imds.BuildConfig(
		inst, virt,
		vc, subnet,
		[]*certificate.Certificate{},
		[]*secret.Secret{},
	)
	if err != nil {
		return
	}

	// TODO Only write on change
	err = conf.Write(virt)
	if err != nil {
		return
	}

	return
}

func (s *Imds) buildDeployInstance(db *database.Database,
	inst *instance.Instance) (err error) {

	virt := s.stat.GetVirt(inst.Id)
	if virt == nil {
		return
	}

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

	unit := s.stat.Unit(deply.Unit)
	if unit == nil {
		println("**************************************************2")
		println(inst.Id.Hex())
		println("**************************************************2")
		return
	}

	certs := []*certificate.Certificate{}
	for _, certId := range unit.Instance.Certificates {
		cert := s.stat.ServiceCert(certId)
		if cert.Organization != inst.Organization {
			continue
		}

		certs = append(certs, cert)
	}

	secrs := []*secret.Secret{}
	for _, secrId := range unit.Instance.Secrets {
		secr := s.stat.ServiceSecret(secrId)
		if secr.Organization != inst.Organization {
			continue
		}

		secrs = append(secrs, secr)
	}

	conf, err := imds.BuildConfig(
		inst, virt,
		vc, subnet,
		certs,
		secrs,
	)
	if err != nil {
		return
	}

	// TODO Only write on change
	err = conf.Write(virt)
	if err != nil {
		return
	}

	return
}

func (s *Imds) Deploy(db *database.Database) (err error) {
	instances := s.stat.Instances()

	for _, inst := range instances {
		if inst.Deployment.IsZero() {
			err = s.buildInstance(db, inst)
			if err != nil {
				return
			}
		} else {
			err = s.buildDeployInstance(db, inst)
			if err != nil {
				return
			}
		}
	}

	return
}

func NewImds(stat *state.State) *Imds {
	return &Imds{
		stat: stat,
	}
}
