package deploy

import (
	"github.com/pritunl/pritunl-cloud/database"
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

	deployementId, reserved, err := unit.Reserve(db)
	if err != nil {
		return
	}

	if !reserved {
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

	updated, err := unit.UpdateDeployement(
		db, deployementId, service.Deployed)
	if err != nil {
		return
	}

	if !updated {
		logrus.WithFields(logrus.Fields{
			"service_id":    unit.Service.Id.Hex(),
			"unit_id":       unit.Id.Hex(),
			"instance_name": inst.Name,
		}).Error("deploy: Failed to update instance deployment")
	}

	return
}

func (s *Services) Deploy(db *database.Database) (err error) {
	units := s.stat.Units()

	for _, unit := range units {
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
