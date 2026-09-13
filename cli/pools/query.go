package pools

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/utils"
)

// buildQuery converts the filter to the database query, this mirrors the
// query built by the admin pools handler with the volume group filter
// of the web interface that the handler does not read.
func buildQuery(filter resource.Filter) bson.M {
	query := bson.M{}

	poolId, ok := utils.ParseObjectId(filter["id"])
	if ok {
		query["_id"] = poolId
	}

	name := strings.TrimSpace(filter["name"])
	if name != "" {
		query["name"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", regexp.QuoteMeta(name)),
			"$options": "i",
		}
	}

	comment := strings.TrimSpace(filter["comment"])
	if comment != "" {
		query["comment"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", regexp.QuoteMeta(comment)),
			"$options": "i",
		}
	}

	vgName := strings.TrimSpace(filter["vg_name"])
	if vgName != "" {
		query["vg_name"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", regexp.QuoteMeta(vgName)),
			"$options": "i",
		}
	}

	return query
}
