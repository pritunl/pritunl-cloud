package task

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/advisory"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/sirupsen/logrus"
)

var advisoryData = &Task{
	Name:       "advisory",
	Version:    1,
	Hours:      []int{0, 3, 6, 9, 12, 15, 18, 21},
	Minutes:    []int{22},
	Handler:    advisoryDataHandler,
	RunOnStart: true,
}

func advisoryDataHandler(db *database.Database) (err error) {
	advisories := map[string]*advisory.Advisory{}

	coll := db.Instances()

	cursor, err := coll.Find(db, &bson.M{})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		inst := &instance.Instance{}
		err = cursor.Decode(inst)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		if inst.Guest == nil {
			continue
		}

		for _, updt := range inst.Guest.Updates {
			details := []*advisory.Advisory{}

			for _, cve := range updt.Cves {
				adv, ok := advisories[cve]
				if !ok {
					adv, err = advisory.GetOneLimit(db, cve)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"cve_id": cve,
							"error":  err,
						}).Error("task: Failed to query CVE")
						err = nil
						adv = nil
					}
					advisories[cve] = adv
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

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func init() {
	register(advisoryData)
}
