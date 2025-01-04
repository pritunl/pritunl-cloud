package deployment

import (
	"fmt"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Deployment struct {
	Id           primitive.ObjectID             `bson:"_id,omitempty" json:"id"`
	Pod          primitive.ObjectID             `bson:"pod" json:"pod"`
	Unit         primitive.ObjectID             `bson:"unit" json:"unit"`
	Spec         primitive.ObjectID             `bson:"spec" json:"spec"`
	Kind         string                         `bson:"kind" json:"kind"`
	State        string                         `bson:"state" json:"state"`
	Status       string                         `bson:"status" json:"status"`
	Zone         primitive.ObjectID             `bson:"zone,omitempty" json:"zone"`
	Node         primitive.ObjectID             `bson:"node,omitempty" json:"node"`
	Instance     primitive.ObjectID             `bson:"instance,omitempty" json:"instance"`
	Image        primitive.ObjectID             `bson:"image,omitempty" json:"image"`
	InstanceData *InstanceData                  `bson:"instance_data,omitempty" json:"instance_data"`
	ImageData    *ImageData                     `bson:"image_data,omitempty" json:"image_data"`
	Actions      map[primitive.ObjectID]*Action `bson:"actions,omitempty", json:"actions"`
}

type InstanceData struct {
	PublicIps        []string `bson:"public_ips" json:"public_ips"`
	PublicIps6       []string `bson:"public_ips6" json:"public_ips6"`
	PrivateIps       []string `bson:"private_ips" json:"private_ips"`
	PrivateIps6      []string `bson:"private_ips6" json:"private_ips6"`
	OraclePrivateIps []string `bson:"oracle_private_ips" json:"oracle_private_ips"`
	OraclePublicIps  []string `bson:"oracle_public_ips" json:"oracle_public_ips"`
}

type ImageData struct {
	State string `bson:"state" json:"state"`
}

type Action struct {
	Statement primitive.ObjectID `bson:"statement" json:"statement"`
	Since     time.Time          `bson:"since" json:"since"`
	Executed  time.Time          `bson:"executed" json:"executed"`
	Action    string             `bson:"action" json:"action"`
}

func (d *Deployment) IsHealthy() bool {
	return d.Status == Healthy
}

func (d *Deployment) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if d.Actions == nil {
		d.Actions = map[primitive.ObjectID]*Action{}
	}

	switch d.State {
	case Reserved:
		break
	case Deployed:
		break
	case Destroy:
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "invalid_state",
			Message: "Deployment state is invalid",
		}
		return
	}

	switch d.Status {
	case Healthy:
		break
	case Unhealthy:
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
	statementId primitive.ObjectID, thresholdSec int, action string) (
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

	d.Id = resp.InsertedID.(primitive.ObjectID)

	return
}
