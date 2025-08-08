package upgrade

import (
	"github.com/pritunl/pritunl-cloud/database"
)

func Upgrade() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	err = nodeUpgrade(db)
	if err != nil {
		return
	}

	err = rolesUpgrade(db)
	if err != nil {
		return
	}

	err = instanceUpgrade(db)
	if err != nil {
		return
	}

	err = createdUpgrade(db)
	if err != nil {
		return
	}

	err = zoneDatacenterUpgrade(db)
	if err != nil {
		return
	}

	err = instStateUpgrade(db)
	if err != nil {
		return
	}

	err = objectIdUpgrade(db)
	if err != nil {
		return
	}

	return
}
