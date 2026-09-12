package balancers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/utils"
)

// buildQuery converts the filter to the database query, this mirrors the
// query built by the admin balancers handler.
func buildQuery(filter resource.Filter) bson.M {
	query := bson.M{}

	balancerId, ok := utils.ParseObjectId(filter["id"])
	if ok {
		query["_id"] = balancerId
	}

	name := strings.TrimSpace(filter["name"])
	if name != "" {
		query["name"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", regexp.QuoteMeta(name)),
			"$options": "i",
		}
	}

	organization, ok := utils.ParseObjectId(filter["organization"])
	if ok {
		query["organization"] = organization
	}

	datacenter, ok := utils.ParseObjectId(filter["datacenter"])
	if ok {
		query["datacenter"] = datacenter
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
