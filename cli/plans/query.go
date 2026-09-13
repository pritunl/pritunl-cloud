package plans

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/utils"
)

// buildQuery converts the filter to the database query, this mirrors the
// query built by the admin plans handler.
func buildQuery(filter resource.Filter) bson.M {
	query := bson.M{}

	planId, ok := utils.ParseObjectId(filter["id"])
	if ok {
		query["_id"] = planId
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

	return query
}
