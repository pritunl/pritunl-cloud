package nodes

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
)

// buildQuery converts the filter to the database query, this mirrors the
// query built by the admin nodes handler.
func buildQuery(filter resource.Filter) bson.M {
	query := bson.M{}

	nodeId, ok := utils.ParseObjectId(filter["id"])
	if ok {
		query["_id"] = nodeId
	}

	name := strings.TrimSpace(filter["name"])
	if name != "" {
		query["name"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", regexp.QuoteMeta(name)),
			"$options": "i",
		}
	}

	zone, ok := utils.ParseObjectId(filter["zone"])
	if ok {
		query["zone"] = zone
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

	// Type filters are yes, no or any like the web checkboxes
	types := []string{}
	notTypes := []string{}
	for _, typ := range []string{node.Admin, node.User, node.Hypervisor} {
		switch filter[typ] {
		case "true":
			types = append(types, typ)
		case "false":
			notTypes = append(notTypes, typ)
		}
	}

	typesQuery := bson.M{}
	if len(types) > 0 {
		typesQuery["$all"] = types
	}
	if len(notTypes) > 0 {
		typesQuery["$nin"] = notTypes
	}
	if len(types) > 0 || len(notTypes) > 0 {
		query["types"] = &typesQuery
	}

	return query
}
