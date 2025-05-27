package virtiofs

import (
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/organization"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
)

func StartAll(db *database.Database,
	virt *vm.VirtualMachine) (err error) {

	mounts := []*vm.Mount{}

	unitPathShares := paths.GetUnitPathShares(virt.Id)
	_, err = utils.RemoveWildcard(unitPathShares)
	if err != nil {
		return
	}

	if len(virt.Mounts) == 0 {
		return
	}

	org, err := organization.Get(db, virt.Organization)
	if err != nil {
		return
	}

	if org == nil {
		err = &errortypes.ParseError{
			errors.New("virtiofs: Failed to get org"),
		}
		return
	}

	for _, mount := range virt.Mounts {
		matchPath := false
		matchRoles := false

		for _, share := range node.Self.Shares {
			if share.MatchPath(mount.HostPath) {
				matchPath = true

				if utils.HasMatchingItem(share.Roles, org.Roles) {
					matchRoles = true
					break
				}
			}
		}

		if !matchPath && !matchRoles {
			err = &errortypes.ParseError{
				errors.Newf("virtiofs: Failed to find matching "+
					"share path for mount '%s'", mount.HostPath),
			}
			return
		}

		if !matchPath || !matchRoles {
			err = &errortypes.ParseError{
				errors.Newf("virtiofs: Failed to find matching "+
					"role for mount '%s'", mount.HostPath),
			}
			return
		}

		mounts = append(mounts, mount)
	}

	for _, mount := range mounts {
		shareId := paths.GetShareId(virt.Id, mount.Name)

		err = Start(db, virt, shareId, mount.HostPath)
		if err != nil {
			return
		}
	}

	time.Sleep(1 * time.Second)

	return
}

func StopAll(virt *vm.VirtualMachine) (err error) {
	unit := paths.GetUnitNameShares(virt.Id)

	_ = systemd.Stop(unit)

	return
}
