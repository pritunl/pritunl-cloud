package ahandlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/completion"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/plan"
	"github.com/pritunl/pritunl-cloud/pod"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/pritunl/pritunl-cloud/unit"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
)

func completionGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	if demo.IsDemo() {
		data := &completion.Completion{}

		for _, item := range demo.Organizations {
			data.Organizations = append(data.Organizations, &database.Named{
				Id:   item.Id,
				Name: item.Name,
			})
		}

		for _, item := range demo.Authorities {
			data.Authorities = append(data.Authorities, &database.Named{
				Id:   item.Id,
				Name: item.Name,
			})
		}

		for _, item := range demo.Policies {
			data.Policies = append(data.Policies, &database.Named{
				Id:   item.Id,
				Name: item.Name,
			})
		}

		for _, item := range demo.Domains {
			data.Domains = append(data.Domains, &domain.Completion{
				Id:           item.Id,
				Name:         item.Name,
				Organization: item.Organization,
			})
		}

		for _, item := range demo.Vpcs {
			data.Vpcs = append(data.Vpcs, &vpc.Completion{
				Id:           item.Id,
				Name:         item.Name,
				Organization: item.Organization,
				VpcId:        item.VpcId,
				Network:      item.Network,
				Subnets:      item.Subnets,
				Datacenter:   item.Datacenter,
			})
		}

		for _, item := range demo.Datacenters {
			data.Datacenters = append(data.Datacenters, &datacenter.Completion{
				Id:          item.Id,
				Name:        item.Name,
				NetworkMode: item.NetworkMode,
			})
		}

		for _, item := range demo.Blocks {
			data.Blocks = append(data.Blocks, &block.Completion{
				Id:   item.Id,
				Name: item.Name,
				Type: item.Type,
			})
		}

		for _, item := range demo.Nodes {
			data.Nodes = append(data.Nodes, &node.Completion{
				Id:    item.Id,
				Name:  item.Name,
				Zone:  item.Zone,
				Types: item.Types,
			})
		}

		for _, item := range demo.Pools {
			data.Pools = append(data.Pools, &pool.Completion{
				Id:   item.Id,
				Name: item.Name,
				Zone: item.Zone,
			})
		}

		for _, item := range demo.Zones {
			data.Zones = append(data.Zones, &zone.Completion{
				Id:         item.Id,
				Datacenter: item.Datacenter,
				Name:       item.Name,
			})
		}

		for _, item := range demo.Shapes {
			data.Shapes = append(data.Shapes, &shape.Completion{
				Id:         item.Id,
				Name:       item.Name,
				Datacenter: item.Datacenter,
				Flexible:   item.Flexible,
				Memory:     item.Memory,
				Processors: item.Processors,
			})
		}

		imgs, err := image.GetAllCompletion(db, &bson.M{})
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
		data.Images = imgs

		for _, item := range demo.Storages {
			data.Storages = append(data.Storages, &storage.Completion{
				Id:   item.Id,
				Name: item.Name,
				Type: item.Type,
			})
		}

		for _, item := range demo.Instances {
			data.Instances = append(data.Instances, &instance.Completion{
				Id:           item.Id,
				Name:         item.Name,
				Organization: item.Organization,
				Zone:         item.Zone,
				Vpc:          item.Vpc,
				Subnet:       item.Subnet,
				Node:         item.Node,
			})
		}

		for _, item := range demo.Plans {
			data.Plans = append(data.Plans, &plan.Completion{
				Id:           item.Id,
				Name:         item.Name,
				Organization: item.Organization,
			})
		}

		for _, item := range demo.Certificates {
			data.Certificates = append(
				data.Certificates,
				&certificate.Completion{
					Id:           item.Id,
					Name:         item.Name,
					Organization: item.Organization,
					Type:         item.Type,
				},
			)
		}

		for _, item := range demo.Secrets {
			data.Secrets = append(data.Secrets, &secret.Completion{
				Id:           item.Id,
				Name:         item.Name,
				Organization: item.Organization,
				Type:         item.Type,
			})
		}

		for _, item := range demo.Pods {
			data.Pods = append(data.Pods, &pod.Completion{
				Id:           item.Id,
				Name:         item.Name,
				Organization: item.Organization,
			})
		}

		for _, item := range demo.Units {
			data.Units = append(data.Units, &unit.Completion{
				Id:           item.Id,
				Pod:          item.Pod,
				Organization: item.Organization,
				Name:         item.Name,
				Kind:         item.Kind,
			})
		}

		c.JSON(200, data)
		return
	}

	cmpl, err := completion.GetCompletion(db, primitive.NilObjectID, nil)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, cmpl)
}
