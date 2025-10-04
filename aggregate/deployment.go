package aggregate

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/journal"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/unit"
	"github.com/pritunl/pritunl-cloud/zone"
)

type DeploymentPipe struct {
	Deployment   `bson:",inline"`
	SpecDocs     []*spec.Spec         `bson:"spec_docs"`
	InstanceDocs []*instance.Instance `bson:"instance_docs"`
	ZoneDocs     []*zone.Zone         `bson:"zone_docs"`
	NodeDocs     []*node.Node         `bson:"node_docs"`
	ImageDocs    []*image.Image       `bson:"image_docs"`
}

type Deployment struct {
	Id                  bson.ObjectID            `bson:"_id" json:"id"`
	Pod                 bson.ObjectID            `bson:"pod" json:"pod"`
	Unit                bson.ObjectID            `bson:"unit" json:"unit"`
	Spec                bson.ObjectID            `bson:"spec" json:"spec"`
	SpecOffset          int                      `bson:"spec_offset" json:"spec_offset"`
	SpecIndex           int                      `bson:"spec_index" json:"spec_index"`
	SpecTimestamp       time.Time                `bson:"spec_timestamp" json:"spec_timestamp"`
	Timestamp           time.Time                `bson:"timestamp" json:"timestamp"`
	Tags                []string                 `bson:"tags" json:"tags"`
	Kind                string                   `bson:"kind" json:"kind"`
	State               string                   `bson:"state" json:"state"`
	Action              string                   `bson:"action" json:"action"`
	Status              string                   `bson:"status" json:"status"`
	Node                bson.ObjectID            `bson:"node" json:"node"`
	Instance            bson.ObjectID            `bson:"instance" json:"instance"`
	InstanceData        *deployment.InstanceData `bson:"instance_data" json:"instance_data"`
	DomainData          *deployment.DomainData   `bson:"domain_data" json:"domain_data"`
	Journals            []*deployment.Journal    `bson:"journals" json:"journals"`
	ImageId             bson.ObjectID            `bson:"image_id" json:"image_id"`
	ImageName           string                   `bson:"image_name" json:"image_name"`
	ZoneName            string                   `bson:"-" json:"zone_name"`
	NodeName            string                   `bson:"-" json:"node_name"`
	InstanceName        string                   `bson:"-" json:"instance_name"`
	InstanceRoles       []string                 `bson:"-" json:"instance_roles"`
	InstanceMemory      int                      `bson:"-" json:"instance_memory"`
	InstanceProcessors  int                      `bson:"-" json:"instance_processors"`
	InstanceStatus      string                   `bson:"-" json:"instance_status"`
	InstanceUptime      string                   `bson:"-" json:"instance_uptime"`
	InstanceState       string                   `bson:"-" json:"instance_state"`
	InstanceAction      string                   `bson:"-" json:"instance_action"`
	InstanceGuestStatus string                   `bson:"-" json:"instance_guest_status"`
	InstanceTimestamp   time.Time                `bson:"-" json:"instance_timestamp"`
	InstanceHeartbeat   time.Time                `bson:"-" json:"instance_heartbeat"`
	InstanceMemoryUsage float64                  `bson:"-" json:"instance_memory_usage"`
	InstanceHugePages   float64                  `bson:"-" json:"instance_hugepages"`
	InstanceLoad1       float64                  `bson:"-" json:"instance_load1"`
	InstanceLoad5       float64                  `bson:"-" json:"instance_load5"`
	InstanceLoad15      float64                  `bson:"-" json:"instance_load15"`
}

func GetDeployments(db *database.Database, unt *unit.Unit) (
	deplys []*Deployment, err error) {

	coll := db.Deployments()
	deplys = []*Deployment{}

	cursor, err := coll.Aggregate(db, []*bson.M{
		&bson.M{
			"$match": &bson.M{
				"unit": unt.Id,
			},
		},
		&bson.M{
			"$sort": &bson.M{
				"timestamp": -1,
			},
		},
		&bson.M{
			"$lookup": &bson.M{
				"from":         "specs",
				"localField":   "spec",
				"foreignField": "_id",
				"as":           "spec_docs",
			},
		},
		&bson.M{
			"$lookup": &bson.M{
				"from":         "instances",
				"localField":   "instance",
				"foreignField": "_id",
				"as":           "instance_docs",
			},
		},
		&bson.M{
			"$lookup": &bson.M{
				"from":         "zones",
				"localField":   "zone",
				"foreignField": "_id",
				"as":           "zone_docs",
			},
		},
		&bson.M{
			"$lookup": &bson.M{
				"from":         "nodes",
				"localField":   "node",
				"foreignField": "_id",
				"as":           "node_docs",
			},
		},
		&bson.M{
			"$lookup": &bson.M{
				"from":         "images",
				"localField":   "_id",
				"foreignField": "deployment",
				"as":           "image_docs",
			},
		},
		&bson.M{
			"$project": &bson.D{
				{"_id", 1},
				{"pod", 1},
				{"unit", 1},
				{"timestamp", 1},
				{"tags", 1},
				{"spec", 1},
				{"kind", 1},
				{"state", 1},
				{"action", 1},
				{"status", 1},
				{"node", 1},
				{"instance", 1},
				{"instance_data", 1},
				{"domain_data", 1},
				{"journals", 1},
				{"spec_docs.index", 1},
				{"spec_docs.timestamp", 1},
				{"instance_docs.name", 1},
				{"instance_docs.roles", 1},
				{"instance_docs.memory", 1},
				{"instance_docs.processors", 1},
				{"instance_docs.state", 1},
				{"instance_docs.action", 1},
				{"instance_docs.timestamp", 1},
				{"instance_docs.restart", 1},
				{"instance_docs.restart_block_ip", 1},
				{"instance_docs.guest", 1},
				{"zone_docs.name", 1},
				{"node_docs.name", 1},
				{"image_docs._id", 1},
				{"image_docs.name", 1},
			},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	latest := true
	for cursor.Next(db) {
		doc := &DeploymentPipe{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		deply := &doc.Deployment

		if deply.Tags == nil {
			deply.Tags = []string{}
		}
		if latest {
			latest = false
			if doc.Kind == deployment.Image {
				deply.Tags = append([]string{"latest"}, deply.Tags...)
			}
		}

		deply.Journals = append([]*deployment.Journal{
			{
				Index: journal.DeploymentAgent,
				Key:   "agent",
				Type:  "agent",
			},
		}, deply.Journals...)

		if len(doc.ZoneDocs) > 0 {
			zne := doc.ZoneDocs[0]
			deply.ZoneName = zne.Name
		}

		if len(doc.NodeDocs) > 0 {
			nde := doc.NodeDocs[0]
			deply.NodeName = nde.Name
		}

		if len(doc.SpecDocs) > 0 {
			spc := doc.SpecDocs[0]

			deply.SpecOffset = spc.Index - unt.SpecIndex
			deply.SpecIndex = spc.Index
			deply.SpecTimestamp = spc.Timestamp
		}

		if len(doc.InstanceDocs) > 0 {
			inst := doc.InstanceDocs[0]
			inst.Json(true)

			deply.InstanceName = inst.Name
			deply.InstanceRoles = inst.Roles
			deply.InstanceMemory = inst.Memory
			deply.InstanceProcessors = inst.Processors
			deply.InstanceStatus = inst.Status
			deply.InstanceUptime = inst.Uptime
			deply.InstanceState = inst.State
			deply.InstanceAction = inst.Action

			if inst.Guest != nil {
				deply.InstanceGuestStatus = inst.Guest.Status
				deply.InstanceTimestamp = inst.Guest.Timestamp
				deply.InstanceHeartbeat = inst.Guest.Heartbeat
				if inst.IsActive() {
					deply.InstanceMemoryUsage = inst.Guest.Memory
					deply.InstanceHugePages = inst.Guest.HugePages
					deply.InstanceLoad1 = inst.Guest.Load1
					deply.InstanceLoad5 = inst.Guest.Load5
					deply.InstanceLoad15 = inst.Guest.Load15
				} else {
					deply.InstanceGuestStatus = types.Offline
				}
			}
		}

		if len(doc.ImageDocs) > 0 {
			img := doc.ImageDocs[0]

			deply.ImageId = img.Id
			deply.ImageName = img.Name
		}

		deplys = append(deplys, deply)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
