package nodeport

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type Mapping struct {
	NodePort     primitive.ObjectID `bson:"node_port" json:"node_port"`
	Protocol     string             `bson:"protocol" json:"protocol"`
	ExternalPort int                `bson:"external_port" json:"external_port"`
	InternalPort int                `bson:"internal_port" json:"internal_port"`
	Delete       bool               `bson:"-" json:"delete"`
}

func (m *Mapping) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	switch m.Protocol {
	case Tcp, Udp:
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "invalid_protocol",
			Message: "Invalid node port protocol",
		}
		return
	}

	if m.ExternalPort != 0 {
		portRanges, e := GetPortRanges()
		if e != nil {
			err = e
			return
		}

		matched := false
		for _, ports := range portRanges {
			if ports.Contains(m.ExternalPort) {
				matched = true
				break
			}
		}

		if !matched {
			errData = &errortypes.ErrorData{
				Error:   "invalid_external_port",
				Message: "Invalid external node port",
			}
			return
		}
	}

	if m.InternalPort <= 0 || m.InternalPort > 65535 {
		errData = &errortypes.ErrorData{
			Error:   "invalid_internal_port",
			Message: "Invalid internal node port",
		}
		return
	}

	return
}

func (m *Mapping) Diff(mapping *Mapping) bool {
	if m.Protocol != mapping.Protocol {
		return true
	}

	if m.ExternalPort != mapping.ExternalPort {
		return true
	}

	if m.InternalPort != mapping.InternalPort {
		return true
	}

	return false
}
