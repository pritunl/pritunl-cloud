package images

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/utils"
)

// buildQuery converts the filter to the database query, this mirrors the
// query built by the admin images handler, the name matches the image
// name or key.
func buildQuery(filter resource.Filter) bson.M {
	query := bson.M{}

	imageId, ok := utils.ParseObjectId(filter["id"])
	if ok {
		query["_id"] = imageId
	}

	name := strings.TrimSpace(filter["name"])
	if name != "" {
		query["$or"] = []*bson.M{
			&bson.M{
				"name": &bson.M{
					"$regex": fmt.Sprintf(".*%s.*",
						regexp.QuoteMeta(name)),
					"$options": "i",
				},
			},
			&bson.M{
				"key": &bson.M{
					"$regex": fmt.Sprintf(".*%s.*",
						regexp.QuoteMeta(name)),
					"$options": "i",
				},
			},
		}
	}

	typ := strings.TrimSpace(filter["type"])
	if typ != "" {
		query["type"] = typ
	}

	organization, ok := utils.ParseObjectId(filter["organization"])
	if ok {
		query["organization"] = organization
	}

	return query
}
