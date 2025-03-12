package upgrade

import (
	"github.com/pritunl/pritunl-cloud/database"
)

func Upgrade() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	err = zoneDatacenterUpgrade(db)
	if err != nil {
		return
	}

	err = actionUpgrade(db)
	if err != nil {
		return
	}

	return
}
