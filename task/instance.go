package task

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
)

var instanceData = &Task{
	Name:    "instance_data",
	Version: 1,
	Hours:   []int{0, 3, 6, 9, 12, 15, 18, 21},
	Minutes: []int{20},
	Handler: instanceDataHandler,
}

func instanceDataHandler(db *database.Database) (err error) {
	instances, err := instance.GetAll(db, &bson.M{})
	if err != nil {
		return
	}

	for _, inst := range instances {
		_ = inst
	}

	return
}

func init() {
	register(instanceData)
}
