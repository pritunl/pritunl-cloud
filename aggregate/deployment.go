package aggregate

import (
	"time"

	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
)

type DeploymentPipe struct {
	Deployment   `bson:",inline"`
	InstanceDocs []*instance.Instance `bson:"instance_docs"`
	NodeDocs     []*node.Node         `bson:"node_docs"`
}

type Deployment struct {
	Id                  primitive.ObjectID     `bson:"_id" json:"id"`
	Service             primitive.ObjectID     `bson:"service" json:"service"`
	Unit                primitive.ObjectID     `bson:"unit" json:"unit"`
	Spec                primitive.ObjectID     `bson:"spec" json:"spec"`
	Kind                string                 `bson:"kind" json:"kind"`
	State               string                 `bson:"state" json:"state"`
	Status              string                 `bson:"status" json:"status"`
	Node                primitive.ObjectID     `bson:"node" json:"node"`
	Instance            primitive.ObjectID     `bson:"instance" json:"instance"`
	InstanceData        *deployment.Deployment `bson:"instance_data" json:"instance_data"`
	NodeName            string                 `bson:"-" json:"node_name"`
	InstanceName        string                 `bson:"-" json:"instance_name"`
	InstanceRoles       []string               `bson:"-" json:"instance_roles"`
	InstanceMemory      int                    `bson:"-" json:"instance_memory"`
	InstanceProcessors  int                    `bson:"-" json:"instance_processors"`
	InstanceStatus      string                 `bson:"-" json:"instance_status"`
	InstanceUptime      string                 `bson:"-" json:"instance_uptime"`
	InstanceState       string                 `bson:"-" json:"instance_state"`
	InstanceVirtState   string                 `bson:"-" json:"instance_virt_state"`
	InstanceHeartbeat   time.Time              `bson:"-" json:"instance_heartbeat"`
	InstanceMemoryUsage float64                `bson:"-" json:"instance_memory_usage"`
	InstanceHugePages   float64                `bson:"-" json:"instance_hugepages"`
	InstanceLoad1       float64                `bson:"-" json:"instance_load1"`
	InstanceLoad5       float64                `bson:"-" json:"instance_load5"`
	InstanceLoad15      float64                `bson:"-" json:"instance_load15"`
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
				"from":         "nodes",
				"localField":   "node",
				"foreignField": "_id",
				"as":           "node_docs",
			},
		},
		&bson.M{
			"$project": &bson.D{
				{"_id", 1},
				{"service", 1},
				{"unit", 1},
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
				{"node_docs.name", 1},
			},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		doc := &DeploymentPipe{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		deply := &doc.Deployment

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
			deply.InstanceState = inst.VirtState

			if inst.Guest != nil {
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

		deplys = append(deplys, deply)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
