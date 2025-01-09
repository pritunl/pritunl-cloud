package aggregate

import (
	"time"

	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/zone"
)

type DeploymentPipe struct {
	Deployment   `bson:",inline"`
	InstanceDocs []*instance.Instance `bson:"instance_docs"`
	ZoneDocs     []*zone.Zone         `bson:"zone_docs"`
	NodeDocs     []*node.Node         `bson:"node_docs"`
	ImageDocs    []*image.Image       `bson:"image_docs"`
}

type Deployment struct {
	Id                  primitive.ObjectID       `bson:"_id" json:"id"`
	Pod                 primitive.ObjectID       `bson:"pod" json:"pod"`
	Unit                primitive.ObjectID       `bson:"unit" json:"unit"`
	Spec                primitive.ObjectID       `bson:"spec" json:"spec"`
	Timestamp           time.Time                `bson:"timestamp" json:"timestamp"`
	Tags                []string                 `bson:"tags" json:"tags"`
	Kind                string                   `bson:"kind" json:"kind"`
	State               string                   `bson:"state" json:"state"`
	Status              string                   `bson:"status" json:"status"`
	Node                primitive.ObjectID       `bson:"node" json:"node"`
	Instance            primitive.ObjectID       `bson:"instance" json:"instance"`
	InstanceData        *deployment.InstanceData `bson:"instance_data" json:"instance_data"`
	ImageId             primitive.ObjectID       `bson:"image_id" json:"image_id"`
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
	InstanceVirtState   string                   `bson:"-" json:"instance_virt_state"`
	InstanceGuestStatus string                   `bson:"-" json:"instance_guest_status"`
	InstanceHeartbeat   time.Time                `bson:"-" json:"instance_heartbeat"`
	InstanceMemoryUsage float64                  `bson:"-" json:"instance_memory_usage"`
	InstanceHugePages   float64                  `bson:"-" json:"instance_hugepages"`
	InstanceLoad1       float64                  `bson:"-" json:"instance_load1"`
	InstanceLoad5       float64                  `bson:"-" json:"instance_load5"`
	InstanceLoad15      float64                  `bson:"-" json:"instance_load15"`
}

func GetDeployments(db *database.Database, unitId primitive.ObjectID) (
	deplys []*Deployment, err error) {

	coll := db.Deployments()
	deplys = []*Deployment{}

	cursor, err := coll.Aggregate(db, []*bson.M{
		&bson.M{
			"$match": &bson.M{
				"unit": unitId,
			},
		},
		&bson.M{
			"$sort": &bson.M{
				"_id": 1,
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
				{"tags", 1},
				{"spec", 1},
				{"kind", 1},
				{"state", 1},
				{"status", 1},
				{"node", 1},
				{"instance", 1},
				{"instance_data", 1},
				{"instance_docs.name", 1},
				{"instance_docs.network_roles", 1},
				{"instance_docs.memory", 1},
				{"instance_docs.processors", 1},
				{"instance_docs.state", 1},
				{"instance_docs.virt_state", 1},
				{"instance_docs.virt_timestamp", 1},
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
			deply.Tags = append([]string{"latest"}, deply.Tags...)
		}

		if len(doc.ZoneDocs) > 0 {
			zne := doc.ZoneDocs[0]
			deply.ZoneName = zne.Name
		}

		if len(doc.NodeDocs) > 0 {
			nde := doc.NodeDocs[0]
			deply.NodeName = nde.Name
		}

		if len(doc.InstanceDocs) > 0 {
			inst := doc.InstanceDocs[0]
			inst.Json(true)

			deply.InstanceName = inst.Name
			deply.InstanceRoles = inst.NetworkRoles
			deply.InstanceMemory = inst.Memory
			deply.InstanceProcessors = inst.Processors
			deply.InstanceStatus = inst.Status
			deply.InstanceUptime = inst.Uptime
			deply.InstanceState = inst.State
			deply.InstanceVirtState = inst.VirtState

			if inst.Guest != nil {
				deply.InstanceGuestStatus = inst.Guest.Status
				deply.InstanceHeartbeat = inst.Guest.Heartbeat
				if inst.IsActive() {
					deply.InstanceMemoryUsage = inst.Guest.Memory
					deply.InstanceHugePages = inst.Guest.HugePages
					deply.InstanceLoad1 = inst.Guest.Load1
					deply.InstanceLoad5 = inst.Guest.Load5
					deply.InstanceLoad15 = inst.Guest.Load15
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
