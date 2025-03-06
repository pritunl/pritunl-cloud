package task

import (
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/pritunl-cloud/balancer"
	"github.com/pritunl/pritunl-cloud/database"
)

var balancerClean = &Task{
	Name:    "balancer_clean",
	Version: 1,
	Hours: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
	Minutes: []int{35},
	Handler: balancerCleanHandler,
}

func balancerCleanHandler(db *database.Database) (err error) {
	balcns, err := balancer.GetAll(db, &bson.M{})

	for _, balnc := range balcns {
		err = balnc.Clean(db)
		if err != nil {
			return
		}
	}

	return
}

func init() {
	register(balancerClean)
}
