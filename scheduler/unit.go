package scheduler

import (
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/service"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/sirupsen/logrus"
)

type InstanceUnit struct {
	unit  *service.Unit
	nodes shape.Nodes
}

func (u *InstanceUnit) Schedule(db *database.Database) (err error) {
	if u.unit.Kind != service.InstanceKind {
		err = &errortypes.ParseError{
			errors.New("scheduler: Invalid unit kind"),
		}
		return
	}

	if u.unit.Instance == nil {
		err = &errortypes.ParseError{
			errors.New("scheduler: Missing instance data"),
		}
		return
	}

	if u.unit.Instance.Shape.IsZero() {
		err = &errortypes.ParseError{
			errors.New("scheduler: Missing shape"),
		}
		return
	}

	schd := &Scheduler{
		Id:    u.unit.Id,
		Count: u.unit.Count,
	}

	shpe, err := shape.Get(db, u.unit.Instance.Shape)
	if err != nil {
		return
	}

	u.nodes, err = shpe.GetAllNodes(db, u.unit.Instance.Processors,
		u.unit.Instance.Memory)
	if err != nil {
		return
	}

	if len(u.nodes) == 0 {
		logrus.WithFields(logrus.Fields{
			"service": u.unit.Service.Id.Hex(),
			"unit":    u.unit.Id.Hex(),
			"shape":   u.unit.Instance.Shape.Hex(),
		}).Error("scheduler: Failed to find nodes to schedule")
		return
	}

	primaryNodes, backupNodes := u.processNodes(u.nodes)

	var tickets TicketsStore
	if u.unit.Count < len(primaryNodes) {
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

	return
}

func (u *InstanceUnit) processNodes(nodes shape.Nodes) (
	primaryNodes, backupNodes shape.Nodes) {

	nodes.Sort()

	for _, nde := range nodes {
		if nde.SizeResource(u.unit.Instance.Memory,
			u.unit.Instance.Processors) {

			primaryNodes = append(primaryNodes, nde)
		} else {
			backupNodes = append(backupNodes, nde)
		}
	}

	return
}

func (i *InstanceUnit) scheduleSimple(db *database.Database,
	primaryNodes, backupNodes shape.Nodes) (tickets TicketsStore, err error) {

	tickets = TicketsStore{}
	count := i.unit.Count
	offset := 0

	for _, nde := range primaryNodes {
		if count <= 0 {
			count = i.unit.Count
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
			count = i.unit.Count
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
	primaryNodes, backupNodes shape.Nodes) (tickets TicketsStore, err error) {

	tickets = TicketsStore{}
	count := u.unit.Count
	offset := 0
	overscheduled := 0

	if primaryNodes.Len() != 0 {
		for _, nde := range primaryNodes {
			tickets[nde.Id] = append(tickets[nde.Id], &Ticket{
				Node:   nde.Id,
				Offset: offset,
			})
			count -= 1

			nde.CpuUnitsRes += u.unit.Instance.Processors
			nde.MemoryUnitsRes += u.unit.Instance.MemoryUnits()

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

			nde.CpuUnitsRes += u.unit.Instance.Processors
			nde.MemoryUnitsRes += u.unit.Instance.MemoryUnits()

			if count <= 0 {
				break
			}
		}
	}

	for i := 0; i < OffsetCount; i++ {
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

				nde.CpuUnitsRes += u.unit.Instance.Processors
				nde.MemoryUnitsRes += u.unit.Instance.MemoryUnits()
				break
			}

			if count <= 0 {
				break
			}
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
				overscheduled += 1

				nde.CpuUnitsRes += u.unit.Instance.Processors
				nde.MemoryUnitsRes += u.unit.Instance.MemoryUnits()
				break
			}

			if count <= 0 {
				break
			}
		}

		if count != 0 {
			err = &errortypes.ParseError{
				errors.New("schedule: Count remaining after complex schedule"),
			}
			return
		}

		count = u.unit.Count
		if offset == 0 {
			offset += OffsetInit
		} else {
			offset += OffsetInc
		}
	}

	if overscheduled > 0 {
		logrus.WithFields(logrus.Fields{
			"service":       u.unit.Service.Id.Hex(),
			"unit":          u.unit.Id.Hex(),
			"shape":         u.unit.Instance.Shape.Hex(),
			"overscheduled": overscheduled,
		}).Info("scheduler: Overscheduled unit")
	}

	return
}

func NewInstanceUnit(unit *service.Unit) (
	instUnit *InstanceUnit) {

	instUnit = &InstanceUnit{
		unit: unit,
	}

	return
}
