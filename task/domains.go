package task

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/settings"
)

var domains = &Task{
	Name:    "domains",
	Version: 1,
	Hours: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
	Minutes: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25,
		26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38,
		39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51,
		52, 53, 54, 55, 56, 57, 58, 59},
	Handler: domainsHandler,
}

func domainsHandler(db *database.Database) (err error) {
	refreshTtl := time.Duration(
		settings.System.DomainRefreshTtl) * time.Second

	domns, err := domain.GetAll(db, &bson.M{
		"last_update": &bson.M{
			"$gte": time.Now().Add(-refreshTtl),
		},
	})
	if err != nil {
		return
	}

	for _, domn := range domns {
		domain.Refresh(db, domn.Id)
	}

	return
}

func init() {
	register(domains)
}
