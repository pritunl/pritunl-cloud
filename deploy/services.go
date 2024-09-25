package deploy

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/service"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/sirupsen/logrus"
)

type Services struct {
	stat *state.State
}

func (s *Services) DeployUnit(db *database.Database,
	unit *service.Unit) (err error) {

	deply := &deployment.Deployment{
		Service: unit.Service.Id,
		Unit:    unit.Id,
		Node:    node.Self.Id,
		Kind:    deployment.Instance,
		State:   deployment.Reserved,
	}

	err = deply.Insert(db)
	if err != nil {
		return
	}

	reserved, err := unit.Reserve(db, deply.Id)
	if err != nil {
		return
	}

	if !reserved {
		err = deployment.Remove(db, deply.Id)
		if err != nil {
			return
		}
		return
	}

	inst := &instance.Instance{
		Organization: unit.Service.Organization,
		Zone:         unit.Instance.Zone,
		Vpc:          unit.Instance.Vpc,
		Subnet:       unit.Instance.Subnet,
		//Shape:               unit.Instance.Shape,
		Node:                node.Self.Id,
		Image:               unit.Instance.Image,
		Uefi:                true,
		SecureBoot:          true,
		Tpm:                 false,
		DhcpServer:          false,
		CloudType:           instance.Linux,
		CloudScript:         "",
		DeleteProtection:    false,
		SkipSourceDestCheck: false,
		Name:                unit.Name,
		Comment:             "",
		InitDiskSize:        10,
		Memory:              2048,
		Processors:          2,
		NetworkRoles:        unit.Instance.Roles,
		NoPublicAddress:     false,
		NoPublicAddress6:    false,
		NoHostAddress:       false,
		Deployment:          deply.Id,
	}

	errData, err := inst.Validate(db)
	if err != nil {
		return
	}

	if errData != nil {
		logrus.WithFields(logrus.Fields{
			"error_code":    errData.Error,
			"error_message": errData.Message,
		}).Error("deploy: Failed to deploy instance")
		return
	}

	err = inst.Insert(db)
	if err != nil {
		return
	}

	deply.State = deployment.Deployed
	deply.Instance = inst.Id
	err = deply.CommitFields(db, set.NewSet("state", "instance"))
	if err != nil {
		return
	}

	return
}

func (s *Services) Deploy(db *database.Database) (err error) {
	units := s.stat.Units()
	deplyIds := s.stat.DeploymentIds()

	for _, unit := range units {
		for _, deply := range unit.Deployments {
			if !deplyIds.Contains(deply.Id) {
				err = unit.RemoveDeployement(db, deply.Id)
				if err != nil {
					return
				}
			}
		}

		if len(unit.Deployments) < unit.Count {
			err = s.DeployUnit(db, unit)
			if err != nil {
				return
			}
		}
	}

	return
}

func NewServices(stat *state.State) *Services {
	return &Services{
		stat: stat,
	}
}
