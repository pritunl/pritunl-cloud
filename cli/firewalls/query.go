package firewalls

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/utils"
)

// buildQuery converts the filter to the database query, this mirrors the
// query built by the admin firewalls handler.
func buildQuery(filter resource.Filter) bson.M {
	query := bson.M{}

	fireId, ok := utils.ParseObjectId(filter["id"])
	if ok {
		query["_id"] = fireId
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
