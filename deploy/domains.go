package deploy

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/state"
)

type Domains struct {
	stat *state.State
}

func (d *Domains) create(db *database.Database, inst *instance.Instance) {
	pubAddr6 := ""
	if inst.PublicIps6 != nil && len(inst.PublicIps6) > 0 {
		pubAddr6 = inst.PublicIps6[0]
	}

	if pubAddr6 == "" {
		return
	}

	logrus.WithFields(logrus.Fields{
		"instance": inst.Id.Hex(),
		"address6": pubAddr6,
	}).Info("deploy: Creating domain record")

	recrd := &domain.Record{
		Organization: inst.Organization,
		Domain:       inst.Domain,
		Node:         node.Self.Id,
		Name:         inst.Name,
		Instance:     inst.Id,
		Timestamp:    time.Now(),
	}

	err := recrd.Upsert(db, "", pubAddr6)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"instance": recrd.Instance.Hex(),
			"error":    err,
		}).Error("deploy: Failed to create domain record")

		return
	}

	return
}

func (d *Domains) update(db *database.Database, recrd *domain.Record,
	addr, addr6 string) {

	logrus.WithFields(logrus.Fields{
		"record":       recrd.Id.Hex(),
		"instance":     recrd.Instance.Hex(),
		"cur_address6": recrd.Address,
		"new_address6": addr6,
	}).Info("deploy: Updating domain record")

	err := recrd.Upsert(db, addr, addr6)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"record":   recrd.Id.Hex(),
			"instance": recrd.Instance.Hex(),
			"error":    err,
		}).Error("deploy: Failed to update domain record")

		return
	}

	return
}

func (d *Domains) remove(db *database.Database, recrd *domain.Record) {
	logrus.WithFields(logrus.Fields{
		"record":   recrd.Id.Hex(),
		"instance": recrd.Instance.Hex(),
	}).Info("deploy: Removing domain record")

	err := recrd.Remove(db)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"record":   recrd.Id.Hex(),
			"instance": recrd.Instance.Hex(),
			"error":    err,
		}).Error("deploy: Failed to remove domain record")

		return
	}

	err = domain.RemoveRecord(db, recrd.Id)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"record":   recrd.Id.Hex(),
			"instance": recrd.Instance.Hex(),
			"error":    err,
		}).Error("deploy: Failed to remove domain record")

		return
	}

	return
}

func (d *Domains) Deploy() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	instances := d.stat.Instances()

	for _, inst := range instances {
		recrds := d.stat.DomainRecords(inst.Id)
		if recrds != nil {
			var curRecrd *domain.Record

			for _, recrd := range recrds {
				if curRecrd == nil &&
					inst.Domain == recrd.Domain &&
					inst.Name == recrd.Name {

					curRecrd = recrd
				} else {
					d.remove(db, recrd)
				}
			}

			if curRecrd != nil {
				pubAddr6 := ""
				if inst.PublicIps6 != nil && len(inst.PublicIps6) > 0 {
					pubAddr6 = inst.PublicIps6[0]
				}

				if pubAddr6 == "" {
					d.remove(db, curRecrd)
					continue
				} else if pubAddr6 != curRecrd.Address6 {
					d.update(db, curRecrd, "", pubAddr6)
					continue
				}

				continue
			}
		}

		if !inst.Domain.IsZero() {
			d.create(db, inst)
		}
	}

	return
}

func NewDomains(stat *state.State) *Domains {
	return &Domains{
		stat: stat,
	}
}
