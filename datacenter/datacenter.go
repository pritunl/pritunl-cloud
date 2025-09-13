package datacenter

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Datacenter struct {
	Id                  bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name                string          `bson:"name" json:"name"`
	Comment             string          `bson:"comment" json:"comment"`
	MatchOrganizations  bool            `bson:"match_organizations" json:"match_organizations"`
	Organizations       []bson.ObjectID `bson:"organizations" json:"organizations"`
	NetworkMode         string          `bson:"network_mode" json:"network_mode"`
	WgMode              string          `bson:"wg_mode" json:"wg_mode"`
	JumboMtu            int             `bson:"jumbo_mtu" json:"jumbo_mtu"`
	PublicStorages      []bson.ObjectID `bson:"public_storages" json:"public_storages"`
	PrivateStorage      bson.ObjectID   `bson:"private_storage,omitempty" json:"private_storage"`
	PrivateStorageClass string          `bson:"private_storage_class" json:"private_storage_class"`
	BackupStorage       bson.ObjectID   `bson:"backup_storage,omitempty" json:"backup_storage"`
	BackupStorageClass  string          `bson:"backup_storage_class" json:"backup_storage_class"`
}

type Completion struct {
	Id          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string        `bson:"name" json:"name"`
	NetworkMode string        `bson:"network_mode" json:"network_mode"`
}

func (d *Datacenter) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	d.Name = utils.FilterName(d.Name)

	if d.Organizations == nil || !d.MatchOrganizations {
		d.Organizations = []bson.ObjectID{}
	}

	if d.PublicStorages == nil {
		d.PublicStorages = []bson.ObjectID{}
	}

	switch d.NetworkMode {
	case Default:
		break
	case VxlanVlan:
		break
	case WgVxlanVlan:
		break
	case "":
		d.NetworkMode = Default
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "invalid_network_mode",
			Message: "Network mode invalid",
		}
		return
	}

	if d.NetworkMode == WgVxlanVlan {
		switch d.WgMode {
		case Wg4:
			break
		case Wg6:
			break
		case "":
			d.WgMode = Wg4
			break
		default:
			errData = &errortypes.ErrorData{
				Error:   "invalid_wg_mode",
				Message: "WireGuard mode invalid",
			}
			return
		}
	} else {
		d.WgMode = ""
	}

	return
}

func (d *Datacenter) Vxlan() bool {
	return d.NetworkMode == VxlanVlan || d.NetworkMode == WgVxlanVlan
}

func (d *Datacenter) GetBaseInternalMtu() (mtuSize int) {
	if node.Self.JumboFrames || node.Self.JumboFramesInternal {
		mtuSize = settings.Hypervisor.JumboMtu
	} else {
		mtuSize = settings.Hypervisor.NormalMtu
	}
	return
}

func (d *Datacenter) GetBaseExternalMtu() (mtuSize int) {
	if node.Self.JumboFrames {
		mtuSize = settings.Hypervisor.JumboMtu
	} else {
		mtuSize = settings.Hypervisor.NormalMtu
	}
	return
}

func (d *Datacenter) GetOverlayMtu() (mtuSize int) {
	if d.NetworkMode == WgVxlanVlan {
		if node.Self.JumboFrames {
			mtuSize = settings.Hypervisor.JumboMtu
		} else {
			mtuSize = settings.Hypervisor.NormalMtu
		}

		if d.WgMode == Wg6 {
			mtuSize -= 150
		} else {
			mtuSize -= 110
		}
	} else {
		if node.Self.JumboFrames || node.Self.JumboFramesInternal {
			mtuSize = settings.Hypervisor.JumboMtu
		} else {
			mtuSize = settings.Hypervisor.NormalMtu
		}

		if d.NetworkMode == VxlanVlan {
			mtuSize -= 50
		}
	}

	return
}

func (d *Datacenter) GetInstanceMtu() (mtuSize int) {
	if d.NetworkMode == WgVxlanVlan {
		if node.Self.JumboFrames {
			mtuSize = settings.Hypervisor.JumboMtu
		} else {
			mtuSize = settings.Hypervisor.NormalMtu
		}

		if d.WgMode == Wg6 {
			mtuSize -= 154
		} else {
			mtuSize -= 114
		}
	} else {
		if node.Self.JumboFrames || node.Self.JumboFramesInternal {
			mtuSize = settings.Hypervisor.JumboMtu
		} else {
			mtuSize = settings.Hypervisor.NormalMtu
		}

		if d.NetworkMode == VxlanVlan {
			mtuSize -= 54
		}
	}

	return
}

func (d *Datacenter) Commit(db *database.Database) (err error) {
	coll := db.Datacenters()

	err = coll.Commit(d.Id, d)
	if err != nil {
		return
	}

	return
}

func (d *Datacenter) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Datacenters()

	err = coll.CommitFields(d.Id, d, fields)
	if err != nil {
		return
	}

	return
}

func (d *Datacenter) Insert(db *database.Database) (err error) {
	coll := db.Datacenters()

	if !d.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("datacenter: Datacenter already exists"),
		}
		return
	}

	resp, err := coll.InsertOne(db, d)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	d.Id = resp.InsertedID.(bson.ObjectID)

	return
}
