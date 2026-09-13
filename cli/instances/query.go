package instances

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/utils"
)

// buildQuery converts the filter to the database query, this mirrors the
// query built by the admin instances handler.
func buildQuery(filter resource.Filter) bson.M {
	query := bson.M{}

	instId, ok := utils.ParseObjectId(filter["id"])
	if ok {
		query["_id"] = instId
	}

	name := strings.TrimSpace(filter["name"])
	if name != "" {
		query["name"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", regexp.QuoteMeta(name)),
			"$options": "i",
		}
	}

	role := strings.TrimSpace(filter["role"])
	if role != "" {
		if strings.HasPrefix(role, "~") {
			role := role[1:]
			if strings.HasPrefix(role, "!") {
				query["roles"] = &bson.M{
					"$not": &bson.M{
						"$regex": fmt.Sprintf(".*%s.*",
							regexp.QuoteMeta(role[1:])),
						"$options": "i",
					},
				}
			} else {
				query["$or"] = []*bson.M{
					&bson.M{
						"roles": &bson.M{
							"$regex": fmt.Sprintf(".*%s.*",
								regexp.QuoteMeta(role)),
							"$options": "i",
						},
					},
				}
			}
		} else {
			if strings.HasPrefix(role, "!") {
				role = strings.TrimLeft(role, "!")
				query["roles"] = &bson.M{
					"$ne": role,
				}
			} else {
				query["roles"] = role
			}
		}
	}

	networkNamespace := strings.TrimSpace(filter["network_namespace"])
	if networkNamespace != "" {
		query["network_namespace"] = networkNamespace
	}

	address := strings.TrimSpace(filter["address"])
	if address != "" {
		address = strings.SplitN(address, "/", 2)[0]
		addressQuery := &bson.M{
			"$regex": fmt.Sprintf("^%s(/\\d+)?$",
				regexp.QuoteMeta(address)),
			"$options": "i",
		}
		query["$and"] = []*bson.M{
			&bson.M{
				"$or": []*bson.M{
					&bson.M{"public_ips": addressQuery},
					&bson.M{"public_ips6": addressQuery},
					&bson.M{"private_ips": addressQuery},
					&bson.M{"private_ips6": addressQuery},
					&bson.M{"gateway_ips": addressQuery},
					&bson.M{"gateway_ips6": addressQuery},
					&bson.M{"cloud_private_ips": addressQuery},
					&bson.M{"cloud_public_ips": addressQuery},
					&bson.M{"cloud_public_ips6": addressQuery},
					&bson.M{"host_ips": addressQuery},
					&bson.M{"node_port_ips": addressQuery},
				},
			},
		}
	}

	nodeId, ok := utils.ParseObjectId(filter["node"])
	if ok {
		query["node"] = nodeId
	}

	zoneId, ok := utils.ParseObjectId(filter["zone"])
	if ok {
		query["zone"] = zoneId
	}

	vpcId, ok := utils.ParseObjectId(filter["vpc"])
	if ok {
		query["vpc"] = vpcId
	}

	subnetId, ok := utils.ParseObjectId(filter["subnet"])
	if ok {
		query["subnet"] = subnetId
	}

	organization, ok := utils.ParseObjectId(filter["organization"])
	if ok {
		query["organization"] = organization
	}

	comment := strings.TrimSpace(filter["comment"])
	if comment != "" {
		query["comment"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", regexp.QuoteMeta(comment)),
			"$options": "i",
		}
	}

	return query
}
