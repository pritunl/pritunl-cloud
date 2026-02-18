package task

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/advisory"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
)

var instanceData = &Task{
	Name:       "instance_data",
	Version:    1,
	Hours:      []int{0, 3, 6, 9, 12, 15, 18, 21},
	Minutes:    []int{20},
	Handler:    instanceDataHandler,
	RunOnStart: true,
}

func instanceDataHandler(db *database.Database) (err error) {
	advisories := map[string]*advisory.Advisory{}

	instances, err := instance.GetAll(db, &bson.M{})
	if err != nil {
		return
	}

	for _, inst := range instances {
		if inst.Guest == nil {
			continue
		}

		for _, updt := range inst.Guest.Updates {
			details := []*advisory.Advisory{}

			for _, cve := range updt.Cves {
				adv := advisories[cve]
				if adv == nil {
					for i := 0; i < 3; i++ {
						adv, err = advisory.GetOneLimit(db, cve)
						if err != nil {
							return
						}
					}
				}

				if adv != nil {
					details = append(details, adv)
				}
			}

			updt.Details = details
		}

		err = inst.CommitFields(db, set.NewSet("guest"))
		if err != nil {
			return
		}
	}

	return
}

func init() {
	register(instanceData)
}
