package deployment

import (
	"fmt"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Deployment struct {
	Id           bson.ObjectID             `bson:"_id,omitempty" json:"id"`
	Pod          bson.ObjectID             `bson:"pod" json:"pod"`
	Unit         bson.ObjectID             `bson:"unit" json:"unit"`
	Organization bson.ObjectID             `bson:"organization" json:"organization"`
	Timestamp    time.Time                 `bson:"timestamp" json:"timestamp"`
	Tags         []string                  `bson:"tags" json:"tags"`
	Spec         bson.ObjectID             `bson:"spec" json:"spec"`
	NewSpec      bson.ObjectID             `bson:"new_spec" json:"new_spec"`
	Kind         string                    `bson:"kind" json:"kind"`
	State        string                    `bson:"state" json:"state"`
	Action       string                    `bson:"action" json:"action"`
	Status       string                    `bson:"status" json:"status"`
	Datacenter   bson.ObjectID             `bson:"datacenter" json:"datacenter"`
	Zone         bson.ObjectID             `bson:"zone" json:"zone"`
	Node         bson.ObjectID             `bson:"node" json:"node"`
	Instance     bson.ObjectID             `bson:"instance" json:"instance"`
	Image        bson.ObjectID             `bson:"image" json:"image"`
	Mounts       []*Mount                  `bson:"mounts" json:"mounts"`
	Journals     []*Journal                `bson:"journals" json:"journals"`
	InstanceData *InstanceData             `bson:"instance_data,omitempty" json:"instance_data"`
	ImageData    *ImageData                `bson:"image_data,omitempty" json:"image_data"`
	DomainData   *DomainData               `bson:"domain_data,omitempty" json:"domain_data"`
	Actions      map[bson.ObjectID]*Action `bson:"actions,omitempty" json:"actions"`
}

type InstanceData struct {
	HostIps         []string `bson:"host_ips" json:"host_ips"`
	PublicIps       []string `bson:"public_ips" json:"public_ips"`
	PublicIps6      []string `bson:"public_ips6" json:"public_ips6"`
	PrivateIps      []string `bson:"private_ips" json:"private_ips"`
	PrivateIps6     []string `bson:"private_ips6" json:"private_ips6"`
	CloudPrivateIps []string `bson:"cloud_private_ips" json:"cloud_private_ips"`
	CloudPublicIps  []string `bson:"cloud_public_ips" json:"cloud_public_ips"`
	CloudPublicIps6 []string `bson:"cloud_public_ips6" json:"cloud_public_ips6"`
}

type DomainData struct {
	Records []*RecordData `bson:"records" json:"records"`
}

type RecordData struct {
	Domain string `bson:"domain" json:"domain"`
	Value  string `bson:"value" json:"value"`
}

type ImageData struct {
	State string `bson:"state" json:"state"`
}

type Mount struct {
	Disk bson.ObjectID `bson:"disk" json:"disk"`
	Path string        `bson:"path" json:"path"`
	Uuid string        `bson:"uuid" json:"uuid"`
}

type Journal struct {
	Index int32  `bson:"index" json:"index"`
	Key   string `bson:"key" json:"key"`
	Type  string `bson:"type" json:"type"`
}

type Action struct {
	Statement bson.ObjectID `bson:"statement" json:"statement"`
	Since     time.Time     `bson:"since" json:"since"`
	Executed  time.Time     `bson:"executed" json:"executed"`
	Action    string        `bson:"action" json:"action"`
}

func (d *Deployment) IsHealthy() bool {
	return d.Status == Healthy
}

func (d *Deployment) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if d.Timestamp.IsZero() {
		d.Timestamp = time.Now()
	}

	if d.Tags == nil {
		d.Tags = []string{}
	}

	tags := []string{}
	for _, tag := range d.Tags {
		tag = utils.FilterName(tag)
		if tag == "" || tag == "latest" {
			continue
		}

		tags = append(tags, tag)
	}
	d.Tags = tags

	if d.Actions == nil {
		d.Actions = map[bson.ObjectID]*Action{}
	}

	if !ValidStates.Contains(d.State) {
		errData = &errortypes.ErrorData{
			Error:   "invalid_state",
			Message: "Invalid deployment state",
		}
		return
	}

	if !ValidActions.Contains(d.Action) {
		errData = &errortypes.ErrorData{
			Error:   "invalid_action",
			Message: "Invalid deployment action",
		}
		return
	}

	switch d.Status {
	case Healthy:
		break
	case Unhealthy:
		break
	case Unknown:
		break
	case "":
		d.Status = Unhealthy
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "invalid_status",
			Message: "Deployment status is invalid",
		}
		return
	}

	switch d.Kind {
	case Instance:
		break
	case Image:
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "invalid_kind",
			Message: "Deployment kind is invalid",
		}
		return
	}

	return
}

func (d *Deployment) HandleStatement(db *database.Database,
	statementId bson.ObjectID, thresholdSec int, action string) (
	newAction string, err error) {

	thresholdSec = utils.Max(ThresholdMin, thresholdSec)
	threshold := time.Duration(thresholdSec) * time.Second

	if action != "" {
		curAction := d.Actions[statementId]
		if curAction == nil {
			err = d.CommitAction(db, &Action{
				Statement: statementId,
				Since:     time.Now(),
				Action:    action,
			})
			if err != nil {
				return
			}

			newAction = ""
			return
		} else if curAction.Action != action {
			if !curAction.Executed.IsZero() && time.Since(
				curAction.Executed) < ActionLimit {

				newAction = ""
				return
			}

			curAction.Since = time.Now()
			curAction.Executed = time.Time{}
			curAction.Action = action

			err = d.CommitAction(db, curAction)
			if err != nil {
				return
			}

			newAction = ""
			return
		} else if time.Since(curAction.Since) >= threshold {
			if !curAction.Executed.IsZero() && time.Since(
				curAction.Executed) < ActionLimit {

				newAction = ""
				return
			}

			curAction.Executed = time.Now()

			err = d.CommitAction(db, curAction)
			if err != nil {
				return
			}

			newAction = action
			return
		}
	} else {
		curAction := d.Actions[statementId]
		if curAction != nil {
			if !curAction.Executed.IsZero() && time.Since(
				curAction.Executed) < ActionLimit {

				newAction = ""
				return
			}

			err = d.RemoveAction(db, curAction)
			if err != nil {
				return
			}
		}

		newAction = ""
		return
	}

	return
}

func (d *Deployment) SetImageState(state string) {
	if d.ImageData == nil {
		d.ImageData = &ImageData{}
	}
	d.ImageData.State = state
}

func (d *Deployment) GetImageState() string {
	if d.ImageData == nil {
		return ""
	}
	return d.ImageData.State
}

func (d *Deployment) ImageReady() bool {
	return !d.Image.IsZero() && d.ImageData != nil &&
		d.ImageData.State == Complete
}

func (d *Deployment) CommitAction(db *database.Database,
	action *Action) (err error) {

	coll := db.Deployments()

	_, err = coll.UpdateOne(db, bson.M{
		"_id": d.Id,
	}, bson.M{
		"$set": bson.M{
			fmt.Sprintf("actions.%s", action.Statement.Hex()): action,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (d *Deployment) RemoveAction(db *database.Database,
	action *Action) (err error) {

	coll := db.Deployments()

	_, err = coll.UpdateOne(db, bson.M{
		"_id": d.Id,
	}, bson.M{
		"$unset": bson.M{
			fmt.Sprintf("actions.%s", action.Statement.Hex()): "",
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (d *Deployment) Commit(db *database.Database) (err error) {
	coll := db.Deployments()

	err = coll.Commit(d.Id, d)
	if err != nil {
		return
	}

	return
}

func (d *Deployment) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Deployments()

	err = coll.CommitFields(d.Id, d, fields)
	if err != nil {
		return
	}

	return
}

func (d *Deployment) Insert(db *database.Database) (err error) {
	coll := db.Deployments()

	resp, err := coll.InsertOne(db, d)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	d.Id = resp.InsertedID.(bson.ObjectID)

	return
}
