package scheduler

import (
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/unit"
	"github.com/sirupsen/logrus"
)

type InstanceUnit struct {
	unit  *unit.Unit
	spec  *spec.Spec
	count int
	nodes spec.Nodes
}

func (u *InstanceUnit) Schedule(db *database.Database, count int) (err error) {
	if u.unit.Kind != deployment.Instance && u.unit.Kind != deployment.Image {
		err = &errortypes.ParseError{
			errors.New("scheduler: Invalid unit kind"),
		}
		return
	}

	if u.spec.Instance == nil {
		err = &errortypes.ParseError{
			errors.New("scheduler: Missing instance data"),
		}
		return
	}

	if u.spec.Instance.Shape.IsZero() && u.spec.Instance.Node.IsZero() {
		err = &errortypes.ParseError{
			errors.New("scheduler: Missing shape or node"),
		}
		return
	}

	overrideCount := 0
	if count == 0 {
		u.count = u.unit.Count - len(u.unit.Deployments)
	} else {
		u.count = count
		overrideCount = len(u.unit.Deployments) + count
	}

	schd := &Scheduler{
		Id:            u.unit.Id,
		Kind:          InstanceUnitKind,
		Spec:          u.spec.Id,
		Count:         u.count,
		OverrideCount: overrideCount,
		Failures:      map[bson.ObjectID]int{},
	}

	if !u.spec.Instance.Node.IsZero() {
		nde, e := node.Get(db, u.spec.Instance.Node)
		if e != nil {
			err = e
			return
		}
		u.nodes = []*node.Node{nde}
	} else {
		ndes, offlineCount, noMountCount, e := u.spec.GetAllNodes(db)
		if e != nil {
			err = e
			return
		}
		u.nodes = ndes

		if len(u.nodes) == 0 {
			logrus.WithFields(logrus.Fields{
				"unit":                u.unit.Id.Hex(),
				"shape":               u.spec.Instance.Shape.Hex(),
				"offline_count":       offlineCount,
				"missing_mount_count": noMountCount,
			}).Error("scheduler: Failed to find nodes to schedule")
			return
		}
	}

	if u.count == 0 {
		err = &errortypes.ParseError{
			errors.New("scheduler: Cannot schedule zero count unit"),
		}
		return
	}

	primaryNodes, backupNodes := u.processNodes(u.nodes)

	var tickets TicketsStore
	if u.count < len(primaryNodes) {
		tickets, err = u.scheduleSimple(db, primaryNodes, backupNodes)
		if err != nil {
			return
		}
	} else {
		tickets, err = u.scheduleComplex(db, primaryNodes, backupNodes)
		if err != nil {
			return
		}
	}

	schd.Tickets = tickets
	schd.Created = time.Now()
	schd.Modified = time.Now()

	logrus.WithFields(logrus.Fields{
		"unit":          u.unit.Id.Hex(),
		"count":         u.count,
		"primary_nodes": len(primaryNodes),
		"backup_nodes":  len(backupNodes),
		"tickets":       len(tickets),
	}).Info("scheduler: Scheduling unit")

	err = schd.Insert(db)
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		} else {
			return
		}
		return
	}

	return
}

func (u *InstanceUnit) processNodes(nodes spec.Nodes) (
	primaryNodes, backupNodes spec.Nodes) {

	nodes.Sort()

	for _, nde := range nodes {
		if nde.SizeResource(u.spec.Instance.Memory,
			u.spec.Instance.Processors) {

			primaryNodes = append(primaryNodes, nde)
		} else {
			backupNodes = append(backupNodes, nde)
		}
	}

	return
}

func (u *InstanceUnit) scheduleSimple(db *database.Database,
	primaryNodes, backupNodes spec.Nodes) (tickets TicketsStore, err error) {

	tickets = TicketsStore{}
	count := u.count
	offset := 0

	for _, nde := range primaryNodes {
		if count <= 0 {
			count = u.count
			if offset == 0 {
				offset += OffsetInit
			} else {
				offset += OffsetInc
			}
		}

		tickets[nde.Id] = append(tickets[nde.Id], &Ticket{
			Node:   nde.Id,
			Offset: offset,
		})
		count -= 1
	}

	for _, nde := range backupNodes {
		if count <= 0 {
			count = u.count
			if offset == 0 {
				offset += OffsetInit
			} else {
				offset += OffsetInc
			}
		}

		tickets[nde.Id] = append(tickets[nde.Id], &Ticket{
			Node:   nde.Id,
			Offset: offset,
		})
		count -= 1
	}

	return
}

func (u *InstanceUnit) scheduleComplex(db *database.Database,
	primaryNodes, backupNodes spec.Nodes) (tickets TicketsStore, err error) {

	tickets = TicketsStore{}
	count := u.count
	offset := 0
	overscheduled := 0

	if primaryNodes.Len() != 0 {
		for _, nde := range primaryNodes {
			tickets[nde.Id] = append(tickets[nde.Id], &Ticket{
				Node:   nde.Id,
				Offset: offset,
			})
			count -= 1

			nde.CpuUnitsRes += u.spec.Instance.Processors
			nde.MemoryUnitsRes += u.spec.Instance.MemoryUnits()

			if count <= 0 {
				break
			}
		}
	} else {
		for _, nde := range backupNodes {
			tickets[nde.Id] = append(tickets[nde.Id], &Ticket{
				Node:   nde.Id,
				Offset: offset,
			})
			count -= 1
			overscheduled += 1

			nde.CpuUnitsRes += u.spec.Instance.Processors
			nde.MemoryUnitsRes += u.spec.Instance.MemoryUnits()

			if count <= 0 {
				break
			}
		}
	}

	for i := 0; i < OffsetCount; i++ {
		attempts := 0
		for attempts = 0; attempts < 100; attempts++ {
			if count <= 0 {
				break
			}

			for {
				primaryNodes, _ = u.processNodes(u.nodes)
				if primaryNodes.Len() == 0 {
					break
				}

				for _, nde := range primaryNodes {
					tickets[nde.Id] = append(tickets[nde.Id], &Ticket{
						Node:   nde.Id,
						Offset: offset,
					})
					count -= 1

					nde.CpuUnitsRes += u.spec.Instance.Processors
					nde.MemoryUnitsRes += u.spec.Instance.MemoryUnits()
					break
				}

				if count <= 0 {
					break
				}
			}

			if count <= 0 {
				break
			}

			for {
				_, backupNodes = u.processNodes(u.nodes)
				if backupNodes.Len() == 0 {
					break
				}

				for _, nde := range backupNodes {
					tickets[nde.Id] = append(tickets[nde.Id], &Ticket{
						Node:   nde.Id,
						Offset: offset,
					})
					count -= 1
					if i == 0 {
						overscheduled += 1
					}

					nde.CpuUnitsRes += u.spec.Instance.Processors
					nde.MemoryUnitsRes += u.spec.Instance.MemoryUnits()
					break
				}

				if count <= 0 {
					break
				}
			}
		}

		if count != 0 {
			err = &errortypes.ParseError{
				errors.Newf("schedule: Count %d remaining after %d "+
					"complex schedule attempts", count, attempts),
			}
			return
		}

		count = u.count
		if offset == 0 {
			offset += OffsetInit
		} else {
			offset += OffsetInc
		}
	}

	if overscheduled > 0 {
		logrus.WithFields(logrus.Fields{
			"unit":          u.unit.Id.Hex(),
			"kind":          u.unit.Kind,
			"shape":         u.spec.Instance.Shape.Hex(),
			"overscheduled": overscheduled,
		}).Info("scheduler: Overscheduled unit")
	}

	return
}

func NewInstanceUnit(unt *unit.Unit, spc *spec.Spec) (
	instUnit *InstanceUnit) {

	instUnit = &InstanceUnit{
		unit: unt,
		spec: spc,
	}

	return
}
